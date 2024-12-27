package main

import (
	"embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "网络调试工具",
		Width:  1280,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Menu:             createMenu(app),
		Bind: []interface{}{
			app,
			app.TcpClient,
			app.TcpServer,
			app.Message,
			app.TcpServerConn,
			app.UdpClient,
			app.UdpServer,
			app.UdpServerConn,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
