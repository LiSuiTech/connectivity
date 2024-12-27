package control

import (
	"bytes"
	"connectivity/models"
	"connectivity/types"
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FuncUdpClient struct {
	mu              sync.Mutex
	Connections     map[int]net.Conn
	ScheduledTasks  map[int]*ScheduledUdpTask
	ClientReadTasks map[int]*ClientReadUdpTask
	Db              *sql.DB
	Ctx             context.Context
}

// SendScheduledMessage 定时发送消息到 Udp 连接
// 存储定时任务的map
type ScheduledUdpTask struct {
	ticker *time.Ticker
	done   chan bool
}

// ClientReadTask 客户端读取任务
type ClientReadUdpTask struct {
	conn net.Conn
	done chan bool
}

// AddUdpClient 添加 Udp 客户端
func (a *FuncUdpClient) AddUdpClient(config types.ServerClient) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, conn := range a.Connections {
		if conn.LocalAddr().String() == fmt.Sprintf("%s:%d", config.Host, config.Port) {
			return types.ConnectResult{
				Success: false,
				Message: "连接已存在",
			}
		}
	}
	if err := models.AddServerClient(a.Db, config); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("添加客户端失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "添加客户端成功",
	}
}

// UpdateUdpClient 更新 Udp 客户端
func (a *FuncUdpClient) UpdateUdpClient(config types.ServerClient) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.Connections[config.ID]; exists {
		return types.ConnectResult{
			Success: false,
			Message: "连接在线无法修改，请断线后再修改",
		}
	}
	if err := models.UpdateServerClient(a.Db, config); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("更新客户端失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "更新客户端成功",
	}
}

// GetAllUdpClients 获取所有 Udp 客户端
func (a *FuncUdpClient) GetAllUdpClients() types.ConnectResult {
	result, err := models.GetAllServerClients(a.Db, "udp")
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取客户端失败: %v", err),
		}
	}
	for _, client := range result {
		if _, exists := a.Connections[client.ID]; exists {
			client.Status = "online"
		} else {
			client.Status = "offline"
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "获取客户端成功",
		Data:    result,
	}
}

// DeleteUdpClient 删除 Udp 客户端
func (a *FuncUdpClient) DeleteUdpClient(id int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	if conn, exists := a.Connections[id]; exists {
		conn.Close()
		delete(a.Connections, id)
	}

	if err := models.DeleteServerClient(a.Db, id); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("删除客户端失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "删除客户端成功",
	}
}

// handleUdpConnection 处理 Udp 客户端连接
func (a *FuncUdpClient) handleUdpConnectionTask(clientID int) {

	task, exists := a.ClientReadTasks[clientID]
	if !exists {
		return
	}

	defer task.conn.Close()

	for {
		select {
		case <-task.done:
			return
		default:
			// 创建一个固定大小的缓冲区
			buffer := make([]byte, 1024)
			n, err := task.conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					runtime.LogError(a.Ctx, fmt.Sprintf("连接已断开: %v", err))
					return
				}
				continue
			}

			// 只取实际读取的数据
			data := string(buffer[:n])

			if err := models.AddMessage(a.Db, clientID, data, "Udp", "text", "utf-8", "incoming"); err != nil {
				runtime.LogError(a.Ctx, fmt.Sprintf("添加消息失败: %v", err))
			}
			runtime.EventsEmit(a.Ctx, "client_event", types.ServerEvent{
				Type:     "data_received",
				ServerId: clientID,
				Message: &types.Message{
					ID:            clientID,
					Content:       data,
					Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					Direction:     "incoming",
					InputMethod:   "Udp",
					DisplayMethod: "text",
					Encoding:      "utf-8",
				},
			})
		}
	}
}

