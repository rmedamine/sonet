package api

import (
	"log"
	"net/http"
	"strconv"

	"social/config"
	"social/models"
	"social/utils"
)

type User struct {
	// 5od li briti hna bdl b full name wla user name
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

func UsersSearch(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Query parameter 'q' is required", nil)
		return
	}
	if len(query) > 30 {
		utils.WriteJSON(w, http.StatusBadRequest, "Query parameter 'q' must be at most 30 characters long", nil)
		return
	}

	userRepo = models.NewUserRepository()

	users, err := userRepo.SearchUser(query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		http.Error(w, "Error searching users", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		utils.WriteJSON(w, http.StatusNotFound, "No users found", nil)
		return
	}

	response := map[string]any{
		"success": true,
		"users":   users,
		"count":   len(users),
	}

	utils.WriteJSON(w, http.StatusOK, "OK", response)
}

func UserSearch(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Query parameter 'q' is required", nil)
		return
	}
	if len(query) > 20 {
		utils.WriteJSON(w, http.StatusBadRequest, "Query parameter 'q' must be at most 20 characters long", nil)
		return
	}
	groupId := r.URL.Query().Get("groupId")
	var groupIdInt int
	if groupId != "" {
		r, err := strconv.Atoi(groupId)
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, "Query parameter 'groupId' must be an integer", nil)
			return
		}
		groupIdInt = r
	}
	userRepo = models.NewUserRepository()
	groupRepo := models.NewGroupRepository()

	users, err := userRepo.SearchUsers(query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		http.Error(w, "Error searching users", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		utils.WriteJSON(w, http.StatusNotFound, "No users found", nil)
		return
	}

	filtredUsers := []models.User{}
	if groupIdInt > 0 {
		for _, user := range users {
			if user.ID == session.UserId {
				continue
			}
			isMember, err := groupRepo.IsMember(int64(groupIdInt), user.ID)
			if err != nil {
				log.Printf("Error checking group membership: %v", err)
				utils.WriteJSON(w, http.StatusInternalServerError, "Error checking group membership", nil)
				return
			}
			if isMember {
				continue
			}
			filtredUsers = append(filtredUsers, user)
		}
	} else {
		filtredUsers = users
	}

	response := map[string]any{
		"success": true,
		"users":   filtredUsers,
		"count":   len(filtredUsers),
	}

	utils.WriteJSON(w, http.StatusOK, "OK", response)
}
