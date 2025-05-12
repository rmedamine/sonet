package models

import (
	"database/sql"
	"fmt"
	"time"

	"social/config"
)

type Follow struct {
	ID            int64     `json:"id"`
	FollowerName  string    `json:"follower_name"`
	FollowingName string    `json:"following_name"`
	FollowerID    int64     `json:"follower_id"`
	FollowingID   int64     `json:"following_id"`
	Accepted      string    `json:"accepted"`
	CreatedAt     time.Time `json:"createdAt"`
}

type FollowRepository struct {
	db *sql.DB
}

func NewFollowRepository() *FollowRepository {
	return &FollowRepository{db: config.DB}
}

func (r *FollowRepository) Create(follow *Follow) error {
	UserRepo := NewUserRepository()
	followerName, _ := UserRepo.GetName(follow.FollowerID)
	followingName, _ := UserRepo.GetName(follow.FollowingID)
	query := `INSERT INTO follows (followerId, followingId, followerName, followingName, createdAt) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, follow.FollowerID, follow.FollowingID, followerName, followingName, follow.CreatedAt)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	follow.ID = id
	return nil
}

func (r *FollowRepository) GetFollowers(userId int64) ([]Follow, error) {
	query := `SELECT id, followerId, followingId, followingName, followerName, createdAt FROM follows WHERE followingId = ? AND accepted = 'accepted'`
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var followers []Follow
	for rows.Next() {
		var follower Follow
		err := rows.Scan(&follower.ID, &follower.FollowerID, &follower.FollowingID, &follower.FollowingName, &follower.FollowerName, &follower.CreatedAt)
		if err != nil {
			return nil, err
		}
		fmt.Println(follower.CreatedAt)
		followers = append(followers, follower)
	}
	return followers, nil
}

func (r *FollowRepository) GetFollowersFollowing(userId int64) ([]Follow, error) {
	query := `SELECT id, followerId, followingId, followingName, followerName, createdAt FROM follows WHERE followingId = ? OR followerId = ? AND accepted = 'accepted'`
	rows, err := r.db.Query(query, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var followers []Follow
	for rows.Next() {
		var follower Follow
		err := rows.Scan(&follower.ID, &follower.FollowerID, &follower.FollowingID, &follower.FollowingName, &follower.FollowerName, &follower.CreatedAt)
		if err != nil {
			return nil, err
		}
		fmt.Println(follower.CreatedAt)
		followers = append(followers, follower)
	}
	return followers, nil
}

func (r *FollowRepository) GetFollowing(userId int64) ([]Follow, error) {
	query := `SELECT id, followerId, followingId, createdAt FROM follows WHERE followerId = ? AND accepted = 'accepted'`
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var following []Follow
	for rows.Next() {
		var follow Follow
		err := rows.Scan(&follow.ID, &follow.FollowerID, &follow.FollowingID, &follow.CreatedAt)
		if err != nil {
			return nil, err
		}
		following = append(following, follow)
	}
	return following, nil
}

func (r *FollowRepository) GetFollowRequests(userId int64) ([]Follow, error) {
	query := `SELECT id, followerId, followingId, createdAt FROM follows WHERE followingId = ? AND accepted = 'pending'`
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var requests []Follow
	for rows.Next() {
		var request Follow
		err := rows.Scan(&request.ID, &request.FollowerID, &request.FollowingID, &request.CreatedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (r *FollowRepository) FollowRequestExists(followerId, followingId int64) (bool, error) {
	query := `SELECT id FROM follows WHERE followerId = ? AND followingId = ? AND accepted = 'pending'`
	row := r.db.QueryRow(query, followerId, followingId)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *FollowRepository) AcceptRequest(followerId, followingId int64) error {
	query := `UPDATE follows SET accepted = 'accepted' WHERE followerId = ? AND followingId = ?`
	_, err := r.db.Exec(query, followerId, followingId)
	return err
}

func (r *FollowRepository) GetFollowById(followId int64) (*Follow, error) {
	query := `SELECT id, followerId, followingId, createdAt FROM follows WHERE id = ?`
	row := r.db.QueryRow(query, followId)
	var follow Follow
	err := row.Scan(&follow.ID, &follow.FollowerID, &follow.FollowingID, &follow.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *FollowRepository) RejectRequest(followerId, followingId int64) error {
	query := `DELETE FROM follows WHERE followerId = ? AND followingId = ?`
	_, err := r.db.Exec(query, followerId, followingId)
	return err
}

func (r *FollowRepository) Unfollow(followerId, followingId int64) error {
	query := `DELETE FROM follows WHERE followerId = ? AND followingId = ?`
	_, err := r.db.Exec(query, followerId, followingId)
	return err
}

func (r *FollowRepository) IsFollowing(followerId, followingId int64) (bool, error) {
	query := `SELECT id FROM follows WHERE followerId = ? AND followingId = ?`
	row := r.db.QueryRow(query, followerId, followingId)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