func (a *FuncUdpClient) handleUdpConnection(clientID int, conn net.Conn) {
	defer conn.Close()

	for {
		// 创建一个固定大小的缓冲区
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				runtime.LogError(a.Ctx, fmt.Sprintf("连接已断开: %v", err))
				return
			}
			continue
		}

		// 只取实际读取的数据
		data := string(buffer[:n])
		if err := models.AddMessage(a.Db, clientID, data, "Udp", "text", "utf-8", "incoming"); err != nil {
			runtime.LogError(a.Ctx, fmt.Sprintf("添加消息失败: %v", err))
		}
		runtime.EventsEmit(a.Ctx, "client_event", types.ServerEvent{
			Type:     "data_received",
			ServerId: clientID,
			Message: &types.Message{
				ID:            clientID,
				Content:       data,
				Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
				Direction:     "incoming",
				InputMethod:   "Udp",
				DisplayMethod: "text",
				Encoding:      "utf-8",
			},
		})
	}
}

// ConnectUdpClient 连接 Udp 客户端
func (a *FuncUdpClient) ConnectUdpClient(clientID int) types.ConnectResult {

	client, err := models.GetServerClientData(a.Db, clientID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取客户端数据失败: %v", err),
		}
	}

	client.Status = "online"
	err = models.UpdateServerClient(a.Db, client)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("更新客户端状态失败: %v", err),
		}
	}

	if client.Host == "" {
		return types.ConnectResult{
			Success: false,
			Message: "主机地址不能为空",
		}
	}
	if client.Port == 0 {
		return types.ConnectResult{
			Success: false,
			Message: "端口号不能为空",
		}
	}

	ip := net.ParseIP(client.Host)
	if ip == nil {
		return types.ConnectResult{
			Success: false,
			Message: "主机地址不是有效的 IP 地址",
		}
	}

	addr := fmt.Sprintf("%s:%d", client.Host, client.Port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: %v", err),
		}
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: %v", err),
		}
	}

	// 发送一个空的数据包以探测连接
	_, err = conn.Write([]byte{})
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("连接探测失败: %v", err),
		}
	}

	a.mu.Lock()
	a.Connections[client.ID] = conn
	a.mu.Unlock()

	if client.RepeatSend {
		// 创建读取任务
		a.ClientReadTasks[client.ID] = &ClientReadUdpTask{
			conn: conn,
			done: make(chan bool, 1),
		}
		go a.handleUdpConnectionTask(client.ID) // 启动处理连接的 goroutine

	} else {
		go a.handleUdpConnection(client.ID, conn) // 启动处理连接的 goroutine
	}

	return types.ConnectResult{
		Success: true,
		Message: "连接成功",
	}
}

func (a *FuncUdpClient) GetUdpClientStatus(clientID int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	client, err := models.GetServerClientData(a.Db, clientID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取客户端数据失败: %v", err),
			Data:    client,
		}
	}
	if _, exists := a.Connections[clientID]; exists {
		return types.ConnectResult{
			Success: true,
			Message: "连接在线",
			Data:    client,
		}
	} else {
		client.Status = "offline"
		err = models.UpdateServerClient(a.Db, client)
		if err != nil {
			return types.ConnectResult{
				Success: false,
				Message: fmt.Sprintf("更新客户端状态失败: %v", err),
				Data:    client,
			}
		}
		return types.ConnectResult{
			Success: false,
			Message: "连接不在线",
			Data:    client,
		}
	}
}

func (a *FuncUdpClient) GetUdpClientData(clientID int) types.ConnectResult {
	data, err := models.GetServerClientData(a.Db, clientID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取客户端数据失败: %v", err),
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "获取客户端数据成功",
		Data:    data,
	}
}

