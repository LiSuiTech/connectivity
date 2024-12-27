package models

import (
	"connectivity/types"
	"database/sql"
)

func InsertServerConn(db *sql.DB, serverID int, connStatus string, connHost string, connPort int) error {
	stmt, err := db.Prepare("INSERT INTO server_conn (server_id, conn_status, conn_host, conn_port, conn_create_time) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(serverID, connStatus, connHost, connPort)
	if err != nil {
		return err
	}
	return nil
}

func UpdateServerConn(db *sql.DB, serverID int, id int, connStatus string) error {
	stmt, err := db.Prepare("UPDATE server_conn SET conn_status = ?, conn_update_time = CURRENT_TIMESTAMP WHERE server_id = ? AND conn_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(connStatus, serverID, id)
	if err != nil {
		return err
	}
	return nil
}

func UpdateServerConnStatus(db *sql.DB, serverID, id int, connStatus string) error {
	stmt, err := db.Prepare("UPDATE server_conn SET conn_status = ? WHERE server_id = ? AND conn_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(connStatus, serverID, id)
	if err != nil {
		return err
	}
	return nil
}

func DeleteServerConn(db *sql.DB, serverID int, id int) error {
	stmt, err := db.Prepare("DELETE FROM server_conn WHERE server_id = ? AND conn_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(serverID, id)
	if err != nil {
		return err
	}
	return nil
}

func DeleteServerConnByServerID(db *sql.DB, serverID int) error {
	stmt, err := db.Prepare("DELETE FROM server_conn WHERE server_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(serverID)
	if err != nil {
		return err
	}
	return nil
}

func GetServerConn(db *sql.DB, serverID int) ([]*types.ServerConn, error) {
	var conns []*types.ServerConn
	stmt, err := db.Prepare("SELECT conn_id, server_id, conn_status, conn_host, conn_port, conn_create_time, conn_update_time FROM server_conn WHERE server_id = ? ORDER BY conn_id DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		conn := &types.ServerConn{}
		err = rows.Scan(&conn.ID, &conn.ServerID, &conn.ConnStatus, &conn.ConnHost, &conn.ConnPort, &conn.ConnCreateTime, &conn.ConnUpdateTime)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}

	return conns, nil
}

func FindServerConnOne(db *sql.DB, serverID int, connPort int) (*types.ServerConn, error) {
	stmt, err := db.Prepare("SELECT conn_id, server_id, conn_status, conn_host, conn_port, conn_create_time, conn_update_time FROM server_conn WHERE server_id = ? AND conn_port = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(serverID, connPort)
	conn := &types.ServerConn{}
	err = row.Scan(&conn.ID, &conn.ServerID, &conn.ConnStatus, &conn.ConnHost, &conn.ConnPort, &conn.ConnCreateTime, &conn.ConnUpdateTime)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
