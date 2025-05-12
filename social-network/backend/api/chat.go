package api

import (
	"fmt"
	"net/http"
	"strconv"

	"social/config"
	"social/models"
	"social/utils"
)

// MessageApi handles fetching private and group messages
func MessageApi(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)

	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Session not found", nil)
		return
	}
	chatRepo := models.NewChatRepository()
	userID := session.UserId
	userIDParam := r.URL.Query().Get("user_id")
	groupIDParam := r.URL.Query().Get("group_id")
	lastMsgID, _ := strconv.ParseInt(r.URL.Query().Get("last_msg_id"), 10, 64)
	if userIDParam != "" {
		otherUserID, err := strconv.ParseInt(userIDParam, 10, 64)
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
			return
		}
		messages, err := chatRepo.GetPrivateMessages(userID, otherUserID, lastMsgID)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Error fetching messages", nil)
			return
		}
		utils.WriteJSON(w, http.StatusOK, "Success", messages)
	} else if groupIDParam != "" {
		fmt.Println("groupIDParam", groupIDParam)
		groupID, err := strconv.ParseInt(groupIDParam, 10, 64)
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
			return
		}

		messages, err := chatRepo.GetGroupMessages(groupID, lastMsgID)
		if err != nil {
			fmt.Println("error", err.Error())
			utils.WriteJSON(w, http.StatusInternalServerError, "Error fetching group messages", nil)
			return
		}

		utils.WriteJSON(w, http.StatusOK, "Success", messages)

	} else {
		utils.WriteJSON(w, http.StatusBadRequest, "Missing user_id or group_id parameter", nil)
	}
}
