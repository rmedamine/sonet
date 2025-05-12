package api

import (
	"net/http"
	"social/config"
	"social/utils"
	"time"
)

type LogoutResponse struct {
	Token string `json:"token"`
}

func deleteCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func LogoutApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSON(w, http.StatusUnauthorized, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	var token LogoutResponse
	if err := utils.ReadJSON(r, &token); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid JSON format.", nil)
		return
	}
	// sessionId := utils.GetSessionCookie(r)
	config.SESSION.DeleteSession(token.Token)
	deleteCookie(w)
	utils.WriteJSON(w, http.StatusOK, "You have successfully logged out.", nil)
}
