package models

import (
	"connectivity/types"
	"database/sql"
)

// 添加 TCP 服务器
func AddServer(db *sql.DB, server types.Server) error {
	_, err := db.Exec(`INSERT INTO server (remark, host, port, status, type) VALUES (?, ?, ?, ?, ?)`,
		server.Remark, server.Host, server.Port, server.Status, server.Type)
	return err
}

func GetAllServers(db *sql.DB, typer string) ([]types.Server, error) {
	rows, err := db.Query(`SELECT id, remark, host, port, status, type FROM server WHERE type = '` + typer + `' order by id`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var servers []types.Server
	for rows.Next() {
		var server types.Server
		if err := rows.Scan(&server.ID, &server.Remark, &server.Host, &server.Port, &server.Status, &server.Type); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, nil
}

// 更新 TCP 服务器
func UpdateServer(db *sql.DB, server types.Server) error {
	_, err := db.Exec(`UPDATE server SET remark=?, host=?, port=?, status=?, type=? WHERE id=?`,
		server.Remark, server.Host, server.Port, server.Status, server.Type, server.ID)
	return err
}

// 删除 TCP 服务器
func DeleteServer(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM server WHERE id=?`, id)
	return err
}

func FindServerOne(db *sql.DB, id int) (types.Server, error) {
	var server types.Server
	err := db.QueryRow(`SELECT id, remark, host, port, status, type FROM server WHERE id=?`, id).Scan(&server.ID, &server.Remark, &server.Host, &server.Port, &server.Status, &server.Type)
	return server, err
}
