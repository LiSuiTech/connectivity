package control

import (
	"connectivity/models"
	"connectivity/types"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FuncTcpServer struct {
	mu      sync.Mutex
	Servers map[int]NetListener
	Conn    map[string]ServerConn
	Ctx     context.Context
	Db      *sql.DB
	Wg      sync.WaitGroup
}

type NetListener struct {
	ID       int
	Ctx      context.Context
	Cancel   context.CancelFunc
	Listener net.Listener
}

type ServerConn struct {
	Conn net.Conn
}

func (a *FuncTcpServer) AddTCPServer(config types.Server) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查是否有相同的 Host 和 Port 的服务器
	servers, _ := models.GetAllServers(a.Db, "tcp")
	for _, server := range servers {
		if server.Host == config.Host && server.Port == config.Port {
			return types.ConnectResult{
				Success: false,
				Message: "服务器已存在",
			}
		}
	}

	config.Status = "stopped"

	if err := models.AddServer(a.Db, config); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("添加服务器失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "添加服务器成功",
	}
}

func (a *FuncTcpServer) GetAllTCPServers() types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	servers, err := models.GetAllServers(a.Db, "tcp")
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器失败: %v", err),
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "获取服务器成功",
		Data:    servers,
	}
}

func (a *FuncTcpServer) UpdateTCPServer(config types.Server) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	server, err := models.FindServerOne(a.Db, config.ID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器失败: %v", err),
		}
	} else {
		if server.Host != config.Host || server.Port != config.Port {
			// 检查是否有相同的 Host 和 Port 的服务器
			servers, _ := models.GetAllServers(a.Db, "tcp")
			for _, serverConn := range servers {
				if serverConn.Host == config.Host && serverConn.Port == config.Port {
					return types.ConnectResult{
						Success: false,
						Message: "服务器已存在",
					}
				}
			}
			if a.Servers[server.ID].Listener != nil {
				a.StopTCPServer(server.ID)
			}
		}

		server.Remark = config.Remark
		server.Host = config.Host
		server.Port = config.Port
		server.Status = "stopped"
	}

	if err := models.UpdateServer(a.Db, server); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("更新服务器失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "更新服务器成功",
	}
}

func (a *FuncTcpServer) DeleteTCPServer(id int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if conn, exists := a.Servers[id]; exists {
		delete(a.Servers, id)
		conn.Listener.Close()
		conn.Cancel()
	}

	// 删除所有与该服务器相关的连接
	models.DeleteServerConnByServerID(a.Db, id)

	models.DeleteMessageByServerID(a.Db, id, 0)

	if err := models.DeleteServer(a.Db, id); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("删除服务器失败: %v", err),
		}
	}
	return types.ConnectResult{
		Success: true,
		Message: "删除服务器成功",
	}
}

// StartTCPServer 启动 TCP 服务器
func (a *FuncTcpServer) StartTCPServer(id int) types.ConnectResult {
	config, err := models.FindServerOne(a.Db, id)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器失败: %v", err),
		}
	}

	if config.Status == "running" && a.Servers[config.ID].Listener != nil {
		return types.ConnectResult{
			Success: false,
			Message: "服务器已运行",
		}
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			return types.ConnectResult{
				Success: false,
				Message: "服务器已运行",
			}
		}
		if strings.Contains(err.Error(), "permission denied") {
			return types.ConnectResult{
				Success: false,
				Message: "权限不足",
			}
		}
		if strings.Contains(err.Error(), "address family not supported") {
			return types.ConnectResult{
				Success: false,
				Message: "地址族不支持",
			}
		}
		// bind: can't assign requested address
		if strings.Contains(err.Error(), "bind: can't assign requested address") {
			return types.ConnectResult{
				Success: false,
				Message: "无法分配请求的地址",
			}
		}
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("启动失败: %v", err),
		}
	}

	a.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.Servers[config.ID] = NetListener{
		ID:       config.ID,
		Ctx:      ctx,
		Cancel:   cancel,
		Listener: listener,
	}
	a.mu.Unlock()

	config.Status = "running"
	if err := models.UpdateServer(a.Db, config); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("更新服务器失败: %v", err),
		}
	}
	a.Wg.Add(1)
	// 优化：使用独立的函数处理连接，以提高代码可读性和可维护性
	go a.acceptConnections(ctx, config, listener)

	return types.ConnectResult{
		Success: true,
		Message: "服务器启动成功",
	}
}

