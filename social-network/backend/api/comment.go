package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"social/config"
	"social/models"
	"social/utils"
)

func AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var comment models.Comment
	userRepo := models.NewUserRepository()
	err := utils.ReadJSON(r, &comment)
	comment.UserID = session.UserId
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if strings.TrimSpace(comment.Comment) == "" || !utils.IsBetween(comment.Comment, 2, 1000) {
		utils.WriteJSON(w, http.StatusBadRequest, "The comment must be between 2 and 1000 characters", nil)
		return
	}
	postRepo := models.NewPostRepository()
	commentRepo := models.NewCommentRepository()
	isExist, err := postRepo.IsPostExist(comment.PostID)
	if err != nil {
		log.Println(err)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "It looks like the post you're trying to comment on doesn't exist anymore.", nil)
		return
	}
	err = commentRepo.Create(&comment)
	comment.CreatedAt = time.Now()
	if err != nil {
		log.Println(err)
		return
	}
	comment.Name, _ = userRepo.GetName(int64(comment.UserID))

	post, err := postRepo.GetPostById(comment.PostID)
	if err != nil {
		log.Println(err)
		return
	}

	if post.UserID != comment.UserID {
		CommentNotify(post.UserID, post.ID)
	}

	utils.WriteJSON(w, 200, "Your comment has been added successfully! Thanks for sharing your thoughts!", comment)
}
