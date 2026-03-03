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
    $json = json_encode($event, JSON_UNESCAPED_SLASHES | JSON_PARTIAL_OUTPUT_ON_ERROR);
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

function phant_emit_value($var, bool $isDd): void {
    $payload = json_decode(json_encode($var, JSON_PARTIAL_OUTPUT_ON_ERROR), true);
    if ($payload === null && json_last_error() !== JSON_ERROR_NONE) {
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
        'trace' => [],
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
