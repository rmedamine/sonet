package models

import (
	"database/sql"
	"fmt"
	"time"

	"social/config"
)

type GroupInvitation struct {
	ID         int64     `json:"id"`
	GroupId    int64     `json:"group_id"`
	SenderId   int64     `json:"sender_id"`
	ReceiverId int64     `json:"receiver_id"`
	IsRead     int       `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type GroupJoinRequest struct {
	ID        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	GroupId   int64     `json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Updated Group struct
type Group struct {
	ID          int64     `json:"id"`
	CreatorID   int64     `json:"creatorId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"createdAt"`
	IsMember    bool      `json:"isMember"`
	IsInvited   bool      `json:"isInvited"`
	IsRequested bool      `json:"isRequested"`
	Role        string    `json:"role,omitempty"`

	Invites  []GroupInvitation  `json:"invites,omitempty"`
	Requests []GroupJoinRequest `json:"requests,omitempty"` // Fixed typo in field name
}

type Event struct {
	ID          int64     `json:"id,omitempty"`
	GroupID     int64     `json:"groupId,omitempty"`
	UserId      int64     `json:"userId,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	Start       time.Time `json:"start,omitempty"`
	End         time.Time `json:"end,omitempty"`
}

type CreateEvent struct {
	GroupID     int64  `json:"groupId,omitempty"`
	UserId      int64  `json:"userId,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	StartStr    string `json:"event_date_start,omitempty"`
	EndStr      string `json:"event_date_end,omitempty"`
}

type EventResponse struct {
	GroupID   int64     `json:"groupId,omitempty"`
	UserID    int64     `json:"userId,omitempty"`
	EventID   int64     `json:"eventId,omitempty"`
	Response  string    `json:"response,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository() *GroupRepository {
	return &GroupRepository{db: config.DB}
}

func (r *GroupRepository) IsGroupNameExists(groupName string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM groups g WHERE g.title like @name)`
	var IsNamed bool
	err := r.db.QueryRow(query, sql.Named("name", groupName)).Scan(&IsNamed)
	if err != nil || IsNamed {
		return IsNamed, err
	}
	return IsNamed, nil
}

func (r *GroupRepository) CreateGroup(group *Group) (Group, error) {
	rGroup := *group
	query := `
	INSERT INTO groups (creatorId, title, description, image, createdAt) 
	VALUES (?, ?, ?, ?, ?) 
	RETURNING id, createdAt
	`

	err := r.db.QueryRow(query, group.CreatorID, group.Title, group.Description, group.Image, time.Now()).Scan(&rGroup.ID, &rGroup.CreatedAt)
	if err != nil {
		return rGroup, err
	}
	query = `INSERT INTO group_members (user_id, group_id, role) VALUES (?, ?, 'creator')`
	_, err = r.db.Exec(query, group.CreatorID, rGroup.ID)
	if err != nil {
		return rGroup, err
	}
	return rGroup, nil
}

func (r *GroupRepository) DeleteGroup(userId, groupId int64) error {
	query := `DELETE FROM groups WHERE id = ? AND creatorId = ?`
	_, err := r.db.Exec(query, groupId, userId)
	return err
}

func (r *GroupRepository) GetGroups(userId int64) ([]Group, error) {
	query := `SELECT id, creatorId, title, description, image, createdAt FROM groups`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var group Group
		err := rows.Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.Image, &group.CreatedAt)
		if err != nil {
			return nil, err
		}
		isMember, err := r.IsMember(group.ID, userId)
		if err != nil {
			return nil, err
		}

		isInvited, err := r.IsUserInvited(group.ID, userId)
		if err != nil {
			return nil, err
		}

		isRequested, err := r.IsUserRequestedToJoin(userId, group.ID)
		if err != nil {
			return nil, err
		}

		isCreator, err := r.IsUserCreator(group.ID, userId)
		if err != nil {
			return nil, err
		}

		group.IsMember = isMember
		group.IsInvited = isInvited
		group.IsRequested = isRequested
		if isCreator {
			group.Role = "creator"
		} else {
			group.Role = "member"
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// IsMember checks if a user is a member of a specific group
func (r *GroupRepository) IsMember(groupId int64, userId int64) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM group_members 
		WHERE group_id = ? AND user_id = ?`

	var count int
	err := r.db.QueryRow(query, groupId, userId).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *GroupRepository) GetGroup(id int64) (*Group, error) {
	// Get the basic group information
	query := `SELECT id, creatorId, title, description, image, createdAt FROM groups WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var group Group
	err := row.Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.Image, &group.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Get the group requests with all fields
	requestsQuery := `
        SELECT id, user_id, group_id, created_at
        FROM group_join_requests 
        WHERE group_id = ?
    `
	requestRows, err := r.db.Query(requestsQuery, id)
	if err != nil {
		return nil, err
	}
	defer requestRows.Close()

	for requestRows.Next() {
		var request GroupJoinRequest
		err := requestRows.Scan(&request.ID, &request.UserId, &request.GroupId, &request.CreatedAt)
		if err != nil {
			return nil, err
		}
		group.Requests = append(group.Requests, request)
	}

	if err = requestRows.Err(); err != nil {
		return nil, err
	}

	// Get the group invitations with all fields
	invitesQuery := `
        SELECT id, group_id, sender_id, receiver_id, is_read, created_at
        FROM group_invitations 
        WHERE group_id = ?
    `
	inviteRows, err := r.db.Query(invitesQuery, id)
	if err != nil {
		return nil, err
	}
	defer inviteRows.Close()

	for inviteRows.Next() {
		var invite GroupInvitation
		err := inviteRows.Scan(
			&invite.ID,
			&invite.GroupId,
			&invite.SenderId,
			&invite.ReceiverId,
			&invite.IsRead,
			&invite.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		group.Invites = append(group.Invites, invite)
	}

	if err = inviteRows.Err(); err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *GroupRepository) JoinGroup(userId, groupId int64) error {
	query := `INSERT INTO group_members (user_id, group_id, role) VALUES (?, ?, 'member')`
	_, err := r.db.Exec(query, userId, groupId)
	return err
}

func (r *GroupRepository) LeaveGroup(userId, groupId int64) error {
	query := `DELETE FROM group_members WHERE user_id = ? AND group_id = ?`
	_, err := r.db.Exec(query, userId, groupId)
	return err
}

func (r *GroupRepository) InviteToGroup(userId, groupId int64) error {
	query := `INSERT INTO group_invitations (receiver_id, group_id, sender_id) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, userId, groupId, userId)
	return err
}

func (r *GroupRepository) DeleteGroupInvitation(userId, groupId int64) error {
	query := `DELETE FROM group_invitations WHERE userId = ? AND groupId = ?`
	_, err := r.db.Exec(query, userId, groupId)
	if err != nil {
		return err
	}
	return r.JoinGroup(userId, groupId)
}

func (r *GroupRepository) AcceptGroupInvitation(userId, groupId int64) error {
	query := `DELETE FROM group_invitations WHERE receiver_id = ? AND group_id = ?`
	_, err := r.db.Exec(query, userId, groupId)
	if err != nil {
		return err
	}
	return r.JoinGroup(userId, groupId)
}

func (r *GroupRepository) RejectGroupInvitation(userId, groupId int64) error {
	query := `DELETE FROM group_invitations WHERE receiver_id = ? AND group_id = ?`
	_, err := r.db.Exec(query, userId, groupId)
	return err
}

func (r *GroupRepository) KickMember(userId, groupId int64) error {
	query := `DELETE FROM group_members WHERE userId = ? AND groupId = ?`
	_, err := r.db.Exec(query, userId, groupId)
	if err != nil {
		return err
	}
	return nil
}

func (r *GroupRepository) GetGroupMembers(groupId int64) ([]User, error) {
	query := `
		SELECT u.id, u.firstname, u.lastname, u.email, u.nickname, u.avatar, u.created_at
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
	`
	rows, err := r.db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []User
	for rows.Next() {
		var member User
		err := rows.Scan(&member.ID, &member.Firstname, &member.Lastname, &member.Email, &member.Nickname, &member.Avatar, &member.CreatedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *GroupRepository) GetGroupInvitations(userId int64) ([]Group, error) {
	query := `
		SELECT g.id, g.creatorId, g.title, g.description, g.image, g.createdAt
		FROM groups g
		JOIN group_invitations gi ON g.id = gi.group_id
		WHERE gi.receiver_id = ?
	`
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var group Group
		err := rows.Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.Image, &group.CreatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *GroupRepository) GetGroupEvents(groupId int64) ([]Event, error) {
	query := `SELECT id, group_id, user_id,title, description, start_date, end_date, created_at FROM group_events WHERE group_id = ?`
	rows, err := r.db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.ID, &event.UserId, &event.GroupID, &event.Title, &event.Description, &event.Start, &event.End, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *GroupRepository) CreateGroupEvent(event *Event) error {
	query := `INSERT INTO group_events (group_id, user_id, title, description, start_date, end_date, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, event.GroupID, event.UserId, event.Title, event.Description, event.Start, event.End, time.Now())
	return err
}

func (r *GroupRepository) GetGroupEvent(groupId, eventId int64) (*Event, error) {
	query := `SELECT id, group_id, user_id, title, description, start_date, end_date, created_at FROM group_events WHERE group_id = ? AND id = ?`
	row := r.db.QueryRow(query, groupId, eventId)
	var event Event
	err := row.Scan(&event.ID, &event.GroupID, &event.UserId, &event.Title, &event.Description, &event.Start, &event.End, &event.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *GroupRepository) SaveEventResponse(groupId, userId, eventId int64, response string) error {
	query := `
		INSERT INTO group_event_responses (user_id, event_id, response, created_at, group_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, event_id) DO UPDATE
		SET response = excluded.response
	`
	_, err := r.db.Exec(query, userId, eventId, response, time.Now(), groupId)
	fmt.Println(err)
	return err
}

func (r *GroupRepository) GetGroupEventResponses(groupId, eventId int64) ([]EventResponse, error) {
	query := `SELECT user_id, event_id, group_id, response FROM event_responses WHERE event_id = ? AND group_id = ?`
	rows, err := r.db.Query(query, eventId, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var responses []EventResponse
	for rows.Next() {
		var response EventResponse
		err := rows.Scan(&response.UserID, &response.EventID, &response.Response)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (r *GroupRepository) GetGroupEventResponse(userId, eventId int64) (*EventResponse, error) {
	query := `SELECT user_id, event_id, response FROM group_event_responses WHERE user_id = ? AND event_id = ?`
	row := r.db.QueryRow(query, userId, eventId)
	var response EventResponse
	err := row.Scan(&response.UserID, &response.EventID, &response.Response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (r *GroupRepository) RequestJoinGroup(userId, groupId int64) error {
	query := `INSERT INTO group_join_requests (user_id, group_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, userId, groupId)
	return err
}

func (r *GroupRepository) ApproveJoinGroupRequest(userId, groupId int64) error {
	query := `DELETE FROM group_join_requests WHERE user_id = ? AND group_id = ?`
	_, err := r.db.Exec(query, userId, groupId)
	if err != nil {
		return err
	}
	return r.JoinGroup(userId, groupId)
}

func (r *GroupRepository) RejectJoinGroupRequest(userId, groupId int64) error {
	query := `DELETE FROM group_join_requests WHERE user_id = ? AND group_id = ?`
	_, err := r.db.Exec(query, userId, groupId)
	return err
}

func (r *GroupRepository) IsUserMember(userId, groupId int64) (bool, error) {
	query := `SELECT COUNT(*) FROM group_members WHERE user_id = ? AND group_id = ?`
	var count int
	err := r.db.QueryRow(query, userId, groupId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GroupRepository) IsUserCreator(userId, groupId int64) (bool, error) {
	query := `SELECT COUNT(*) FROM groups WHERE creatorId = ? AND id = ?`
	var count int
	err := r.db.QueryRow(query, userId, groupId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GroupRepository) IsUserInvited(userId, groupId int64) (bool, error) {
	query := `SELECT COUNT(*) FROM group_invitations WHERE receiver_id = ? AND group_id = ?`
	var count int
	err := r.db.QueryRow(query, userId, groupId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GroupRepository) IsUserRequestedToJoin(userId, groupId int64) (bool, error) {
	query := `SELECT COUNT(*) FROM group_join_requests WHERE user_id = ? AND group_id = ?`
	var count int
	err := r.db.QueryRow(query, userId, groupId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GroupRepository) IsGroupExists(groupId int64) (bool, error) {
	query := `SELECT COUNT(*) FROM groups WHERE id = ?`
	var count int
	err := r.db.QueryRow(query, groupId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
