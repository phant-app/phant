package services

import "phant/internal/dump"

type DumpService struct {
	runtime *collectorRuntime
}

func (s *DumpService) SupportedDumpEventSchemaVersion() int {
	return dump.SchemaVersion
}

func (s *DumpService) DecodeDumpEventNDJSONLine(line string) (*dump.Event, error) {
	return dump.DecodeNDJSONLine(line)
}

func (s *DumpService) GetCollectorStatus() CollectorStatus {
	return s.runtime.getCollectorStatus()
}

func (s *DumpService) GetRecentEvents(limit int) []dump.Event {
	return s.runtime.getRecentEvents(limit)
}

func (s *DumpService) DumpEventChannelName() string {
	return DumpEventRuntimeChannel
}
