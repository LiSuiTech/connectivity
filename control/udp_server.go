package control

import (
	"connectivity/models"
	"connectivity/types"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FuncUdpServer struct {
	Mu      sync.Mutex
	Servers map[int]NetListenerUdp
	Conn    map[string]ServerConnUdp
	Ctx     context.Context
	Db      *sql.DB
	Wg      sync.WaitGroup
}

type NetListenerUdp struct {
	ID       int
	Ctx      context.Context
	Cancel   context.CancelFunc
	Listener net.PacketConn
}

type ServerConnUdp struct {
	Conn net.PacketConn
}

func (a *FuncUdpServer) AddUdpServer(config types.Server) types.ConnectResult {
	a.Mu.Lock()
	defer a.Mu.Unlock()

	// 检查是否有相同的 Host 和 Port 的服务器
	servers, _ := models.GetAllServers(a.Db, "udp")
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

func (a *FuncUdpServer) GetAllUdpServers() types.ConnectResult {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	servers, err := models.GetAllServers(a.Db, "udp")
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

func (a *FuncUdpServer) UpdateUdpServer(config types.Server) types.ConnectResult {
	a.Mu.Lock()
	defer a.Mu.Unlock()

	server, err := models.FindServerOne(a.Db, config.ID)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取服务器失败: %v", err),
		}
	} else {
		if server.Host != config.Host || server.Port != config.Port {
			// 检查是否有相同的 Host 和 Port 的服务器
			servers, _ := models.GetAllServers(a.Db, "udp")
			for _, serverConn := range servers {
				if serverConn.Host == config.Host && serverConn.Port == config.Port {
					return types.ConnectResult{
						Success: false,
						Message: "服务器已存在",
					}
				}
			}
			if _, ok := a.Servers[server.ID]; ok {
				a.StopUdpServer(server.ID)
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

func (a *FuncUdpServer) DeleteUdpServer(id int) types.ConnectResult {
	a.Mu.Lock()
	defer a.Mu.Unlock()

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

// StopServer 停止指定类型的服务器
func (a *FuncUdpServer) StopUdpServer(serverID int) types.ConnectResult {
	server, exists := a.Servers[serverID]
	if !exists {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("服务器未运行: %d", serverID),
		}
	}

	// 取消服务器的上下文
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

	return types.ConnectResult{
		Success: true,
		Message: "停止服务器成功",
	}
}

// StartTCPServer 启动 TCP 服务器
func (a *FuncUdpServer) StartUdpServer(id int) types.ConnectResult {
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
	listener, err := net.ListenPacket("udp", addr)
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

	a.Mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.Servers[config.ID] = NetListenerUdp{
		ID:       config.ID,
		Ctx:      ctx,
		Cancel:   cancel,
		Listener: listener,
	}
	a.Mu.Unlock()

	config.Status = "running"
	if err := models.UpdateServer(a.Db, config); err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("更新服务器失败: %v", err),
		}
	}
	a.Wg.Add(1)
	// 优化：使用独立的函数处理连接，以提高代码可读性和可维护性
	go a.handleUdpConnection(ctx, config.ID, listener)

	return types.ConnectResult{
		Success: true,
		Message: "服务器启动成功",
	}
}

func (a *FuncUdpServer) handleUdpConnection(ctx context.Context, serverID int, conn net.PacketConn) {
	defer func() {
		conn.Close()
		a.Wg.Done()
		// 从连接映射中删除连接
		a.Mu.Lock()
		connKey := fmt.Sprintf("%d:%d", serverID, conn.LocalAddr().(*net.UDPAddr).Port)
		delete(a.Conn, connKey)
		a.Mu.Unlock()

		// 更新数据库中的连接状态
		if err := models.UpdateServerConnStatus(a.Db, serverID, conn.LocalAddr().(*net.UDPAddr).Port, "disconnected"); err != nil {
			runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
				Type:     "error",
				ServerId: serverID,
				Message: &types.Message{
					ID:            serverID,
					Content:       fmt.Sprintf("更新连接状态失败: %v", err),
					Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					Direction:     "system",
					InputMethod:   "udp",
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
			n, clientAddr, err := conn.ReadFrom(buffer)
			if err != nil {
				if err == io.EOF {
					// 客户端主动断开连接
					models.UpdateServerConn(a.Db, serverID, clientAddr.(*net.UDPAddr).Port, "disconnected")
					runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
						Type:     "connection_closed",
						ServerId: serverID,
						Message: &types.Message{
							ID:            serverID,
							ConnID:        strconv.Itoa(clientAddr.(*net.UDPAddr).Port),
							Content:       "客户端已断开连接",
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
							Direction:     "system",
							InputMethod:   "udp",
							DisplayMethod: "text",
							Encoding:      "utf-8",
						},
					})
				} else {
					runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
						Type:     "error",
						ServerId: serverID,
						Message: &types.Message{
							ID:            serverID,
							Content:       fmt.Sprintf("读取数据错误: %v", err),
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
							Direction:     "incoming",
							InputMethod:   "udp",
							DisplayMethod: "text",
							Encoding:      "utf-8",
						},
					})
				}
				return
			} else {
				if _, ok := a.Conn[fmt.Sprintf("%d:%d", serverID, clientAddr.(*net.UDPAddr).Port)]; !ok {
					// 插入新的连接信息
					a.Mu.Lock()
					a.Conn[fmt.Sprintf("%d:%d", serverID, clientAddr.(*net.UDPAddr).Port)] = ServerConnUdp{
						Conn: conn,
					}
					if _, err := models.FindServerConnOne(a.Db, serverID, clientAddr.(*net.UDPAddr).Port); err != nil {
						// 插入新的连接信息
						models.InsertServerConn(a.Db, serverID, "connected", clientAddr.(*net.UDPAddr).IP.String(), clientAddr.(*net.UDPAddr).Port)
						runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
							Type:     "connection_status",
							ServerId: serverID,
							Message: &types.Message{
								Content: "连接已建立",
							},
						})
					}
					a.Mu.Unlock()
				}
			}

			if n > 0 {
				data := buffer[:n]
				connID := fmt.Sprintf("%d:%d", serverID, clientAddr.(*net.UDPAddr).Port)
				models.AddMessageServer(a.Db, serverID, connID, string(data), "tcp", "text", "utf-8", "incoming")
				runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
					Type:     "data_received",
					ServerId: serverID,
					Message: &types.Message{
						ServerID:      int64(serverID),
						ConnID:        strconv.Itoa(clientAddr.(*net.UDPAddr).Port),
						Content:       string(data),
						Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
						Direction:     "incoming",
						InputMethod:   "udp",
						DisplayMethod: "text",
						Encoding:      "utf-8",
					},
				})
				buffer = make([]byte, 1024)
			}
		}
	}
}

