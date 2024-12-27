package control

import (
	"fmt"
	"testing"
)

func TestGetServerAllMessages(t *testing.T) {
	serverStr := &Message{
		Db: initSqlite(),
	}
	defer serverStr.Db.Close()

	fmt.Println(serverStr.GetServerAllMessages(3, 63539))
}
