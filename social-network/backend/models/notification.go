package models

import (
	"database/sql"
	"fmt"
	"time"

	"social/config"
)

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	UserName  string    `json:"userName,omitempty"`
	Type      string    `json:"type"`
	TargetID  int64     `json:"targetId,omitempty"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{
		db: config.DB,
	}
}

func (n *NotificationRepository) CreateNotification(notification *Notification) error {
	query := `INSERT INTO notifications (user_id, type, target_id, content, is_read, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := n.db.Exec(query, notification.UserID, notification.Type, notification.TargetID, notification.Content, notification.IsRead, time.Now())
	if err != nil {
		fmt.Println(err)
		return err
	}
	notification.CreatedAt = time.Now()
	notification.ID, err = result.LastInsertId()
	return err
}

func (repo *NotificationRepository) GetNotifications(userID int64) ([]Notification, error) {
	query := `SELECT n.id, n.user_id, COALESCE(NULLIF(u.nickname, ''), u.firstname || ' ' || u.lastname) AS username, n.type, n.target_id, n.content, n.is_read, n.created_at 
			  FROM notifications n 
			  JOIN users u ON n.user_id = u.id 
			  WHERE n.user_id = ? 
			  ORDER BY n.created_at DESC`
	rows, err := repo.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		err := rows.Scan(&notification.ID, &notification.UserID, &notification.UserName, &notification.Type,
			&notification.TargetID, &notification.Content, &notification.IsRead, &notification.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (repo *NotificationRepository) MarkAsRead(notificationID int64) error {
	query := `UPDATE notifications SET is_read = 1 WHERE id = ?`
	_, err := repo.db.Exec(query, notificationID)
	return err
}

func (repo *NotificationRepository) MarkAllAsRead(userID int64) error {
	query := `UPDATE notifications SET is_read = 1 WHERE user_id = ?`
	_, err := repo.db.Exec(query, userID)
	return err
}

func (repo *NotificationRepository) ClearNotification(notificationID int64) error {
	query := `DELETE FROM notifications WHERE id = ?`
	_, err := repo.db.Exec(query, notificationID)
	return err
}

func (repo *NotificationRepository) ClearAllNotifications(userID int64) error {
	query := `DELETE FROM notifications WHERE user_id = ?`
	_, err := repo.db.Exec(query, userID)
	return err
}

func (repo *NotificationRepository) GetUnreadCount(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`
	var count int
	err := repo.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
