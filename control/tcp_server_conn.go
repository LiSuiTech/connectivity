package control

import (
	"connectivity/models"
	"connectivity/types"
	"context"
	"database/sql"
)

type TcpServerConn struct {
	Ctx context.Context
	Db  *sql.DB
}

func (t *TcpServerConn) GetTCPServerConn(serverID int, port int) *types.ConnectResult {
	resp := &types.ConnectResult{}
	conn, err := models.FindServerConnOne(t.Db, serverID, port)
	if err != nil {
		resp.Success = false
		resp.Message = err.Error()
		return resp
	}
	resp.Success = true
	resp.Data = conn
	return resp
}

func (t *TcpServerConn) GetAllTCPServerConn(id int) *types.ConnectResult {
	resp := &types.ConnectResult{}
	conns, err := models.GetServerConn(t.Db, id)
	if err != nil {
		resp.Success = false
		resp.Message = err.Error()
		return resp
	}
	data := []*types.ServerConn{}
	data = append(data, conns...)
	resp.Success = true
	resp.Data = data
	return resp
}
