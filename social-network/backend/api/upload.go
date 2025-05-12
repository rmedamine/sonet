package api

import (
	"encoding/json"
	"net/http"

	"social/config"
	"social/models"
	"social/services"
	"social/utils"
)

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get session token from cookie
	sessionToken := utils.GetSessionCookie(r)
	session, err := config.SESSION.GetSession(sessionToken)
	if err != nil || session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Please log in.", nil)
		return
	}

	// Retrieve user details
	userRepo := models.NewUserRepository()
	user, err := userRepo.GetUserByID(session.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusNotFound, "User not found", nil)
		return
	}

	// Parse multipart form (limit: 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "File too large", nil)
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid file upload", nil)
		return
	}
	defer file.Close()

	// Validate file extension
	randomFilename, ok := services.Upload(file, handler, "avatar")
	if ok != http.StatusOK {
		utils.WriteJSON(w, ok, randomFilename, nil)
		return
	}

	// Update the user's avatar in the database
	if err := userRepo.UpdateAvatar(user.ID, randomFilename); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Database error", nil)
		return
	}

	// Respond with success message
	response := map[string]string{
		"message":  "Avatar uploaded successfully",
		"filename": randomFilename,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
