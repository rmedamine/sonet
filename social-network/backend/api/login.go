package api

import (
	"net/http"

	"social/config"
	"social/services"
	"social/utils"
)

type AuthResponse struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginApi(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := utils.ReadJSON(r, &creds)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid credentials", nil)
		return
	}
	user, err := services.LoginUser(creds.Email, creds.Password)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	session, err := config.SESSION.CreateSession(user.Email, int64(user.ID))
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "An error occurred while creating a session", nil)
		return
	}
	cookies := http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Domain:   "localhost",
	}
	http.SetCookie(w, &cookies)
	response := AuthResponse{
		SessionID: session.Token,
		UserID:    int(user.ID),
		Email:     user.Email,
	}
	utils.WriteJSON(w, http.StatusOK, "Login successful", response)
}
