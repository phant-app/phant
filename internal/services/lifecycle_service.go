package services

import (
	"context"

	"phant/internal/collector"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type CollectorLifecycleService struct {
	runtime *collectorRuntime
}

func AttachApplication(service *CollectorLifecycleService, app *application.App) {
	service.setApplication(app)
}

func (s *CollectorLifecycleService) setApplication(app *application.App) {
	s.runtime.app = app
}

func (s *CollectorLifecycleService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	return s.runtime.startupCollector()
}

func (s *CollectorLifecycleService) ServiceShutdown() error {
	s.runtime.shutdownCollector()
	return nil
}

func (r *collectorRuntime) startupCollector() error {
	socketPath := r.collectorSocketPath()
	server := collector.NewServer(socketPath, collector.DefaultBufferSize)

	r.collectorStatus = CollectorStatus{
		Running:    false,
		SocketPath: socketPath,
	}

	if err := server.Start(); err != nil {
		r.collectorStatus.LastError = err.Error()
		return err
	}

	r.collector = server
	r.collectorStatus.Running = true
	r.startCollectorEventBridge()

	return nil
}

func (r *collectorRuntime) shutdownCollector() {
	if r.collector == nil {
		return
	}

	r.stopCollectorEventBridge()

	if err := r.collector.Stop(); err != nil {
		r.collectorStatus.LastError = err.Error()
	}

	r.collectorStatus.Dropped = r.collector.DroppedCount()
	r.collectorStatus.Running = false
}

func (r *collectorRuntime) startCollectorEventBridge() {
	if r.collector == nil || r.collectorDone != nil {
		return
	}

	subID, ch := r.collector.Subscribe(256)
	r.collectorSubID = subID
	r.collectorDone = make(chan struct{})
	r.collectorWG.Add(1)

	go func() {
		defer r.collectorWG.Done()

		for {
			select {
			case <-r.collectorDone:
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				if r.app != nil {
					r.app.Event.Emit(DumpEventRuntimeChannel, event)
				}
			}
		}
	}()
}

func (r *collectorRuntime) stopCollectorEventBridge() {
	if r.collector == nil || r.collectorDone == nil {
		return
	}

	close(r.collectorDone)
	r.collector.Unsubscribe(r.collectorSubID)
	r.collectorWG.Wait()
	r.collectorDone = nil
	r.collectorSubID = 0
}
