package models

import (
	"database/sql"
	"errors"
	"time"

	"social/config"
)

// GroupPostReaction represents a reaction to a group post
type GroupPostReaction struct {
	Id           int64  `json:"id"`
	UserId       int64  `json:"userId"`
	GroupPostId  int64  `json:"groupPostId"`
	ReactionType string `json:"reaction_type"`
}

// GroupComment represents a comment on a group post
type GroupComment struct {
	ID          int64     `json:"id"`
	GroupPostId int64     `json:"groupPostId"`
	UserId      int64     `json:"userId"`
	Name        string    `json:"name"`
	Avatar      string    `json:"avatar"`
	Likes       int       `json:"likes"`
	Comment     string    `json:"comment"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"createdAt"`
	IsUserLiked int       `json:"isUserLiked"`
}

// GroupCommentReaction represents a reaction to a group comment
type GroupCommentReaction struct {
	Id           int64  `json:"id"`
	UserId       int64  `json:"userId"`
	CommentId    int64  `json:"commentId"`
	ReactionType string `json:"reaction_type"`
}

// GroupReactionRepository handles reactions for group posts and comments
type GroupReactionRepository struct {
	db *sql.DB
}

// NewGroupReactionRepository creates a new GroupReactionRepository
func NewGroupReactionRepository() *GroupReactionRepository {
	return &GroupReactionRepository{db: config.DB}
}

// ReactToGroupPost handles reactions to group posts
func (r *GroupReactionRepository) ReactToGroupPost(userId, groupPostId int64, reaction string) error {
	isExist, err := r.IsGroupPostReactionExist(userId, groupPostId)
	if err != nil {
		return err
	}

	if reaction == "" {
		if !isExist {
			return errors.New("you can't remove a reaction that does not exist")
		}
		query := `DELETE FROM group_post_reactions WHERE userId = ? AND groupPostId = ?`
		_, err = r.db.Exec(query, userId, groupPostId)
		return err
	}

	if isExist {
		return errors.New("you can't add a reaction that already exists")
	}

	query := `
		INSERT INTO group_post_reactions (userId, groupPostId, reaction_type)
		VALUES (?, ?, ?)
	`
	_, err = r.db.Exec(query, userId, groupPostId, reaction)
	return err
}

// ReactToGroupComment handles reactions to group comments
func (r *GroupReactionRepository) ReactToGroupComment(userId, commentId int64, reaction string) error {
	isExist, err := r.IsGroupCommentReactionExist(userId, commentId, r.db)
	if err != nil {
		return err
	}

	if reaction == "" {
		if !isExist {
			return errors.New("you can't remove a reaction that does not exist")
		}
		query := `DELETE FROM group_comment_reactions WHERE userId = ? AND commentId = ?`
		_, err = r.db.Exec(query, userId, commentId)
		return err
	}

	if isExist {
		return errors.New("you can't add a reaction that already exists")
	}

	query := `
		INSERT INTO group_comment_reactions (userId, commentId, reaction_type)
		VALUES (?, ?, ?)
	`
	_, err = r.db.Exec(query, userId, commentId, reaction)
	return err
}

// IsGroupPostReactionExist checks if a reaction to a group post exists
func (r *GroupReactionRepository) IsGroupPostReactionExist(userId, groupPostId int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_post_reactions WHERE userId = ? AND groupPostId = ?`
	err := r.db.QueryRow(query, userId, groupPostId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// IsGroupCommentReactionExist checks if a reaction to a group comment exists
func (r *GroupReactionRepository) IsGroupCommentReactionExist(userId, commentId int64, db *sql.DB) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_comment_reactions WHERE userId = ? AND commentId = ?`
	err := db.QueryRow(query, userId, commentId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// GroupCommentRepository handles database operations for group comments
type GroupCommentRepository struct {
	db *sql.DB
}

// NewGroupCommentRepository creates a new GroupCommentRepository
func NewGroupCommentRepository() *GroupCommentRepository {
	return &GroupCommentRepository{db: config.DB}
}

// Create creates a new comment on a group post
func (r *GroupCommentRepository) Create(comment *GroupComment) error {
	query := `INSERT INTO group_comments (groupPostId, userId, comment, image) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, comment.GroupPostId, comment.UserId, comment.Comment, comment.Image)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	comment.ID = id
	return nil
}

// IsCommentExist checks if a group comment exists
func (r *GroupCommentRepository) IsCommentExist(commentId int64) (bool, error) {
	query := "SELECT COUNT(id) FROM group_comments WHERE ID = ?"
	var count int64
	err := r.db.QueryRow(query, commentId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// GetGroupPostComments retrieves comments for a specific group post
func (r *GroupCommentRepository) GetGroupPostComments(groupPostId, userId int64) ([]GroupComment, error) {
	query := `SELECT 
		c.id, c.groupPostId, c.userId, c.comment, c.image, c.createdAt, 
		COALESCE(u.nickname, u.firstname || ' ' || u.lastname) AS username,
		u.avatar,
		(SELECT COUNT(*) FROM group_comment_reactions WHERE reaction_type='LIKE' AND commentId=c.id) AS likes
	FROM group_comments c
	LEFT JOIN users u ON c.userId = u.id
	WHERE c.groupPostId = ?
	ORDER BY c.createdAt DESC`

	var comments []GroupComment
	rows, err := r.db.Query(query, groupPostId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment GroupComment
		err := rows.Scan(&comment.ID, &comment.GroupPostId, &comment.UserId, &comment.Comment, &comment.Image,
			&comment.CreatedAt, &comment.Name, &comment.Avatar, &comment.Likes)
		if err != nil {
			if err == sql.ErrNoRows {
				return comments, nil
			}
			return nil, err
		}
		comment.IsUserLiked = 0
		reactRepo := NewGroupReactionRepository()
		isLiked, err := reactRepo.IsGroupCommentReactionExist(userId, comment.ID, r.db)
		if err != nil {
			return nil, err
		}
		if isLiked {
			comment.IsUserLiked = 1
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// DeleteReaction removes a reaction from a group comment
func (r *GroupCommentRepository) DeleteReaction(userId int64, commentId int64) error {
	query := "DELETE FROM group_comment_reactions WHERE userId = ? AND commentId = ?"
	_, err := r.db.Exec(query, userId, commentId)
	if err != nil {
		return err
	}
	return nil
}

// IsReactionExist checks if a specific reaction exists on a group comment
func (r *GroupCommentRepository) IsReactionExist(userId int64, commentId int64, reactionType string) (bool, error) {
	query := "SELECT COUNT(*) FROM group_comment_reactions WHERE userId = ? AND commentId = ? AND reaction_type = ?"
	var count int
	err := r.db.QueryRow(query, userId, commentId, reactionType).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// GetCommentReaction gets the reaction counts for a group comment
func (r *GroupCommentRepository) GetCommentReaction(commentId int64) (int, error) {
	query := `SELECT COUNT(*)
	FROM group_comment_reactions WHERE commentId = ? AND reaction_type = 'LIKE'`
	var likes int
	err := r.db.QueryRow(query, commentId).Scan(&likes)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return likes, nil
}
