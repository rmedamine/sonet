package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"social/config"
	"social/models"
	"social/utils"
)

func FollowUser(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	var follow models.Follow
	var err error
	userRepo := models.NewUserRepository()
	follow.FollowingID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	isExist, err := userRepo.UserExistsById(follow.FollowingID)
	if err != nil {
		log.Println(err)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "It looks like the user you're trying to follow doesn't exist.", nil)
		return
	}

	follow.FollowerID = session.UserId
	if follow.FollowingID == follow.FollowerID {
		utils.WriteJSON(w, http.StatusBadRequest, "You can't follow yourself.", nil)
		return
	}

	followRepo := models.NewFollowRepository()
	isFollowing, err := followRepo.IsFollowing(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if isFollowing {
		utils.WriteJSON(w, http.StatusBadRequest, "You're already following this user.", nil)
		return
	}
	err = followRepo.Create(&follow)
	follow.CreatedAt = time.Now()
	if err != nil {
		log.Println(err)
		return
	}

	profileRepo := models.NewProfileRepository()
	isPublic, err := profileRepo.IsPublic(follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if isPublic {
		err = followRepo.AcceptRequest(follow.FollowerID, follow.FollowingID)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	follow.FollowerName, _ = userRepo.GetName(follow.FollowerID)
	follow.FollowingName, _ = userRepo.GetName(follow.FollowingID)
	FollowNotifiy(follow.FollowingID, follow.FollowerID)
	utils.WriteJSON(w, 200, "You're now following this user! Thanks for connecting!", follow)
}

func AcceptFollowRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var follow models.Follow
	var err error
	follow.FollowerID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	follow.FollowingID = session.UserId
	followRepo := models.NewFollowRepository()
	isExist, err := followRepo.IsFollowing(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "You're not following this user.", nil)
		return
	}
	err = followRepo.AcceptRequest(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "You've accepted the follow request. You're now following this user.", nil)
}

func RejectRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var follow models.Follow
	var err error
	follow.FollowerID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	follow.FollowingID = session.UserId
	followRepo := models.NewFollowRepository()
	isExist, err := followRepo.IsFollowing(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "You're not following this user.", nil)
		return
	}
	pending, err := followRepo.FollowRequestExists(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !pending {
		utils.WriteJSON(w, http.StatusBadRequest, "There's no pending follow request from this user.", nil)
		return
	}
	err = followRepo.RejectRequest(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "You've rejected the follow request.", nil)
}

func Unfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var follow models.Follow
	var err error
	follow.FollowingID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	follow.FollowerID = session.UserId
	followRepo := models.NewFollowRepository()
	isExist, err := followRepo.IsFollowing(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "You're not following this user.", nil)
		return
	}
	err = followRepo.Unfollow(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "You've unfollowed this user.", nil)
}

func RemoveFollower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var follow models.Follow
	var err error
	follow.FollowerID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	follow.FollowingID = session.UserId
	followRepo := models.NewFollowRepository()
	isExist, err := followRepo.IsFollowing(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isExist {
		utils.WriteJSON(w, http.StatusBadRequest, "You're not following this user.", nil)
		return
	}
	err = followRepo.Unfollow(follow.FollowerID, follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "You've removed this user from your followers.", nil)
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var follow models.Follow
	var err error
	follow.FollowingID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	followRepo := models.NewFollowRepository()
	followers, err := followRepo.GetFollowers(follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "Followers retrieved successfully.", followers)
}

func GetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	var follow models.Follow
	var err error
	follow.FollowerID, err = strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID. Please provide a valid user ID.", nil)
		return
	}
	followRepo := models.NewFollowRepository()
	following, err := followRepo.GetFollowing(follow.FollowerID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "Following retrieved successfully.", following)
}

func GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "The HTTP method used in the request is invalid. Please ensure you're using the correct method.", nil)
		return
	}
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	var follow models.Follow
	var err error
	follow.FollowingID = session.UserId
	followRepo := models.NewFollowRepository()
	requests, err := followRepo.GetFollowRequests(follow.FollowingID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, 200, "Follow requests retrieved successfully.", requests)
}
