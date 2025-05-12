package models

import (
	"database/sql"
	"social/config"
)

type PrivateMessage struct {
	Id         int64  `json:"id"`
	SenderId   int64  `json:"senderId"`
	ReceiverId int64  `json:"receiverId"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

type GroupChat struct {
	Id       int64  `json:"id"`
	SenderId int64  `json:"senderId"`
	GroupId  int64  `json:"groupId"`
	Message  string `json:"message"`
	SentAt   string `json:"sentAt"`
}

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository() *ChatRepository {
	return &ChatRepository{db: config.DB}
}

func (r *ChatRepository) CreatePrivateMessage(message *PrivateMessage) error {
	query := `INSERT INTO messages (sender_id, receiver_id, message) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, message.SenderId, message.ReceiverId, message.Message)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	message.Id = id
	return nil
}

func (r *ChatRepository) CreateGroupChat(message *GroupChat) error {
	query := `INSERT INTO group_messages (sender_id, group_id, content) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, message.SenderId, message.GroupId, message.Message)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	message.Id = id
	return nil
}

func (r *ChatRepository) GetPrivateMessages(senderId, receiverId int64, lastMsgId int64) ([]*PrivateMessage, error) {
	var query string
	var rows *sql.Rows
	var err error

	if lastMsgId == 0 {
		query = `SELECT id, sender_id, receiver_id, message, created_at 
		         FROM messages 
		         WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) 
		         ORDER BY id DESC LIMIT 10`
		rows, err = r.db.Query(query, senderId, receiverId, receiverId, senderId)
	} else {
		query = `SELECT id, sender_id, receiver_id, message, created_at 
		         FROM messages 
		         WHERE ((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) 
		         AND id < ? 
		         ORDER BY id DESC LIMIT 10`
		rows, err = r.db.Query(query, senderId, receiverId, receiverId, senderId, lastMsgId)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*PrivateMessage
	for rows.Next() {
		var message PrivateMessage
		if err := rows.Scan(&message.Id, &message.SenderId, &message.ReceiverId, &message.Message, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}

// ✅ Get 10 Group Messages (with pagination)
func (r *ChatRepository) GetGroupMessages(groupId int64, lastMsgId int64) ([]*GroupChat, error) {
	var query string
	var rows *sql.Rows
	var err error

	if lastMsgId == 0 {
		query = `SELECT id, sender_id, group_id, content, sent_at 
		         FROM group_messages 
		         WHERE group_id = ? 
		         ORDER BY id DESC LIMIT 10`
		rows, err = r.db.Query(query, groupId)
	} else {
		query = `SELECT id, sender_id, group_id, content, sent_at 
		         FROM group_messages 
		         WHERE group_id = ? AND id < ? 
		         ORDER BY id DESC LIMIT 10`
		rows, err = r.db.Query(query, groupId, lastMsgId)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*GroupChat
	for rows.Next() {
		var message GroupChat
		if err := rows.Scan(&message.Id, &message.SenderId, &message.GroupId, &message.Message, &message.SentAt); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}

func (r *ChatRepository) GetPrivateChats(userId int64) ([]*PrivateMessage, error) {
	query := `SELECT m.id, m.sender_id, m.receiver_id, m.message, m.created_at FROM messages m WHERE m.receiver_id = ? OR m.sender_id = ? ORDER BY m.created_at DESC`
	rows, err := r.db.Query(query, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*PrivateMessage
	for rows.Next() {
		var message PrivateMessage
		err := rows.Scan(&message.Id, &message.SenderId, &message.ReceiverId, &message.Message, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}

func (r *ChatRepository) GetGroupChats(userId int64) ([]*GroupChat, error) {
	query := `SELECT gm.id, gm.sender_id, gm.group_id, gm.content, gm.sent_at FROM group_messages gm JOIN group_members gmbr ON gm.group_id = gmbr.group_id WHERE gmbr.user_id = ? ORDER BY gm.sent_at DESC`
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*GroupChat
	for rows.Next() {
		var message GroupChat
		err := rows.Scan(&message.Id, &message.SenderId, &message.GroupId, &message.Message, &message.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}
