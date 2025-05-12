package api

import (
	"net/http"

	"social/config"
	"social/models"
	"social/utils"
)

/*
userchat >>>
*/
func PrivateChat(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "Unothorized", nil)
		return
	}

	followRepo := models.NewFollowRepository()
	allFollows, err := followRepo.GetFollowersFollowing(session.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Success", allFollows)
}
