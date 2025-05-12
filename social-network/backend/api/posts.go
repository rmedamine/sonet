package api

import (
	"math"
	"net/http"
	"strconv"

	"social/config"
	"social/models"
	"social/utils"
)

type IndexStruct struct {
	Posts       []*models.Post
	TotalPages  int
	CurrentPage int
	Query       string
	Option      int
}

func GetPostsApi(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	userId := session.UserId
	pageStr := r.URL.Query().Get("page")
	currPage, err := strconv.Atoi(pageStr)
	if err != nil || currPage < 1 {
		currPage = 1
	}
	limit := config.LIMIT_PER_PAGE
	postRep := models.NewPostRepository()
	posts, err := getPosts(currPage, limit, userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	count, err := postRep.Count()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	page := NewPageStruct("Social Network", "", nil)
	page.Data = IndexStruct{
		Posts:       posts,
		TotalPages:  int(math.Ceil(float64(count) / config.LIMIT_PER_PAGE)),
		CurrentPage: currPage,
	}
	utils.WriteJSON(w, http.StatusOK, "OK", page)
}

func getPosts(currPage, limit int, userId int64) ([]*models.Post, error) {
	postRep := models.NewPostRepository()
	posts, err := postRep.GetPostPerPage(currPage, limit, userId)
	if err != nil {
		return nil, err
	}
	for _, post := range posts {
		post.Content = post.Content[0:min(len(post.Content), 200)]
	}
	return posts, nil
}
