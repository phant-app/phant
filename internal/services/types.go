package services

import "phant/internal/dump"

const DumpEventSchemaVersion = dump.SchemaVersion
const DumpEventRuntimeChannel = "phant:dump:event"

var ErrUnsupportedSchemaVersion = dump.ErrUnsupportedSchemaVersion

type CollectorStatus struct {
	Running    bool   `json:"running"`
	SocketPath string `json:"socketPath"`
	LastError  string `json:"lastError"`
	Dropped    uint64 `json:"dropped"`
}