// DisconnectUdpClient 断开 Udp 客户端连接
func (a *FuncUdpClient) DisconnectUdpClient(clientId int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	client, err := models.GetServerClientData(a.Db, clientId)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取客户端数据失败: %v", err),
		}
	}
	client.Status = "offline"
	err = models.UpdateServerClient(a.Db, client)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("更新客户端状态失败: %v", err),
		}
	}
	if client.RepeatSend {
		task, exists := a.ClientReadTasks[clientId]
		if exists {
			task.done <- true
			delete(a.ClientReadTasks, clientId)
		}
	}

	conn, exists := a.Connections[clientId]
	if !exists {
		return types.ConnectResult{
			Success: false,
			Message: "连接不存在",
		}
	}

	if err := conn.Close(); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("断开连接失败: %v", err),
		}
	}

	delete(a.Connections, clientId)

	return types.ConnectResult{
		Success: true,
		Message: "已断开连接",
	}
}

// SendMessage 发送消息到 Udp 连接
func (a *FuncUdpClient) SendMessage(clientID int, message string, inputMethod string) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	conn, exists := a.Connections[clientID]
	if !exists {
		return types.ConnectResult{
			Success: false,
			Message: "连接不存在",
		}
	}

	buf := bytes.NewBuffer(nil)
	if inputMethod == "hex" {
		buf.WriteString(hex.EncodeToString([]byte(message)))
	} else {
		buf.WriteString(message)
	}

	_, err := conn.Write(buf.Bytes())
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("发送失败: %v", err),
		}
	}

	if err := models.AddMessage(a.Db, clientID, message, inputMethod, "text", "utf-8", "outgoing"); err != nil {
		runtime.LogError(a.Ctx, fmt.Sprintf("添加消息失败: %v", err))
	}

	return types.ConnectResult{
		Success: true,
		Message: "消息发送成功",
	}
}

// SendScheduledMessage 定时发送消息到 Udp 连接
func (a *FuncUdpClient) SendScheduledMessage(clientID int, message string, inputMethod string, interval int) types.ConnectResult {
	a.mu.Lock()
	_, exists := a.Connections[clientID]
	if !exists {
		a.mu.Unlock()
		return types.ConnectResult{
			Success: false,
			Message: "连接不存在",
		}
	}
	a.mu.Unlock()

	// 创建定时器和停止通道
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	done := make(chan bool, 1)

	// 保存定时任务
	a.ScheduledTasks[clientID] = &ScheduledUdpTask{
		ticker: ticker,
		done:   done,
	}

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				a.mu.Lock()
				if conn, ok := a.Connections[clientID]; ok {
					buf := bytes.NewBuffer(nil)
					if inputMethod == "hex" {
						buf.WriteString(hex.EncodeToString([]byte(message)))
					} else {
						buf.WriteString(message)
					}
					_, err := conn.Write(buf.Bytes())
					if err != nil {
						runtime.LogError(a.Ctx, fmt.Sprintf("发送消息失败: %v", err))
						ticker.Stop()
						delete(a.ScheduledTasks, clientID)
					}
					if err := models.AddMessage(a.Db, clientID, message, inputMethod, "text", "utf-8", "outgoing"); err != nil {
						runtime.LogError(a.Ctx, fmt.Sprintf("添加消息失败: %v", err))
					}
					runtime.EventsEmit(a.Ctx, "client_event", types.ServerEvent{
						Type:     "data_sent",
						ServerId: clientID,
						Message: &types.Message{
							ID:            clientID,
							Content:       message,
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
							Direction:     "outgoing",
							InputMethod:   inputMethod,
							DisplayMethod: "text",
							Encoding:      "utf-8",
						},
					})
				} else {
					ticker.Stop()
					delete(a.ScheduledTasks, clientID)
				}
				a.mu.Unlock()
			}
		}
	}()

	return types.ConnectResult{
		Success: true,
		Message: "定时发送任务已启动",
	}
}

// StopScheduledMessage 停止定时发送消息
func (a *FuncUdpClient) StopScheduledMessage(clientID int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if task, exists := a.ScheduledTasks[clientID]; exists {
		task.done <- true
		delete(a.ScheduledTasks, clientID)
		return types.ConnectResult{
			Success: true,
			Message: "定时发送任务已停止",
		}
	}

	return types.ConnectResult{
		Success: false,
		Message: "未找到定时发送任务",
	}
}
