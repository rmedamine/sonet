package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"social/config"
)

var (
	ErrDB           = errors.New("database error")
	ErrUserNotFound = errors.New("user not found")
)

type Profile struct {
	ID          int      `json:"id"`
	Email       string   `json:"email"`
	UserId      int64    `json:"userId"`
	Firstname   string   `json:"firstname"`
	Lastname    string   `json:"lastname"`
	DateOfBirth string   `json:"date_of_birth"`
	Nickname    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
	About       string   `json:"about"`
	IsPublic    bool     `json:"is_public"`
	CreatedAt   string   `json:"created_at"`
	Posts       []*Post  `json:"posts"`
	Followers   []Follow `json:"followers"`
	Following   []Follow `json:"following"`
}

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository() *ProfileRepository {
	return &ProfileRepository{db: config.DB}
}

func (r *ProfileRepository) GetMyProfile(userId int64) (*Profile, error) {
	var profile Profile
	postRepo := NewPostRepository()
	followRepo := NewFollowRepository()
	query := `
		SELECT id, email, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(query, userId).Scan(&profile.ID, &profile.Email, &profile.Firstname, &profile.Lastname, &profile.DateOfBirth, &profile.Nickname, &profile.Avatar, &profile.About, &profile.IsPublic, &profile.CreatedAt)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	profile.Posts, err = postRepo.GetMyPosts(userId)
	if err != nil {
		return nil, err
	}
	profile.Followers, err = followRepo.GetFollowers(userId)
	if err != nil {
		return nil, err
	}
	profile.Following, err = followRepo.GetFollowing(userId)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) GetProfile(loggedId, targetId int64) (*Profile, error) {
	var profile Profile
	postRepo := NewPostRepository()
	followRepo := NewFollowRepository()
	query := `
		SELECT id, email, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(query, targetId).Scan(&profile.ID, &profile.Email, &profile.Firstname, &profile.Lastname, &profile.DateOfBirth, &profile.Nickname, &profile.Avatar, &profile.About, &profile.IsPublic, &profile.CreatedAt)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	profile.Posts, err = postRepo.GetProfilePosts(loggedId, targetId)
	if err != nil {
		return nil, err
	}
	profile.Followers, err = followRepo.GetFollowers(targetId)
	if err != nil {
		return nil, err
	}
	profile.Following, err = followRepo.GetFollowing(targetId)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) IsPublic(userId int64) (bool, error) {
	query := `SELECT COUNT(*) > 0 FROM users u WHERE u.id = $1 AND u.is_public = 1`
	var isPublic bool
	if err := r.db.QueryRow(query, userId).Scan(&isPublic); err != nil {
		return false, err
	}

	return isPublic, nil
}
