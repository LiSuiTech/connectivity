package models

import (
	"database/sql"
	"fmt"
)

func InitDB(db *sql.DB) error {
	// 检查并创建 tcp_client 表
	if err := createTableIfNotExists(db, `CREATE TABLE IF NOT EXISTS server_client (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		remark TEXT,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		status TEXT,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		type TEXT CHECK(type IN ('tcp', 'udp')) NOT NULL,
		repeat_send INTEGER DEFAULT 0,
		repeat_interval REAL DEFAULT 1000.0,
		send_content TEXT
	);`); err != nil {
		return err
	}

	// 检查并创建 message 表
	if err := createTableIfNotExists(db, `CREATE TABLE IF NOT EXISTS message (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id INTEGER,
		server_id INTEGER,
		conn_id TEXT,
		content TEXT NOT NULL,
		direction TEXT NOT NULL,
		input_method TEXT NOT NULL,
		display_method TEXT NOT NULL, 
		encoding TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (client_id) REFERENCES tcp_client(id),
		FOREIGN KEY (server_id) REFERENCES tcp_server(id)
	);`); err != nil {
		return err
	}

	// 检查并创建 tcp_server 表
	if err := createTableIfNotExists(db, `CREATE TABLE IF NOT EXISTS server (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		remark TEXT,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		status TEXT,
		type TEXT CHECK(type IN ('tcp', 'udp')) NOT NULL,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		return err
	}

	// 检查并创建 tcp_server_conn 表
	if err := createTableIfNotExists(db, `CREATE TABLE IF NOT EXISTS server_conn (
		conn_id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER,
		conn_status TEXT,
		conn_create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		conn_update_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		conn_host TEXT NOT NULL,
		conn_port INTEGER NOT NULL,
		FOREIGN KEY (server_id) REFERENCES server(id)
	);`); err != nil {
		return err
	}

	return nil
}

// 辅助函数：创建表
func createTableIfNotExists(db *sql.DB, createTableSQL string) error {
	// 获取表名
	var tableName string
	// 解析 SQL 语句以获取表名
	if _, err := fmt.Sscanf(createTableSQL, "CREATE TABLE IF NOT EXISTS %s", &tableName); err != nil {
		return err
	}

	// 检查表是否存在
	var exists int
	checkTableSQL := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?;`
	err := db.QueryRow(checkTableSQL, tableName).Scan(&exists)
	if err != nil {
		return err
	}

	// 如果表不存在，则创建表
	if exists == 0 {
		_, err := db.Exec(createTableSQL)
		return err
	}

	return nil
}
