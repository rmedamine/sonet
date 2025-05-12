package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"social/config"
	"social/models"
	"social/services"
	"social/utils"
)

// handleGroupPost handles creation of a new post in a group
func HandleGroupPost(w http.ResponseWriter, r *http.Request) {
	// Authentication check
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	session, _ = config.SESSION.GetSession(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusBadRequest, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	// Extract group ID from path or form
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}

	// Initialize post object
	var post models.PostGroup
	userRepo := models.NewUserRepository()
	post.Content = r.FormValue("content")
	post.UserId = session.UserId
	post.GroupId = groupId

	// Validate content
	if utils.IsEmpty(post.Content) {
		utils.WriteJSON(w, http.StatusBadRequest, "The Content cannot be empty. Please provide content and try again.", nil)
		return
	}

	// Handle file upload if present
	err = r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Error parsing form data", nil)
		return
	}
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		file, fileHeader = nil, nil // No image uploaded, continue without error
	} else {
		defer file.Close()
	}

	// Get user avatar for the post
	avatar, err := userRepo.GetAvatar(post.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	post.Avatar = avatar

	// Create the post
	err = services.CreateNewGroupPost(&post, file, fileHeader)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Get user name for the response
	post.Name, err = userRepo.GetName(post.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Return success response
	utils.WriteJSON(w, http.StatusCreated, "The group post has been created successfully.", post)
}

// GetGroupPostsApi handles fetching posts for a specific group with pagination
func GetGroupPostsApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}

	// Authentication check
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	// Extract group ID from path
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}

	// Parse page and limit parameters
	page := 1
	limit := 10
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = limitNum
		}
	}

	// Get posts with pagination
	posts, totalCount, err := services.GetGroupPosts(groupId, page, limit, session.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Prepare response data
	response := struct {
		Posts       []*models.PostGroup `json:"posts"`
		TotalCount  int                 `json:"totalCount"`
		CurrentPage int                 `json:"currentPage"`
		PerPage     int                 `json:"perPage"`
	}{
		Posts:       posts,
		TotalCount:  totalCount,
		CurrentPage: page,
		PerPage:     limit,
	}

	utils.WriteJSON(w, http.StatusOK, "OK", response)
}

// GetGroupPostApi handles fetching a single post with its comments
type GroupPostData struct {
	Post     models.PostGroup
	Comments []models.GroupComment
}

func GetGroupPostApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}

	// Authentication check
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	// Extract post ID from path
	postId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid post ID", nil)
		return
	}

	// Get the post
	post, err := services.GetGroupPostById(postId)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteJSON(w, http.StatusNotFound, "Post not found", nil)
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Get comments for the post
	comRepo := models.NewGroupCommentRepository()
	comments, err := comRepo.GetGroupPostComments(postId, session.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Prepare response data
	postData := GroupPostData{
		Post:     *post,
		Comments: comments,
	}

	page := NewGroupPostPage("Group Post", "", postData)
	utils.WriteJSON(w, http.StatusOK, "OK", page)
}

// NewGroupPostPage is a helper function to create a page structure for responses
// This function should match the one used in your original code
func NewGroupPostPage(title string, description string, data interface{}) interface{} {
	return struct {
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Data        interface{} `json:"data"`
	}{
		Title:       title,
		Description: description,
		Data:        data,
	}
}
