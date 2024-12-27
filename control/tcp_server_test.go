package control

import (
	"connectivity/types"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func initSqlite() *sql.DB {
	db, err := sql.Open("sqlite3", "../data.db")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return db
}

func TestAddTCPServer(t *testing.T) {
	server := types.Server{
		Remark: "测试",
		Host:   "127.0.0.1",
		Port:   8088,
	}

	serverStr := &FuncTcpServer{
		Db: initSqlite(),
	}

	defer serverStr.Db.Close()
	fmt.Println(serverStr.AddTCPServer(server))
}

func TestGetAllTCPServers(t *testing.T) {
	serverStr := &FuncTcpServer{
		Db: initSqlite(),
	}
	defer serverStr.Db.Close()

	fmt.Println(serverStr.GetAllTCPServers())
}

func TestGetTCPServerData(t *testing.T) {

	serverStr := &FuncTcpServer{
		Db: initSqlite(),
	}
	defer serverStr.Db.Close()

	fmt.Println(serverStr.GetTCPServerData(4))
}

func TestStartTCPServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	serverStr := &FuncTcpServer{
		Db:      initSqlite(),
		Servers: make(map[int]NetListener),
		Conn:    make(map[string]ServerConn),
		Wg:      sync.WaitGroup{},
		Ctx:     ctx,
	}
	defer serverStr.Db.Close()

	fmt.Println(serverStr.StartTCPServer(4))
	cancel()
}

func TestUpdateTCPServer(t *testing.T) {
	serverStr := &FuncTcpServer{
		Db: initSqlite(),
	}
	defer serverStr.Db.Close()

	fmt.Println(serverStr.UpdateTCPServer(types.Server{ID: 3, Host: "127.0.0.2", Port: 8088}))
}
