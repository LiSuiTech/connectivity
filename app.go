package main

import (
	"connectivity/control"
	"connectivity/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	run "runtime"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	// 定义您的应用程序字段，例如:
	TcpServer     *control.FuncTcpServer
	TcpClient     *control.FuncTcpClient
	TcpServerConn *control.TcpServerConn
	UdpServer     *control.FuncUdpServer
	UdpClient     *control.FuncUdpClient
	UdpServerConn *control.UdpServerConn
	Message       *control.Message
	Db            *sql.DB
	ctx           context.Context
}

func NewApp() *App {
	return &App{
		TcpServer: &control.FuncTcpServer{
			Servers: make(map[int]control.NetListener),
			Conn:    make(map[string]control.ServerConn),
			// 其他初始化...
		},
		TcpClient: &control.FuncTcpClient{
			Connections:     make(map[int]net.Conn),
			ClientReadTasks: make(map[int]*control.ClientReadTask),
			ScheduledTasks:  make(map[int]*control.ScheduledTask),
		},
		TcpServerConn: &control.TcpServerConn{},
		UdpServer: &control.FuncUdpServer{
			Servers: make(map[int]control.NetListenerUdp),
			Conn:    make(map[string]control.ServerConnUdp),
			Wg:      sync.WaitGroup{},
		},
		UdpClient: &control.FuncUdpClient{
			Connections:     make(map[int]net.Conn),
			ClientReadTasks: make(map[int]*control.ClientReadUdpTask),
			ScheduledTasks:  make(map[int]*control.ScheduledUdpTask),
		},
		UdpServerConn: &control.UdpServerConn{},
		Message:       &control.Message{},
	}
}

func (app *App) startup(ctx context.Context) {
	// 在应用启动时执行的逻辑
	log.Println("应用启动")
	// 例如，您可以在这里加载服务器配置
	// 获取数据库文件路径
	dbPath := getAppDataPath()

	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// 如果数据库文件不存在，创建并初始化
		createDatabase(dbPath)
	}

	// 开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 加载服务器配置
	if err := app.loadServerConfigs(db); err != nil {
		log.Fatal(err)
	}

	// 设置数据库
	if err := app.SetDB(); err != nil {
		log.Fatal(err)
	}

	app.ctx = ctx
	app.TcpClient.Ctx = app.ctx
	app.TcpServer.Ctx = app.ctx
	app.Message.Ctx = app.ctx
	app.TcpServerConn.Ctx = app.ctx
	app.UdpClient.Ctx = app.ctx
	app.UdpServer.Ctx = app.ctx
	app.UdpServerConn.Ctx = app.ctx
}

func (app *App) SetDB() error {
	if app.Db == nil {
		return errors.New("数据库未初始化")
	}
	app.TcpClient.Db = app.Db
	app.TcpServer.Db = app.Db
	app.TcpServerConn.Db = app.Db
	app.Message.Db = app.Db
	app.UdpClient.Db = app.Db
	app.UdpServer.Db = app.Db
	app.UdpServerConn.Db = app.Db
	return nil
}

func (app *App) loadServerConfigs(db *sql.DB) error {
	// 实现加载服务器配置的逻辑
	// 例如，从数据库中读取配置并初始化 TcpServer
	if db == nil {
		return errors.New("数据库未初始化")
	}
	app.Db = db
	return nil
}

func (a *App) ShowWarningDialog(title string, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   title,
		Message: message,
	})
}

func getAppDataPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Fatal("Unable to determine home directory")
	}
	var dbPath string
	if run.GOOS == "darwin" { // macOS
		dbPath = filepath.Join(homeDir, "Documents", "SocketTools", "data.db")
	} else if run.GOOS == "windows" { // Windows
		dbPath = filepath.Join(os.Getenv("APPDATA"), "SocketTools", "data.db")
	} else { // Linux
		dbPath = filepath.Join(homeDir, ".local", "share", "SocketTools", "data.db")
	}
	return dbPath
}

// 创建数据库并初始化表格
func createDatabase(dbPath string) {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal("Failed to create directory:", err)
	}

	// 打开数据库文件（如果不存在则会创建）
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 初始化数据库表
	if err := models.InitDB(db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database created and initialized successfully.")
}
