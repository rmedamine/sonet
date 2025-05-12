package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"social/config"
	"social/models"
	"social/services"
	"social/utils"
)

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	var group models.Group

	groupRepo := models.NewGroupRepository()

	group.Description = strings.TrimSpace(r.FormValue("description"))
	group.Title = strings.TrimSpace(r.FormValue("title"))

	if IsNamed, err := groupRepo.IsGroupNameExists(group.Title); err != nil || IsNamed {
		msg := "Group name does not exist"
		status := http.StatusBadRequest
		if err != nil {
			msg = "Something went wrong"
			status = http.StatusInternalServerError
		}
		utils.WriteJSON(w, status, msg, nil)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "File too large", nil)
		return
	}
	file, handler, err := r.FormFile("image")
	if err != nil && file != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid file upload", nil)
		return
	}

	if file != nil {
		defer file.Close()
		imagePath, status := services.Upload(file, handler, "group")
		if status != http.StatusOK {
			utils.WriteJSON(w, status, imagePath, nil)
			return
		}
		group.Image = imagePath
	}
	group.CreatorID = session.UserId
	rGroup, err := groupRepo.CreateGroup(&group)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", rGroup)
}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	group, err := groupRepo.GetGroup(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", group)
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	groups, err := groupRepo.GetGroups(session.UserId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", groups)
}

func RequestJoinGroup(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isMember, err := groupRepo.IsUserMember(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if isMember {
		utils.WriteJSON(w, http.StatusBadRequest, "You are already a member of this group", nil)
		return
	}

	alreadyRequeted, err := groupRepo.IsUserRequestedToJoin(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if alreadyRequeted {
		utils.WriteJSON(w, http.StatusBadRequest, "You have already requested to join this group", nil)
		return
	}

	err = groupRepo.RequestJoinGroup(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func AcceptJoinRequest(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isCreator, err := groupRepo.IsUserCreator(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isCreator {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	isRequested, err := groupRepo.IsUserRequestedToJoin(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isRequested {
		utils.WriteJSON(w, http.StatusBadRequest, "User has not requested to join this group", nil)
		return
	}
	err = groupRepo.ApproveJoinGroupRequest(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func RejectJoinRequest(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isCreator, err := groupRepo.IsUserCreator(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isCreator {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	isRequested, err := groupRepo.IsUserRequestedToJoin(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isRequested {
		utils.WriteJSON(w, http.StatusBadRequest, "User has not requested to join this group", nil)
		return
	}
	err = groupRepo.RejectJoinGroupRequest(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func InviteToGroup(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isMember, err := groupRepo.IsUserMember(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isMember {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	isMember, err = groupRepo.IsUserMember(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if isMember {
		utils.WriteJSON(w, http.StatusBadRequest, "User is already a member of this group", nil)
		return
	}
	err = groupRepo.InviteToGroup(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Send notification to the invited user
	GroupInviteNotify(userId, groupId, session.UserId)

	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func AcceptInvite(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isInvited, err := groupRepo.IsUserInvited(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isInvited {
		utils.WriteJSON(w, http.StatusBadRequest, "User has not been invited to this group", nil)
		return
	}

	err = groupRepo.AcceptGroupInvitation(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func RejectInvite(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isInvited, err := groupRepo.IsUserInvited(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isInvited {
		utils.WriteJSON(w, http.StatusBadRequest, "User has not been invited to this group", nil)
		return
	}
	isCreator, err := groupRepo.IsUserCreator(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isCreator {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	err = groupRepo.RejectGroupInvitation(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func GetUserGroupInvites(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Only allow users to view their own invites
	if userId != session.UserId {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have permission to view these invites", nil)
		return
	}

	groupRepo := models.NewGroupRepository()
	invites, err := groupRepo.GetGroupInvitations(userId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "OK", invites)
}

// func JoinGroup(w http.ResponseWriter, r *http.Request) {
// 	sessionId := utils.GetSessionCookie(r)
// 	session := config.IsAuth(sessionId)
// 	if session == nil {
// 		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
// 		return
// 	}
// 	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
// 	if err != nil {
// 		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
// 		return
// 	}
// 	groupRepo := models.NewGroupRepository()
// 	isGroupExists, err := groupRepo.IsGroupExists(groupId)
// 	if err != nil {
// 		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
// 		return
// 	}
// 	if !isGroupExists {
// 		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
// 		return
// 	}
// 	isMember, err := groupRepo.IsUserMember(session.UserId, groupId)
// 	if err != nil {
// 		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
// 		return
// 	}
// 	if isMember {
// 		utils.WriteJSON(w, http.StatusBadRequest, "You are already a member of this group", nil)
// 		return
// 	}
// 	err = groupRepo.JoinGroup(session.UserId, groupId)
// 	if err != nil {

// 		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
// 		return
// 	}
// 	utils.WriteJSON(w, http.StatusOK, "OK", nil)
// }

func LeaveGroup(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isMember, err := groupRepo.IsUserMember(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isMember {
		utils.WriteJSON(w, http.StatusBadRequest, "You are not a member of this group", nil)
		return
	}
	err = groupRepo.LeaveGroup(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func KickMember(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	userId, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isCreator, err := groupRepo.IsUserCreator(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isCreator {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	isMember, err := groupRepo.IsUserMember(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isMember {
		utils.WriteJSON(w, http.StatusBadRequest, "User is not a member of this group", nil)
		return
	}
	err = groupRepo.KickMember(userId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}
	isCreator, err := groupRepo.IsUserCreator(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isCreator {
		utils.WriteJSON(w, http.StatusForbidden, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	err = groupRepo.DeleteGroup(session.UserId, groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	isGroupExists, err := groupRepo.IsGroupExists(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if !isGroupExists {
		utils.WriteJSON(w, http.StatusNotFound, "Group not found", nil)
		return
	}

	events, err := groupRepo.GetGroupEvents(groupId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", events)
}

func CreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}

	var createEvent models.CreateEvent
	if err := json.NewDecoder(r.Body).Decode(&createEvent); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Bad Request", nil)
		return
	}

	event, err := ParseDates(createEvent)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Error parsing dates", nil)
		return
	}
	groupRepo := models.NewGroupRepository()

	if IsUserMember, err := groupRepo.IsUserMember(event.UserId, event.GroupID); err != nil || !IsUserMember {
		msg := "Group not exist or user does not a Member"
		status := http.StatusNotFound
		if err != nil {
			msg = "Something went wrong"
			status = http.StatusInternalServerError
		}
		utils.WriteJSON(w, status, msg, nil)
		return
	}

	if err := groupRepo.CreateGroupEvent(&event); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Sprintf("error : %v", err), nil)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}

func ParseDates(createEvent models.CreateEvent) (models.Event, error) {
	var err error
	event := models.Event{
		GroupID:     createEvent.GroupID,
		UserId:      createEvent.UserId,
		Title:       createEvent.Title,
		Description: createEvent.Description,
	}
	event.Start, err = time.Parse("2006-01-02T15:04", createEvent.StartStr)
	if err != nil {
		return event, err
	}
	event.End, err = time.Parse("2006-01-02T15:04", createEvent.EndStr)
	if err != nil {
		return event, err
	}
	return event, nil
}

func RespondToGroupEvent(w http.ResponseWriter, r *http.Request) {
	sessionId := utils.GetSessionCookie(r)
	session := config.IsAuth(sessionId)
	if session == nil {
		utils.WriteJSON(w, http.StatusUnauthorized, "You don't have the necessary permissions to access this. Please log in or check your access rights.", nil)
		return
	}
	groupId, err := strconv.ParseInt(r.PathValue("groupId"), 10, 64)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid group ID", nil)
		return
	}
	groupRepo := models.NewGroupRepository()
	if isGroupExists, err := groupRepo.IsGroupExists(groupId); err != nil || !isGroupExists {
		msg := "Group not found"
		status := http.StatusNotFound
		if err != nil {
			msg = err.Error()
			status = http.StatusInternalServerError
		}
		utils.WriteJSON(w, status, msg, nil)
		return
	}

	var eventresponse models.EventResponse
	if err := json.NewDecoder(r.Body).Decode(&eventresponse); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Bad Request", nil)
		return
	}
	fmt.Println(groupId, session.UserId, eventresponse.EventID, eventresponse.Response)

	if isMember, err := groupRepo.IsUserMember(session.UserId, groupId); err != nil || !isMember {
		msg := "You are not a member of this group"
		status := http.StatusBadRequest
		if err != nil {
			msg = err.Error()
			status = http.StatusInternalServerError
		}
		utils.WriteJSON(w, status, msg, nil)
		return
	}
	err = groupRepo.SaveEventResponse(groupId, session.UserId, eventresponse.EventID, eventresponse.Response)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "OK", nil)
}
