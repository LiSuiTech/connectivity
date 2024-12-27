package control

import (
	"fmt"
	"testing"
)

func TestStartUDPServer(t *testing.T) {
	clientStr := &FuncUdpClient{
		Db: initSqlite(),
	}
	defer clientStr.Db.Close()

	fmt.Println(clientStr.GetAllUdpClients())
}

func TestGetAllUdpServer(t *testing.T) {
	clientStr := &FuncUdpServer{
		Db: initSqlite(),
	}
	defer clientStr.Db.Close()

	fmt.Println(clientStr.GetAllUdpServers())
}
