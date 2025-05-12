package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"social/config"
	"social/models"
	"social/utils"
)

const (
	// Notification types
	CommentNotification     = "COMMENT"
	LikeNotification        = "LIKE"
	FollowNotification      = "FOLLOW"
	MessageNotiffication    = "MESSAGE"
	GroupInviteNotification = "GROUP_INVITE"
)

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId
	notificationRepo := models.NewNotificationRepository()
	notifications, err := notificationRepo.GetNotifications(userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", notifications)
}

func ReadNotification(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	notificationId, err := strconv.ParseInt(r.PathValue("notificationId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	notificationRepo := models.NewNotificationRepository()
	err = notificationRepo.MarkAsRead(notificationId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId
	notificationRepo := models.NewNotificationRepository()
	err := notificationRepo.MarkAllAsRead(userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func ClearNotification(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	notificationId, err := strconv.ParseInt(r.PathValue("notificationId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	notificationRepo := models.NewNotificationRepository()
	err = notificationRepo.ClearNotification(notificationId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func ClearAllNotifications(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId
	notificationRepo := models.NewNotificationRepository()
	err := notificationRepo.ClearAllNotifications(userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func FollowNotifiy(userId int64, followerId int64) {
	notification := &models.Notification{
		UserID:    userId,
		TargetID:  followerId,
		Type:      FollowNotification,
		Content:   "You have a new follow request",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	NotificationQueue <- notification
}

func CommentNotify(userId int64, postId int64) {
	notification := &models.Notification{
		UserID:    userId,
		TargetID:  postId,
		Type:      CommentNotification,
		Content:   "You have a new comment",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	NotificationQueue <- notification
}

func LikeNotify(userId int64, postId int64) {
	notification := &models.Notification{
		UserID:    userId,
		TargetID:  postId,
		Type:      LikeNotification,
		Content:   "Your post has been liked",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	NotificationQueue <- notification
}

func MessageNotify(userId int64, receiverId int64) {
	notification := &models.Notification{
		UserID:    userId,
		TargetID:  receiverId,
		Type:      MessageNotiffication,
		Content:   "You received a message",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	NotificationQueue <- notification
}

func GroupMessageNotify(userId int64, groupId int64) {
	notification := &models.Notification{
		UserID:    userId,
		TargetID:  groupId,
		Type:      MessageNotiffication,
		Content:   "You received a message",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	NotificationQueue <- notification
}

func GroupInviteNotify(userId int64, groupId int64, senderId int64) {
	// Get group details
	groupRepo := models.NewGroupRepository()
	group, err := groupRepo.GetGroup(groupId)
	if err != nil {
		log.Printf("Error getting group details: %v", err)
		return
	}

	// Get sender details
	userRepo := models.NewUserRepository()
	sender, err := userRepo.GetUserByID(senderId)
	if err != nil {
		log.Printf("Error getting sender details: %v", err)
		return
	}

	notification := &models.Notification{
		UserID:    userId,
		TargetID:  groupId,
		Type:      GroupInviteNotification,
		Content:   fmt.Sprintf("%s invited you to join the group '%s'", sender.Firstname+" "+sender.Lastname, group.Title),
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	NotificationQueue <- notification
}
