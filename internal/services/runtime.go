package services

import (
	"sync"

	"phant/internal/collector"
	"phant/internal/dump"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type collectorRuntime struct {
	app             *application.App
	socketPath      string
	collector       *collector.Server
	collectorStatus CollectorStatus
	collectorSubID  int
	collectorDone   chan struct{}
	collectorWG     sync.WaitGroup
}

func (r *collectorRuntime) collectorSocketPath() string {
	if r.collectorStatus.SocketPath != "" {
		return r.collectorStatus.SocketPath
	}

	if r.socketPath != "" {
		return r.socketPath
	}

	return collector.DefaultSocketPath()
}

func (r *collectorRuntime) getCollectorStatus() CollectorStatus {
	if r.collector != nil {
		r.collectorStatus.Dropped = r.collector.DroppedCount()
	}

	return r.collectorStatus
}

func (r *collectorRuntime) getRecentEvents(limit int) []dump.Event {
	if r.collector == nil {
		return []dump.Event{}
	}

	events := r.collector.Events()
	if limit <= 0 || limit >= len(events) {
		return events
	}

	return events[len(events)-limit:]
}