// 新增一个独立的函数来处理接受连接的逻辑
func (a *FuncTcpServer) acceptConnections(ctx context.Context, config types.Server, listener net.Listener) {
	defer a.Wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if !isClosedError(err) {
					runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
						Type:     "error",
						ServerId: config.ID,
						Message: &types.Message{
							ID:            config.ID,
							Content:       fmt.Sprintf("接受连接错误: %v", err),
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
							Direction:     "incoming",
							InputMethod:   "tcp",
							DisplayMethod: "text",
							Encoding:      "utf-8",
						},
					})
				}
				return
			}

			a.mu.Lock()
			// 获取连接的key,客户端的 IP，端口
			connKey := fmt.Sprintf("%d:%d", config.ID, conn.RemoteAddr().(*net.TCPAddr).Port)
			fmt.Println(connKey)
			a.Conn[connKey] = ServerConn{
				Conn: conn,
			}

			if err := models.InsertServerConn(a.Db, config.ID, "connected", conn.RemoteAddr().(*net.TCPAddr).IP.String(), conn.RemoteAddr().(*net.TCPAddr).Port); err != nil {
				conn.Close()
				continue
			}

			a.mu.Unlock()
			a.Wg.Add(1)
			go a.handleTCPConnection(ctx, config.ID, conn)
			runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
				Type:     "connection_status",
				ServerId: config.ID,
				Message: &types.Message{
					Content: "连接已建立",
				},
			})
		}
	}
}

func (a *FuncTcpServer) GetTCPServerStatus(serverID int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	client, err := models.FindServerOne(a.Db, serverID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取客户端数据失败: %v", err),
			Data:    client,
		}
	}
	if _, exists := a.Servers[serverID]; exists {
		return types.ConnectResult{
			Success: true,
			Message: "连接在线",
			Data:    client,
		}
	} else {
		client.Status = "stopped"
		err = models.UpdateServer(a.Db, client)
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

func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, net.ErrClosed)
}

// StopServer 停止指定类型的服务器
func (a *FuncTcpServer) StopTCPServer(serverID int) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	fmt.Println(serverID, a.Servers)
	server, exists := a.Servers[serverID]
	if !exists {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("服务器未运行: %d", serverID),
		}
	}
	fmt.Println(server)
	server.Cancel()

	// 关闭监听器
	if err := server.Listener.Close(); err != nil && !isClosedError(err) {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("停止服务器失败: %v", err),
		}
	}

	// 取消并移除所有与该服务器相关的连接
	for key, conn := range a.Conn {
		// 假设连接的 key 包含 serverID，可以根据实际情况调整判断条件
		if strings.HasPrefix(key, fmt.Sprintf("%d:", serverID)) {
			conn.Conn.Close()
			delete(a.Conn, key)
		}
	}

	// 等待所有相关的 goroutine 完成
	a.Wg.Wait()

	// 从服务器列表中删除
	delete(a.Servers, serverID)

	serverData, err := models.FindServerOne(a.Db, serverID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器失败: %v", err),
		}
	} else {
		serverData.Status = "stopped"
		// 更新数据库中的服务器状态
		if err := models.UpdateServer(a.Db, serverData); err != nil {
			return types.ConnectResult{
				Success: false,
				Message: fmt.Sprintf("更新服务器状态失败: %v", err),
			}
		}
	}

	// 发射服务器停止事件
	runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
		Type:     "server_stopped",
		ServerId: serverID,
		Message: &types.Message{
			ID:            serverID,
			Content:       "服务器已停止",
			Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
			Direction:     "system",
			InputMethod:   "system",
			DisplayMethod: "text",
			Encoding:      "utf-8",
		},
	})

	return types.ConnectResult{
		Success: true,
		Message: "停止服务器成功",
	}
}

