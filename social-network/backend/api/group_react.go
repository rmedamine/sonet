package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social/config"
	"social/models"
	"social/utils"
)

// ReactToGroupPost handles reactions to group posts
func ReactToGroupPost(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId

	groupPostId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	reaction := r.URL.Query().Get("reaction")
	if reaction != "LIKE" && reaction != "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid reaction", nil)
		return
	}

	reactionRepo := models.NewGroupReactionRepository()
	err = reactionRepo.ReactToGroupPost(userId, groupPostId, reaction)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	groupPostRepo := models.NewPostGroupRepository()
	_, err = groupPostRepo.GetPostById(groupPostId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

// ReactToGroupComment handles reactions to group comments
func ReactToGroupComment(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId

	commentId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	reaction := r.URL.Query().Get("reaction")
	if reaction != "LIKE" && reaction != "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid reaction", nil)
		return
	}

	reactionRepo := models.NewGroupReactionRepository()
	err = reactionRepo.ReactToGroupComment(userId, commentId, reaction)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

// AddGroupComment handles adding comments to group posts
func AddGroupComment(w http.ResponseWriter, r *http.Request) {
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

	var comment models.GroupComment
	userRepo := models.NewUserRepository()
	err := utils.ReadJSON(r, &comment)
	comment.UserId = session.UserId
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if strings.TrimSpace(comment.Comment) == "" || !utils.IsBetween(comment.Comment, 2, 1000) {
		utils.WriteJSON(w, http.StatusBadRequest, "The comment must be between 2 and 1000 characters", nil)
		return
	}

	postGroupRepo := models.NewPostGroupRepository()
	commentRepo := models.NewGroupCommentRepository()

	isExist, err := postGroupRepo.IsPostExist(comment.GroupPostId)
	if err != nil {
		log.Println(err)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "It looks like the group post you're trying to comment on doesn't exist anymore.", nil)
		return
	}

	err = commentRepo.Create(&comment)
	comment.CreatedAt = time.Now()
	if err != nil {
		log.Println(err)
		return
	}

	comment.Name, _ = userRepo.GetName(int64(comment.UserId))
	comment.Avatar, _ = userRepo.GetAvatar(int64(comment.UserId))

	_, err = postGroupRepo.GetPostById(comment.GroupPostId)
	if err != nil {
		log.Println(err)
		return
	}

	utils.WriteJSON(w, 200, "Your comment has been added successfully! Thanks for sharing your thoughts!", comment)
}

// GetGroupPostComments handles retrieving comments for a group post
func GetGroupPostComments(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId

	groupPostId, err := strconv.ParseInt(r.PathValue("groupPostId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	commentRepo := models.NewGroupCommentRepository()
	comments, err := commentRepo.GetGroupPostComments(groupPostId, userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Comments retrieved successfully", comments)
}
