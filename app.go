package main

import (
	"context"
	"fmt"
	"sync"

	"phant/internal/collector"
	"phant/internal/dump"
	"phant/internal/setup"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const DumpEventSchemaVersion = dump.SchemaVersion
const DumpEventRuntimeChannel = "phant:dump:event"

var ErrUnsupportedSchemaVersion = dump.ErrUnsupportedSchemaVersion

// App struct
type App struct {
	app             *application.App
	collector       *collector.Server
	collectorStatus CollectorStatus
	collectorSubID  int
	collectorDone   chan struct{}
	collectorWG     sync.WaitGroup
}

type CollectorStatus struct {
	Running    bool   `json:"running"`
	SocketPath string `json:"socketPath"`
	LastError  string `json:"lastError"`
	Dropped    uint64 `json:"dropped"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) setApplication(app *application.App) {
	a.app = app
}

func (a *App) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	return a.startupCollector()
}

func (a *App) startupCollector() error {

	socketPath := collector.DefaultSocketPath()
	server := collector.NewServer(socketPath, collector.DefaultBufferSize)

	a.collectorStatus = CollectorStatus{
		Running:    false,
		SocketPath: socketPath,
	}

	if err := server.Start(); err != nil {
		a.collectorStatus.LastError = err.Error()
		return err
	}

	a.collector = server
	a.collectorStatus.Running = true
	a.startCollectorEventBridge()

	return nil
}

func (a *App) ServiceShutdown() error {
	a.shutdownCollector()
	return nil
}

func (a *App) shutdownCollector() {
	if a.collector == nil {
		return
	}

	a.stopCollectorEventBridge()

	if err := a.collector.Stop(); err != nil {
		a.collectorStatus.LastError = err.Error()
	}

	a.collectorStatus.Dropped = a.collector.DroppedCount()
	a.collectorStatus.Running = false
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SupportedDumpEventSchemaVersion() int {
	return dump.SchemaVersion
}

func (a *App) DecodeDumpEventNDJSONLine(line string) (*dump.Event, error) {
	return dump.DecodeNDJSONLine(line)
}

func (a *App) GetCollectorStatus() CollectorStatus {
	if a.collector != nil {
		a.collectorStatus.Dropped = a.collector.DroppedCount()
	}

	return a.collectorStatus
}

func (a *App) GetRecentEvents(limit int) []dump.Event {
	if a.collector == nil {
		return []dump.Event{}
	}

	events := a.collector.Events()
	if limit <= 0 || limit >= len(events) {
		return events
	}

	return events[len(events)-limit:]
}

func (a *App) DumpEventChannelName() string {
	return DumpEventRuntimeChannel
}

func (a *App) GetSetupDiagnostics() setup.Diagnostics {
	return setup.CollectDiagnostics(context.Background())
}

func (a *App) EnableCLIHook() setup.HookInstallResult {
	socketPath := collector.DefaultSocketPath()
	if a.collectorStatus.SocketPath != "" {
		socketPath = a.collectorStatus.SocketPath
	}

	return setup.InstallCLIHook(context.Background(), socketPath)
}

func (a *App) GetValetLinuxVerification() setup.ValetLinuxVerification {
	return setup.VerifyValetLinux(context.Background())
}

func (a *App) ApplyValetLinuxRemediation(confirm bool) setup.ValetLinuxRemediationResult {
	return setup.ApplyValetLinuxRemediation(context.Background(), confirm)
}

func (a *App) startCollectorEventBridge() {
	if a.collector == nil {
		return
	}

	if a.collectorDone != nil {
		return
	}

	subID, ch := a.collector.Subscribe(256)
	a.collectorSubID = subID
	a.collectorDone = make(chan struct{})
	a.collectorWG.Add(1)

	go func() {
		defer a.collectorWG.Done()

		for {
			select {
			case <-a.collectorDone:
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				if a.app != nil {
					a.app.Event.Emit(DumpEventRuntimeChannel, event)
				}
			}
		}
	}()
}

func (a *App) stopCollectorEventBridge() {
	if a.collector == nil || a.collectorDone == nil {
		return
	}

	close(a.collectorDone)
	a.collector.Unsubscribe(a.collectorSubID)
	a.collectorWG.Wait()
	a.collectorDone = nil
	a.collectorSubID = 0
}
