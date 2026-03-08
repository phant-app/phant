package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	service := NewApp()

	app := application.New(application.Options{
		Name: "phant",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	service.setApplication(app)

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "phant",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGBA(27, 38, 54, 255),
	})

	err := app.Run()

	if err != nil {
		panic(err)
	}
}