// handleTCPConnection 处理 TCP 连接
func (a *FuncTcpServer) handleTCPConnection(ctx context.Context, serverID int, conn net.Conn) {
	defer func() {
		conn.Close()
		a.Wg.Done()
		// 从连接映射中删除连接
		a.mu.Lock()
		connKey := fmt.Sprintf("%d:%d", serverID, conn.RemoteAddr().(*net.TCPAddr).Port)
		delete(a.Conn, connKey)
		a.mu.Unlock()

		// 更新数据库中的连接状态
		if err := models.UpdateServerConnStatus(a.Db, serverID, conn.RemoteAddr().(*net.TCPAddr).Port, "disconnected"); err != nil {
			runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
				Type:     "error",
				ServerId: serverID,
				Message: &types.Message{
					ID:            serverID,
					Content:       fmt.Sprintf("更新连接状态失败: %v", err),
					Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					Direction:     "system",
					InputMethod:   "tcp",
					DisplayMethod: "text",
					Encoding:      "utf-8",
				},
			})
		}
	}()

	buffer := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					// 客户端主动断开连接
					runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
						Type:     "connection_closed",
						ServerId: serverID,
						Message: &types.Message{
							ID:            serverID,
							Content:       "客户端已断开连接",
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
							Direction:     "system",
							InputMethod:   "tcp",
							DisplayMethod: "text",
							Encoding:      "utf-8",
						},
					})
				} else {
					// 其他读取错误
					runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
						Type:     "error",
						ServerId: serverID,
						Message: &types.Message{
							ID:            serverID,
							Content:       fmt.Sprintf("读取数据错误: %v", err),
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
							Direction:     "incoming",
							InputMethod:   "tcp",
							DisplayMethod: "text",
							Encoding:      "utf-8",
						},
					})
				}
				return
			}

			if n > 0 {
				data := buffer[:n]
				connID := fmt.Sprintf("%d:%d", serverID, conn.RemoteAddr().(*net.TCPAddr).Port)
				models.AddMessageServer(a.Db, serverID, connID, string(data), "tcp", "text", "utf-8", "incoming")
				runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
					Type:     "data_received",
					ServerId: serverID,
					Message: &types.Message{
						ServerID:      int64(serverID),
						ConnID:        strconv.Itoa(conn.RemoteAddr().(*net.TCPAddr).Port),
						Content:       string(data),
						Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
						Direction:     "incoming",
						InputMethod:   "tcp",
						DisplayMethod: "text",
						Encoding:      "utf-8",
					},
				})
				buffer = make([]byte, 1024)
			}
		}
	}
}

func (a *FuncTcpServer) GetTCPServerData(serverID int) types.ConnectResult {

	data, err := models.FindServerOne(a.Db, serverID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器数据失败: %v", err),
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "获取服务器数据成功",
		Data:    data,
	}
}

// 修改 SendMessage 函数，发送消息到现有连接而非接受新连接
func (a *FuncTcpServer) SendMessage(serverID int, port int, message string) types.ConnectResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := models.FindServerOne(a.Db, serverID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器失败: %v", err),
		}
	}

	connID := fmt.Sprintf("%d:%d", serverID, port)

	conn, exists := a.Conn[connID]
	if !exists {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("连接不存在: %s", connID),
		}
	} else {
		// 遍历所有连接并发送消息
		_, err := conn.Conn.Write([]byte(message))
		if err != nil {
			runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
				Type:     "error",
				ServerId: serverID,
				Message: &types.Message{
					ServerID:      int64(serverID),
					ConnID:        strconv.Itoa(port),
					Content:       fmt.Sprintf("发送消息错误: %v", err),
					Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					Direction:     "outgoing",
					InputMethod:   "tcp",
					DisplayMethod: "text",
					Encoding:      "utf-8",
				},
			})
		} else {
			connID := fmt.Sprintf("%d:%d", serverID, port)
			models.AddMessageServer(a.Db, serverID, connID, message, "tcp", "text", "utf-8", "incoming")
			runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
				Type:     "data_sent",
				ServerId: serverID,
				Message: &types.Message{
					ServerID:      int64(serverID),
					ConnID:        strconv.Itoa(port),
					Content:       message,
					Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					Direction:     "outgoing",
					InputMethod:   "tcp",
					DisplayMethod: "text",
					Encoding:      "utf-8",
				},
			})
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "发送消息成功",
	}
}

// 检查连接状态
func (a *FuncTcpServer) CheckConnectionStatus() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 创建 serverID 到地址的映射
	serverIDMap := make(map[string]int)
	for id, listener := range a.Servers {
		serverIDMap[listener.Listener.Addr().String()] = id
	}

	for key, serverConn := range a.Conn {
		// 设置读取超时
		serverConn.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buffer := make([]byte, 1) // 创建一个小缓冲区

		_, err := serverConn.Conn.Read(buffer)
		if err != nil {
			// 如果读取失败，连接可能已关闭
			serverID, exists := serverIDMap[serverConn.Conn.RemoteAddr().String()]
			if exists {
				runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
					Type:     "connection_status",
					ServerId: serverID,
					Message: &types.Message{
						ID:            serverID,
						Content:       "连接已关闭",
						Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
						Direction:     "system",
						InputMethod:   "tcp",
						DisplayMethod: "text",
						Encoding:      "utf-8",
					},
				})
				delete(a.Conn, key) // 从连接列表中删除无效连接
			}
		}
	}
}

func (a *FuncTcpServer) DisconnectClient(serverID int, port int) types.ConnectResult {

	_, err := models.FindServerConnOne(a.Db, serverID, port)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取连接失败: %v", err),
		}
	}

	return types.ConnectResult{
		Success: true,
		Message: "断开连接成功",
	}
}
