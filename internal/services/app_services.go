package services

type Options struct {
	SocketPath string
}

type AppServices struct {
	Lifecycle *CollectorLifecycleService
	Dump      *DumpService
	Setup     *SetupService
}

func NewAppServices() *AppServices {
	return NewAppServicesWithOptions(Options{})
}

func NewAppServicesWithOptions(options Options) *AppServices {
	runtime := &collectorRuntime{
		socketPath: options.SocketPath,
	}

	return &AppServices{
		Lifecycle: &CollectorLifecycleService{runtime: runtime},
		Dump:      &DumpService{runtime: runtime},
		Setup:     &SetupService{runtime: runtime},
	}
}
