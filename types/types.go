package types

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	ID            int    `json:"id"`            // 配置唯一标识
	Remark        string `json:"remark"`        // 备注
	Type          string `json:"type"`          // 连接类型
	Host          string `json:"host"`          // 地址
	Port          int    `json:"port"`          // 端口
	RepeatSend    bool   `json:"repeatSend"`    // 是否重复发送
	RepeatContent string `json:"repeatContent"` // 重复发送内容
}

// ConnectResult 连接结果
type ConnectResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// TCPClient 结构体
type ServerClient struct {
	ID             int     `json:"id"`             // 唯一标识
	Remark         string  `json:"remark"`         // 备注
	Host           string  `json:"host"`           // 地址
	Port           int     `json:"port"`           // 端口
	Status         string  `json:"status"`         // 状态
	Type           string  `json:"type"`           // 连接类型
	RepeatSend     bool    `json:"repeatSend"`     // 是否重复发送
	RepeatInterval float64 `json:"repeatInterval"` // 重复发送间隔
	SendContent    string  `json:"sendContent"`    // 发送内容
}

// Message 结构体
type Message struct {
	ID            int    `json:"id"`             // 唯一标识
	ClientID      int64  `json:"client_id"`      // 客户端唯一标识
	ServerID      int64  `json:"server_id"`      // 服务端唯一标识
	ConnID        string `json:"conn_id"`        // 连接唯一标识
	Content       string `json:"content"`        // 内容
	Direction     string `json:"direction"`      // "outgoing" 或 "incoming"
	InputMethod   string `json:"input_method"`   // 输入方法
	DisplayMethod string `json:"display_method"` // 显示方法
	Encoding      string `json:"encoding"`       // 编码
	Timestamp     string `json:"timestamp"`      // 时间
}

// TCPServer 结构体
type Server struct {
	ID     int    `json:"id"`
	Remark string `json:"remark"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

type ServerEvent struct {
	Type     string   `json:"type"`
	ServerId int      `json:"server_id"`
	Message  *Message `json:"message"`
}

// TCPServerConn 结构体
type ServerConn struct {
	ID             int    `json:"conn_id"`
	ServerID       int    `json:"server_id"`
	ConnStatus     string `json:"conn_status"`
	ConnHost       string `json:"conn_host"`
	ConnPort       int    `json:"conn_port"`
	ConnCreateTime string `json:"conn_create_time"`
	ConnUpdateTime string `json:"conn_update_time"`
}
