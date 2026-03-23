package main

import (
	"embed"
	"phant/internal/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	appServices := services.NewAppServices()

	app := application.New(application.Options{
		Name: "Phant",
		Services: []application.Service{
			application.NewService(appServices.Lifecycle),
			application.NewService(appServices.Dump),
			application.NewService(appServices.Setup),
			application.NewService(appServices.PHP),
			application.NewService(appServices.Services),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	services.AttachApplication(appServices.Lifecycle, app)

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Phant",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGBA(27, 38, 54, 255),
	})

	err := app.Run()

	if err != nil {
		panic(err)
	}
}
