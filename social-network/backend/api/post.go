package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"social/config"
	"social/models"
	"social/services"
	"social/utils"
)

type Page struct {
	Title string
	User  *config.Session
	Data  any
}

func NewPageStruct(title string, session string, data any) *Page {
	return &Page{
		Title: title,
		User:  config.IsAuth(session),
		Data:  data,
	}
}

func PostApi(w http.ResponseWriter, r *http.Request) {
	handlePost(w, r)
	// switch r.Method {
	// case http.MethodPost:
	// 	handlePost(w, r)
	// default:
	// 	utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
	// }
}

func handlePost(w http.ResponseWriter, r *http.Request) {
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
	var post models.Post
	userRepo := models.NewUserRepository()
	post.Content = r.FormValue("content")
	post.UserID = session.UserId
	if utils.IsEmpty(post.Content) {
		utils.WriteJSON(w, http.StatusBadRequest, "The Content cannot be empty. Please provide both and try again.", nil)
		return
	}
	privacy, err := strconv.ParseInt(r.FormValue("privacy"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid Post privacy", nil)
		return
	}
	post.Privacy = int(privacy)

	if privacy == 2 {
		if jsonUsers := r.FormValue("users"); jsonUsers != "" {
			if err = json.Unmarshal([]byte(jsonUsers), &post.Users); err != nil {
				fmt.Println(err.Error())
				utils.WriteJSON(w, http.StatusBadGateway, "Invalid Users", nil)
				return
			}
		}
	}

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

	post.Avatar, err = userRepo.GetAvatar(post.UserID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	err = services.CreateNewPost(&post, file, fileHeader)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	post.Name, err = userRepo.GetName(post.UserID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "The post has been created successfully.", post)
}

type PostData struct {
	Post     models.Post
	Comments []models.Comment
}

func GetPostApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSON(w, http.StatusUnauthorized, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	postId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	postRepo := models.NewPostRepository()
	comRepo := models.NewCommentRepository()
	post, err := postRepo.GetPostById(postId)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteJSON(w, http.StatusNotFound, "Not found", nil)
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	comment, err := comRepo.GetPostComments(postId, session.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	postData := PostData{
		Comments: comment,
		Post:     *post,
	}
	page := NewPageStruct(post.Title, "", postData)
	utils.WriteJSON(w, http.StatusOK, "OK", page)
}
