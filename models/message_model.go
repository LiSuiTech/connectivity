package models

import (
	"connectivity/types"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// 添加消息
func AddMessage(db *sql.DB, clientID int, content string, inputMethod string, displayMethod string, encoding string, direction string) error {
	_, err := db.Exec(`INSERT INTO message (client_id, content, input_method, display_method, encoding, direction, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`, clientID, content, inputMethod, displayMethod, encoding, direction, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

func AddMessageServer(db *sql.DB, serverID int, connID string, content string, inputMethod string, displayMethod string, encoding string, direction string) error {
	_, err := db.Exec(`INSERT INTO message (server_id, conn_id, content, input_method, display_method, encoding, direction, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, serverID, connID, content, inputMethod, displayMethod, encoding, direction, time.Now().Format("2006-01-02 15:04:05"))

	return err
}

// 获取所有消息
func GetAllMessages(db *sql.DB, clientID int) ([]types.Message, error) {
	rows, err := db.Query(`SELECT id, client_id, content, input_method, display_method, encoding, direction, timestamp FROM message WHERE client_id=?`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var message types.Message
		if err := rows.Scan(&message.ID, &message.ClientID, &message.Content, &message.InputMethod, &message.DisplayMethod, &message.Encoding, &message.Direction, &message.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func GetServerAllMessages(db *sql.DB, serverID int, connID int) ([]*types.Message, error) {
	rows, err := db.Query(`SELECT id, server_id, conn_id, content, input_method, display_method, encoding, direction, timestamp FROM message WHERE conn_id='` + fmt.Sprintf("%d:%d", serverID, connID) + `' limit 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*types.Message
	for rows.Next() {
		message := &types.Message{}
		if err := rows.Scan(&message.ID, &message.ServerID, &message.ConnID, &message.Content, &message.InputMethod, &message.DisplayMethod, &message.Encoding, &message.Direction, &message.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func DeleteMessageByServerID(db *sql.DB, serverID int, connID int) error {
	if connID == 0 {
		_, err := db.Exec(`DELETE FROM message WHERE server_id=` + strconv.Itoa(serverID))
		return err
	}
	_, err := db.Exec(`DELETE FROM message WHERE server_id=` + strconv.Itoa(serverID) + ` AND conn_id='` + fmt.Sprintf("%d:%d", serverID, connID) + `'`)
	return err
}

// 删除消息
func DeleteMessage(db *sql.DB, clientID int) error {
	_, err := db.Exec(`DELETE FROM message WHERE client_id=?`, clientID)
	return err
}
