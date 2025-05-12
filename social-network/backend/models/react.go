package models

import (
	"database/sql"
	"errors"

	"social/config"
)

type ReactionPost struct {
	Id           int64  `json:"id"`
	UserId       int64  `json:"userId"`
	PostId       int64  `json:"postId"`
	ReactionType string `json:"reaction_type"`
}

type ReactionComment struct {
	Id           int64  `json:"id"`
	UserId       int64  `json:"userId"`
	CommentId    int64  `json:"commentId"`
	ReactionType string `json:"reaction_type"`
}

type ReactionRepository struct {
	db *sql.DB
}

func NewReactionRepository() *ReactionRepository {
	return &ReactionRepository{db: config.DB}
}

func (r *ReactionRepository) ReactToPost(userId, postId int64, reaction string) error {
	isExist, err := r.IsPostReactExist(userId, postId)
	if err != nil {
		return err
	}
	if reaction == "" {
		if !isExist {
			return errors.New("you can't remove a reaction that does not exist")
		}
		query := `DELETE FROM post_reactions WHERE userId = ? AND postId = ?`
		_, err = r.db.Exec(query, userId, postId)
		return err
	}
	if isExist {
		return errors.New("you can't add a reaction that already exists")
	}
	query := `
		INSERT INTO post_reactions (userId, postId, reaction_type)
		VALUES (?, ?, ?)
	`
	_, err = r.db.Exec(query, userId, postId, reaction)
	return err
}

func (r *ReactionRepository) ReactToComment(userId, commentId int64, reaction string) error {
	isExist, err := r.IsReactCommentExist(userId, commentId)
	if err != nil {
		return err
	}
	if reaction == "" {
		if !isExist {
			return errors.New("you can't remove a reaction that does not exist")
		}
		// Remove reaction if no reaction type is provided
		query := `DELETE FROM comment_reactions WHERE userId = ? AND commentId = ?`
		_, err = r.db.Exec(query, userId, commentId)
		return err
	}
	if isExist {
		return errors.New("you can't add a reaction that already exists")
	}
	query := `
		INSERT INTO comment_reactions (userId, commentId, reaction_type)
		VALUES (?, ?, ?)
	`
	_, err = r.db.Exec(query, userId, commentId, reaction)
	return err
}

func (r *ReactionRepository) IsPostReactExist(userId, postId int64) (bool, error) {
	var exist bool
	query := `
		SELECT EXISTS(SELECT * FROM post_reactions WHERE userId = ? AND postId = ?)
	`
	if err := r.db.QueryRow(query, userId, postId).Scan(&exist); err != nil {
		return false, err
	}
	return exist, nil
}

func IsGroupPostReactExist(userId, postId int64, db *sql.DB) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_post_reactions WHERE userId = ? AND groupPostId = ?`
	err := db.QueryRow(query, userId, postId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (r *ReactionRepository) IsReactCommentExist(userId, commentId int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM comment_reactions WHERE userId = ? AND commentId = ?`
	err := r.db.QueryRow(query, userId, commentId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
