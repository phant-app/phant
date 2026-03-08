package main

import (
	"phant/internal/services"
	"strings"
	"testing"
)

func TestDecodeDumpEventNDJSONLine_EmptyLineIsIgnored(t *testing.T) {
	dumpService := services.NewAppServices().Dump
	event, err := dumpService.DecodeDumpEventNDJSONLine("   \n")
	if err != nil {
		t.Fatalf("expected no error for empty line, got %v", err)
	}
	if event != nil {
		t.Fatalf("expected nil event for empty line")
	}
}

func TestDecodeDumpEventNDJSONLine_ValidHTTPEvent(t *testing.T) {
	dumpService := services.NewAppServices().Dump
	line := `{"schemaVersion":1,"id":"01JNFKEC8Q4Y8S97R2M5W12Q9H","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"http","projectRoot":"/home/ronald/code/example-app","phpSapi":"fpm-fcgi","requestId":"f2a1a3d2-2087-4dc4-9fc4-3f8e75ae3202","http":{"method":"GET","scheme":"https","host":"example.test","path":"/users/42"},"isDd":false,"payloadFormat":"json","payload":{"user":{"id":42}},"trace":[],"host":{"hostname":"ronald-linux","pid":48211}}`

	event, err := dumpService.DecodeDumpEventNDJSONLine(line)
	if err != nil {
		t.Fatalf("expected valid event, got error %v", err)
	}
	if event == nil {
		t.Fatalf("expected non-nil event")
	}
}

func TestDecodeDumpEventNDJSONLine_ValidCLIEventWithNullRequestID(t *testing.T) {
	dumpService := services.NewAppServices().Dump
	line := `{"schemaVersion":1,"id":"01JNFKEPA3A4CNV3K2E12YVYTG","timestamp":"2026-02-28T11:21:18.011Z","sourceType":"cli","projectRoot":"/home/ronald/code/example-app","phpSapi":"cli","requestId":null,"command":{"name":"artisan","args":["queue:work"]},"isDd":false,"payloadFormat":"json","payload":{"ok":true},"trace":[],"host":{"hostname":"ronald-linux","pid":49302}}`

	event, err := dumpService.DecodeDumpEventNDJSONLine(line)
	if err != nil {
		t.Fatalf("expected valid event, got error %v", err)
	}
	if event == nil {
		t.Fatalf("expected non-nil event")
	}
}

func TestDecodeDumpEventNDJSONLine_InvalidCases(t *testing.T) {
	dumpService := services.NewAppServices().Dump
	tests := []struct {
		name    string
		line    string
		wantErr string
	}{
		{
			name:    "missing required key",
			line:    `{"schemaVersion":1,"timestamp":"2026-02-28T11:20:31.331Z","sourceType":"http","projectRoot":"/x","phpSapi":"fpm-fcgi","requestId":"a","http":{"method":"GET","scheme":"https","host":"example.test","path":"/"},"isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "missing required dump event field: id",
		},
		{
			name:    "unsupported schema version",
			line:    `{"schemaVersion":2,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"http","projectRoot":"/x","phpSapi":"fpm-fcgi","requestId":"a","http":{"method":"GET","scheme":"https","host":"example.test","path":"/"},"isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: services.ErrUnsupportedSchemaVersion.Error(),
		},
		{
			name:    "invalid sourceType",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"job","projectRoot":"/x","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "sourceType must be one of: http, cli, worker, cron",
		},
		{
			name:    "requestId wrong type",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"cli","projectRoot":"/x","phpSapi":"cli","requestId":123,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "requestId must be null or string",
		},
		{
			name:    "isDd wrong type",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"cli","projectRoot":"/x","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":"no","payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "isDd must be a boolean",
		},
		{
			name:    "trace wrong type",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"cli","projectRoot":"/x","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":{},"host":{"hostname":"h","pid":1}}`,
			wantErr: "trace must be an array",
		},
		{
			name:    "payloadFormat not json",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"cli","projectRoot":"/x","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"text","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "payloadFormat must be json for schemaVersion 1",
		},
		{
			name:    "http source missing http meta",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"http","projectRoot":"/x","phpSapi":"fpm-fcgi","requestId":"a","isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "http metadata is required when sourceType is http",
		},
		{
			name:    "cli source missing command meta",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"cli","projectRoot":"/x","phpSapi":"cli","requestId":null,"isDd":false,"payloadFormat":"json","payload":{"k":"v"},"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "command metadata is required when sourceType is cli, worker, or cron",
		},
		{
			name:    "invalid payload json",
			line:    `{"schemaVersion":1,"id":"1","timestamp":"2026-02-28T11:20:31.331Z","sourceType":"cli","projectRoot":"/x","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"json","payload":,"trace":[],"host":{"hostname":"h","pid":1}}`,
			wantErr: "invalid character",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := dumpService.DecodeDumpEventNDJSONLine(test.line)
			if err == nil {
				t.Fatalf("expected error, got nil and event %#v", event)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected error containing %q, got %q", test.wantErr, err.Error())
			}
		})
	}
}
