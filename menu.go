package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func createMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()
	// // 设置菜单组
	settingsMenu := appMenu.AddSubmenu("设置")
	// settingsMenu.AddText("通用设置", nil, func(cd *menu.CallbackData) {
	// 	runtime.EventsEmit(app.ctx, "switch-page", "settings")
	// })
	// settingsMenu.AddText("主题设置", nil, func(cd *menu.CallbackData) {
	// 	runtime.EventsEmit(app.ctx, "switch-page", "theme")
	// })

	// 添加分隔线
	settingsMenu.AddSeparator()

	// 添加退出选项
	settingsMenu.AddText("退出", nil, func(cd *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})
	// // TCP 菜单组
	// tcpMenu := appMenu.AddSubmenu("TCP")
	// tcpMenu.AddText("TCP客户端", nil, func(cd *menu.CallbackData) {
	// 	runtime.EventsEmit(app.ctx, "switch-page", "tcp-client")
	// })
	// tcpMenu.AddText("TCP服务端", nil, func(cd *menu.CallbackData) {
	// 	runtime.EventsEmit(app.ctx, "switch-page", "tcp-server")
	// })

	// // UDP 菜单组
	// udpMenu := appMenu.AddSubmenu("UDP")
	// udpMenu.AddText("UDP客户端", nil, func(cd *menu.CallbackData) {
	// 	runtime.EventsEmit(app.ctx, "switch-page", "udp-client")
	// })
	// udpMenu.AddText("UDP服务端", nil, func(cd *menu.CallbackData) {
	// 	runtime.EventsEmit(app.ctx, "switch-page", "udp-server")
	// })

	return appMenu
}
