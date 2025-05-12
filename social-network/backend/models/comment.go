package models

import (
	"database/sql"
	"time"

	"social/config"
)

type CommentReaction struct {
	CommentId int64 `json:"commentId"`
	Likes     int64 `json:"likes"`
	Dislikes  int64 `json:"dislikes"`
}

type CommentLike struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"userId"`
	CommentId    int64  `json:"commentId"`
	ReactionType string `json:"reactionType"` // Changed to string to reflect the reaction type (LIKE/DISLIKE)
}

type Comment struct {
	ID          int64     `json:"id"`
	PostID      int64     `json:"postId"`
	UserID      int64     `json:"userId"`
	Name        string    `json:"name"`
	Likes       int       `json:"likes"`
	Comment     string    `json:"comment"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"createdAt"`
	IsUserLiked int       `json:"isUserLiked"`
}

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{db: config.DB}
}

func (r *CommentRepository) Create(comment *Comment) error {
	query := `INSERT INTO comments (postId, userId, comment, image) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, comment.PostID, comment.UserID, comment.Comment, comment.Image)
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

func (r *CommentRepository) IsCommentExist(commentId int64) (bool, error) {
	query := "SELECT COUNT(id) FROM comments WHERE ID = ?"
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

func (r *CommentRepository) GetPostComments(postID, userId int64) ([]Comment, error) {
	query := `SELECT 
		c.id, c.postId, c.userId, c.comment, c.image, c.createdAt, 
		COALESCE(u.nickname, u.firstname || ' ' || u.lastname) AS username, 
		(SELECT COUNT(*) FROM comment_reactions WHERE reaction_type='LIKE' AND commentId=c.id) AS likes
	FROM comments c
	LEFT JOIN users u ON c.userId = u.id
	WHERE c.postId = ?
	ORDER BY c.createdAt DESC`

	var comments []Comment
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Comment, &comment.Image, &comment.CreatedAt,
			&comment.Name, &comment.Likes)
		if err != nil {
			if err == sql.ErrNoRows {
				return comments, nil
			}
			return nil, err
		}
		comment.IsUserLiked = 0
		reactRepo := NewReactionRepository()
		IsLiked, err := reactRepo.IsReactCommentExist(userId, comment.ID)
		if err != nil {
			return nil, err
		}
		if IsLiked {
			comment.IsUserLiked = 1
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *CommentRepository) DeleteReaction(userId int64, commentId int64) error {
	query := "DELETE FROM comment_reactions WHERE userId = ? AND commentId = ?"
	_, err := r.db.Exec(query, userId, commentId)
	if err != nil {
		return err
	}
	return nil
}

func (r *CommentRepository) IsReactionExist(userId int64, commentId int64, reactionType string) (bool, error) {
	query := "SELECT COUNT(*) FROM comment_reactions WHERE userId = ? AND commentId = ? AND reaction_type = ?"
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

func (r *CommentRepository) ReactComment(like CommentLike) error {
	query := `
        INSERT INTO comment_reactions (userId, commentId, reaction_type)
        VALUES (?, ?, ?)`

	_, err := r.db.Exec(query, like.UserID, like.CommentId, like.ReactionType, like.ReactionType)
	return err
}

func (r *CommentRepository) GetCommentReaction(commentId int64) (*CommentReaction, error) {
	query := `SELECT COUNT(*)
	FROM comment_reactions WHERE commentId = ? GROUP BY commentId`
	var commentReaction CommentReaction
	commentReaction.CommentId = commentId
	err := r.db.QueryRow(query, commentId).Scan(&commentReaction.Likes, &commentReaction.Dislikes)
	if err != nil {
		if err == sql.ErrNoRows {
			return &commentReaction, nil
		}
		return nil, err
	}
	return &commentReaction, nil
}
