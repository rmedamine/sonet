package api

import (
	"net/http"
	"strconv"

	"social/config"
	"social/models"
	"social/utils"
)

func ReactToPost(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId
	postId, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	reaction := r.URL.Query().Get("reaction")
	if reaction != "LIKE" && reaction != "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid reaction", nil)
		return
	}
	reactionRepo := models.NewReactionRepository()
	err = reactionRepo.ReactToPost(userId, postId, reaction)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	postRepo := models.NewPostRepository()
	post, err := postRepo.GetPostById(postId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if reaction == "LIKE" && post.UserID != userId {
		LikeNotify(post.UserID, post.ID)
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func ReactToComment(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId
	commentId, err := strconv.ParseInt(r.PathValue("commentId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	reaction := r.URL.Query().Get("reaction")
	if reaction != "LIKE" && reaction != "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid reaction", nil)
		return
	}
	reactionRepo := models.NewReactionRepository()
	err = reactionRepo.ReactToComment(userId, commentId, reaction)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}
