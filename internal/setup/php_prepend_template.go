package setup

const phpPrependTemplate = `<?php
if (defined('PHANT_PREPEND_LOADED')) {
    return;
}
define('PHANT_PREPEND_LOADED', true);

$phantSocket = getenv('PHANT_COLLECTOR_SOCKET');
if ($phantSocket === false || $phantSocket === '') {
    $phantSocket = '{{SOCKET_PATH}}';
}

$phantIsDd = false;

function phant_send_event(array $event): void {
    global $phantSocket;

    $client = @stream_socket_client('unix://' . $phantSocket, $errno, $errstr, 0.02);
    if ($client === false) {
        return;
    }

    @stream_set_blocking($client, false);
    $json = json_encode($event, JSON_UNESCAPED_SLASHES | JSON_PARTIAL_OUTPUT_ON_ERROR | JSON_INVALID_UTF8_SUBSTITUTE);
    if ($json !== false) {
        @fwrite($client, $json . "\n");
    }
    @fclose($client);
}

function phant_source_type(): string {
    if (PHP_SAPI !== 'cli') {
        return 'http';
    }

    $argv = $_SERVER['argv'] ?? [];
    $command = implode(' ', $argv);

    if (stripos($command, 'horizon') !== false || stripos($command, 'queue:work') !== false) {
        return 'worker';
    }

    if (stripos($command, 'schedule:run') !== false) {
        return 'cron';
    }

    return 'cli';
}

function phant_command_meta(): ?array {
    if (PHP_SAPI !== 'cli') {
        return null;
    }

    $argv = $_SERVER['argv'] ?? [];
    $name = count($argv) > 0 ? (string)$argv[0] : 'php';

    return [
        'name' => $name,
        'args' => array_values(array_slice($argv, 1)),
        'cwd' => getcwd() ?: '',
    ];
}

function phant_http_meta(): ?array {
    if (PHP_SAPI === 'cli') {
        return null;
    }

    return [
        'method' => $_SERVER['REQUEST_METHOD'] ?? 'GET',
        'scheme' => (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http',
        'host' => $_SERVER['HTTP_HOST'] ?? ($_SERVER['SERVER_NAME'] ?? 'localhost'),
        'path' => $_SERVER['REQUEST_URI'] ?? '/',
    ];
}

function phant_trace_callsite(): array {
    $frames = debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS, 32);

    foreach ($frames as $frame) {
        $file = isset($frame['file']) ? (string)$frame['file'] : '';
        $line = isset($frame['line']) ? (int)$frame['line'] : 0;
        $function = isset($frame['function']) ? (string)$frame['function'] : '';
        $class = isset($frame['class']) ? (string)$frame['class'] : '';

        if ($file === '' || $line <= 0) {
            continue;
        }

        if ($file === __FILE__) {
            continue;
        }

        if (str_contains($file, '/vendor/symfony/var-dumper/')) {
            continue;
        }

        if (str_starts_with($function, 'phant_')) {
            continue;
        }

        if (str_contains($function, 'phant_install_vardumper_handler')) {
            continue;
        }

        if ($function === 'dump' || $function === 'dd') {
            continue;
        }

        if ($class === 'Symfony\\Component\\VarDumper\\VarDumper') {
            continue;
        }

        $func = $function;
        if ($func === '' && $class !== '') {
            $func = $class;
        }
        if ($func === '') {
            $func = '{main}';
        }

        return [[
            'file' => $file,
            'line' => $line,
            'func' => $func,
        ]];
    }

    return [];
}

function phant_normalize_property_name(string $rawKey, string $className): string {
    if (str_starts_with($rawKey, "\0")) {
        $parts = explode("\0", $rawKey, 3);
        if (count($parts) === 3) {
            $declaring = (string)$parts[1];
            $property = (string)$parts[2];

            if ($declaring === '*') {
                return '#' . $property;
            }

            if ($declaring === $className) {
                return '-' . $property;
            }

            return '-' . $declaring . '::' . $property;
        }
    }

    return '+' . str_replace("\0", '', $rawKey);
}

function phant_normalize_value($value, int $depth = 0, array &$seen) {
    if ($depth > 10) {
        return ['__phantType' => 'max-depth'];
    }

    if (is_array($value)) {
        $result = [];
        foreach ($value as $key => $nested) {
            $result[$key] = phant_normalize_value($nested, $depth + 1, $seen);
        }

        return $result;
    }

    if (is_object($value)) {
        $className = get_class($value);
        $objectID = spl_object_id($value);

        if (isset($seen[$objectID])) {
            return [
                '__phantType' => 'object-ref',
                '__className' => $className,
                '__objectId' => $objectID,
            ];
        }

        $seen[$objectID] = true;

        $properties = [];
        foreach ((array)$value as $rawKey => $nested) {
            $propertyName = phant_normalize_property_name((string)$rawKey, $className);
            $properties[$propertyName] = phant_normalize_value($nested, $depth + 1, $seen);
        }

        return [
            '__phantType' => 'object',
            '__className' => $className,
            '__objectId' => $objectID,
            '__properties' => $properties,
        ];
    }

    if (is_resource($value) || (function_exists('is_resource_type') && is_resource_type($value))) {
        return [
            '__phantType' => 'resource',
            '__resourceType' => get_resource_type($value),
        ];
    }

    return $value;
}

function phant_emit_value($var, bool $isDd): void {
    $seen = [];
    $payload = phant_normalize_value($var, 0, $seen);

    $payloadProbe = json_encode($payload, JSON_PARTIAL_OUTPUT_ON_ERROR | JSON_INVALID_UTF8_SUBSTITUTE);
    if ($payloadProbe === false) {
        $payload = ['string' => (string)$var];
    }

    $event = [
        'schemaVersion' => 1,
        'id' => uniqid('phant_', true),
        'timestamp' => gmdate('Y-m-d\\TH:i:s\\Z'),
        'sourceType' => phant_source_type(),
        'projectRoot' => getcwd() ?: '',
        'phpSapi' => PHP_SAPI,
        'requestId' => $_SERVER['HTTP_X_REQUEST_ID'] ?? null,
        'http' => phant_http_meta(),
        'command' => phant_command_meta(),
        'isDd' => $isDd,
        'payloadFormat' => 'json',
        'payload' => $payload,
        'trace' => phant_trace_callsite(),
        'host' => [
            'hostname' => gethostname() ?: 'unknown',
            'pid' => getmypid() ?: 0,
        ],
    ];

    phant_send_event($event);
}

function phant_install_vardumper_handler(): bool {
    static $installed = false;

    if ($installed) {
        return true;
    }

    if (!class_exists('\\Symfony\\Component\\VarDumper\\VarDumper')) {
        return false;
    }

    $previousHandler = null;
    $previousHandler = \Symfony\Component\VarDumper\VarDumper::setHandler(function ($var) use (&$previousHandler): void {
        global $phantIsDd;

        phant_emit_value($var, $phantIsDd === true);

        if (is_callable($previousHandler)) {
            $previousHandler($var);
            return;
        }

        var_dump($var);
    });

    $installed = true;

    return true;
}

phant_install_vardumper_handler();

if (!function_exists('dump')) {
    function dump(...$vars) {
        $handlerInstalled = phant_install_vardumper_handler();

        global $phantIsDd;
        $phantIsDd = false;

        foreach ($vars as $var) {
            if ($handlerInstalled && class_exists('\\Symfony\\Component\\VarDumper\\VarDumper')) {
                \Symfony\Component\VarDumper\VarDumper::dump($var);
                continue;
            }

            phant_emit_value($var, false);
            var_dump($var);
        }

        if (count($vars) === 1) {
            return $vars[0];
        }

        return $vars;
    }
}

if (!function_exists('dd')) {
    function dd(...$vars): void {
        $handlerInstalled = phant_install_vardumper_handler();

        global $phantIsDd;
        $phantIsDd = true;

        foreach ($vars as $var) {
            if ($handlerInstalled && class_exists('\\Symfony\\Component\\VarDumper\\VarDumper')) {
                \Symfony\Component\VarDumper\VarDumper::dump($var);
                continue;
            }

            phant_emit_value($var, true);
            var_dump($var);
        }

        $phantIsDd = false;

        exit(1);
    }
}
`
