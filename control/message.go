package control

import (
	"connectivity/models"
	"connectivity/types"
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type Message struct {
	Db  *sql.DB
	mu  sync.Mutex
	Ctx context.Context
}

func (m *Message) AddMessage(clientId int, content string, inputMethod string, displayMethod string, encoding string, direction string) types.ConnectResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	resp := types.ConnectResult{}
	resp.Success = true
	resp.Message = "消息添加成功"

	if err := models.AddMessage(m.Db, clientId, content, inputMethod, displayMethod, encoding, direction); err != nil {
		resp.Success = false
		resp.Message = fmt.Sprintf("消息添加失败: %v", err)
	}

	return resp
}
func (m *Message) GetServerAllMessages(serverID int64, connID int64) types.ConnectResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	messages, err := models.GetServerAllMessages(m.Db, int(serverID), int(connID))
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取消息失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "获取消息成功",
		Data:    messages,
	}
}

func (m *Message) GetAllMessages(clientID int) types.ConnectResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	messages, err := models.GetAllMessages(m.Db, clientID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取消息失败: %v", err),
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "获取消息成功",
		Data:    messages,
	}
}
func (m *Message) DeleteMessage(id int) types.ConnectResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := models.DeleteMessage(m.Db, id); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("删除消息失败: %v", err),
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "删除消息成功",
	}
}

func (m *Message) DeleteMessageByServerID(serverID int, connID int) types.ConnectResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := models.DeleteMessageByServerID(m.Db, serverID, connID); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("删除消息失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "删除消息成功",
	}
}
