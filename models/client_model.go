package models

import (
	"connectivity/types"
	"database/sql"
)

func AddServerClient(db *sql.DB, client types.ServerClient) error {
	_, err := db.Exec(`INSERT INTO server_client (remark, host, port, status, type, repeat_send, repeat_interval, send_content) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		client.Remark, client.Host, client.Port, client.Status, client.Type, client.RepeatSend, client.RepeatInterval, client.SendContent)
	return err
}

func GetAllServerClients(db *sql.DB, typer string) ([]*types.ServerClient, error) {
	rows, err := db.Query(`SELECT id, remark, host, port, status, type, repeat_send, repeat_interval, send_content FROM server_client WHERE type='` + typer + `'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var clients []*types.ServerClient
	for rows.Next() {
		client := &types.ServerClient{}
		if err := rows.Scan(&client.ID, &client.Remark, &client.Host, &client.Port, &client.Status, &client.Type, &client.RepeatSend, &client.RepeatInterval, &client.SendContent); err != nil {
			return nil, err
		}

		clients = append(clients, client)
	}
	return clients, nil
}

func UpdateServerClient(db *sql.DB, client types.ServerClient) error {
	_, err := db.Exec(`UPDATE server_client SET remark=?, host=?, port=?, status=?, type=?, repeat_send=?, repeat_interval=?, send_content=? WHERE id=?`,
		client.Remark, client.Host, client.Port, client.Status, client.Type, client.RepeatSend, client.RepeatInterval, client.SendContent, client.ID)
	return err
}

func DeleteServerClient(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM server_client WHERE id=?`, id)
	return err
}

func FindServerClientOne(db *sql.DB, id int) (types.ServerClient, error) {
	var client types.ServerClient
	err := db.QueryRow(`SELECT id, remark, host, port, status, type, repeat_send, repeat_interval, send_content FROM server_client WHERE id=?`, id).Scan(&client.ID, &client.Remark, &client.Host, &client.Port, &client.Status, &client.Type, &client.RepeatSend, &client.RepeatInterval, &client.SendContent)
	return client, err
}

func GetServerClientData(db *sql.DB, id int) (types.ServerClient, error) {
	var client types.ServerClient
	err := db.QueryRow(`SELECT id, remark, host, port, status, type, repeat_send, repeat_interval, send_content FROM server_client WHERE id=?`, id).Scan(&client.ID, &client.Remark, &client.Host, &client.Port, &client.Status, &client.Type, &client.RepeatSend, &client.RepeatInterval, &client.SendContent)
	return client, err
}
