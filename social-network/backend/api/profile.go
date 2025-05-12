package api

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"social/config"
	"social/models"
	"social/services"
	"social/utils"
)

func Update(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		UpdateProfile(w, r)
	case http.MethodGet:
		GetUser(w, r)
	default:
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
	}
}

func Profile(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	profileRepo := models.NewProfileRepository()
	profile, err := profileRepo.GetProfile(session.UserId, userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", profile)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	profile, err := services.GetUserData(sessionId)
	if err != nil {
		if err.Error() == "invalid user session" {
			utils.WriteJSON(w, http.StatusUnauthorized, "Invalid user session", nil)
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "OK", profile)
}

// update profile

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	var user models.User
	profileRepo := models.NewUserRepository()
	user.ID = session.UserId
	user.Email = strings.TrimSpace(r.FormValue("email"))
	user.Firstname = strings.TrimSpace(r.FormValue("firstname"))
	user.Lastname = strings.TrimSpace(r.FormValue("lastname"))
	user.DateOfBirth = r.FormValue("date_of_birth")
	user.Nickname = strings.TrimSpace(r.FormValue("nickname"))
	user.About = strings.TrimSpace(r.FormValue("about"))
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "File too large", nil)
		return
	}
	user.Avatar = r.FormValue("avatar")
	isPublic := r.FormValue("is_public")
	user.IsPublic = true
	if isPublic == "true" {
		user.IsPublic = true
	} else if isPublic == "false" {
		user.IsPublic = false
	} else {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid profile type", nil)
		return
	}
	// Validate input
	if err := validateUser(&user, true); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	file, handler, err := r.FormFile("avatar")
	if err != nil && file != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid file upload", nil)
		return
	}

	if file != nil {
		defer file.Close()
		AvatarPath, status := services.Upload(file, handler, "avatar")
		if status != http.StatusOK {
			utils.WriteJSON(w, status, AvatarPath, nil)
			return
		}
		user.Avatar = AvatarPath
		oldAvatar, err := profileRepo.GetAvatar(user.ID)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		if oldAvatar != "" {
			err = os.Remove(oldAvatar)
			if err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
				return
			}
		}
		err = profileRepo.UpdateAvatar(user.ID, AvatarPath)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}
	err = profileRepo.UpdateUser(&user)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", user)
}