func (a *FuncUdpServer) GetUdpServerStatus(serverID int) types.ConnectResult {
	a.Mu.Lock()
	defer a.Mu.Unlock()
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

func (a *FuncUdpServer) GetUdpServerData(serverID int) types.ConnectResult {

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
func (a *FuncUdpServer) SendMessage(serverID int, port int, message string) types.ConnectResult {
	a.Mu.Lock()
	defer a.Mu.Unlock()

	client, err := models.FindServerConnOne(a.Db, serverID, port)
	if err != nil {
		return types.ConnectResult{
			Success: false,
			Message: fmt.Sprintf("获取连接失败: %v", err),
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
		// 假设目标 IP 和端口是由变量 addr 传递进来
		targetAddr := fmt.Sprintf("%s:%d", client.ConnHost, client.ConnPort) // conn.IP 和 conn.Port 是假设你已经在连接中获得目标信息
		addr, err := net.ResolveUDPAddr("udp", targetAddr)
		if err != nil {
			runtime.EventsEmit(a.Ctx, "server_event", types.ServerEvent{
				Type:     "error",
				ServerId: serverID,
				Message: &types.Message{
					ServerID:      int64(serverID),
					ConnID:        strconv.Itoa(port),
					Content:       fmt.Sprintf("解析目标地址失败: %v", err),
					Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					Direction:     "outgoing",
					InputMethod:   "tcp",
					DisplayMethod: "text",
					Encoding:      "utf-8",
				},
			})
			return types.ConnectResult{
				Success: false,
				Message: fmt.Sprintf("解析目标地址失败: %v", err),
			}
		}

		_, err = conn.Conn.WriteTo([]byte(message), addr)
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
			return types.ConnectResult{
				Success: false,
				Message: fmt.Sprintf("发送消息失败: %v", err),
			}
		} else {
			connID := fmt.Sprintf("%d:%d", serverID, port)
			models.AddMessageServer(a.Db, serverID, connID, message, "tcp", "text", "utf-8", "outgoing")
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

func (a *FuncUdpServer) DisconnectClient(serverID int, port int) types.ConnectResult {

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
