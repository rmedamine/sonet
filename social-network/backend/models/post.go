package models

import (
	"database/sql"
	"time"

	"social/config"
)

const (
	ALL = iota
	MY_POST
	LIKED_POST
)

const (
	PUBLIC       = 0
	SEMI_PRIVATE = 1
	PRIVATE      = 2
)

type Post struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	UserID      int64     `json:"userId"`
	Avatar      string    `json:"avatar"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Privacy     int       `json:"privacy"`
	Users       []int64   `json:"users"`
	Likes       int       `json:"likes"`
	IsUserLiked int       `json:"isUserLiked"`
}

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository() *PostRepository {
	return &PostRepository{db: config.DB}
}

func (r *PostRepository) Create(post *Post) error {
	query := `INSERT INTO posts (userId, avatar, title, content, createdAt, image, privacy) VALUES (?,?, ?, ?,?,?,?)`

	result, err := r.db.Exec(query, post.UserID, post.Avatar, post.Title, post.Content, post.CreatedAt, post.Image, post.Privacy)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	post.ID = id
	return nil
}

func (r *PostRepository) GetPostPerPage(page int, limit int, userId int64) ([]*Post, error) {
	offset := (page - 1) * limit
	query := `SELECT 
    p.id,
    p.title,
	p.avatar,
    p.content,
    p.createdAt,
    p.image,
	p.userId,
    COALESCE(NULLIF(u.nickname, ''), u.firstname || ' ' || u.lastname) AS username,
    (SELECT COUNT(*) FROM post_reactions) AS likes
FROM 
    posts p
LEFT JOIN 
    users u ON p.userId = u.id
LEFT JOIN 
    post_reactions pr ON p.id = pr.postId
LEFT JOIN 
    post_privacy pp ON p.id = pp.postId
LEFT JOIN 
    follows f ON f.followingId = p.userId AND f.followerId = $1
WHERE (
    p.userId = $1 OR
    (
        (u.is_public = 0 AND followerId IS NOT NULL
        AND (
            p.privacy = 0 OR p.privacy = 1 OR (
                p.privacy = 2 AND $1 in ( SELECT userId FROM post_privacy WHERE postId = p.id)
            )
        ))
        OR
        (
            u.is_public = 1 AND (
                p.privacy = 0 OR (p.privacy = 1 AND followerId IS NOT NULL)OR (
                p.privacy = 2 AND $1 in ( SELECT userId FROM post_privacy WHERE postId = p.id)
                )
            )
        )
	)
)
GROUP BY 
    p.id, u.id
ORDER BY 
    p.createdAt DESC
LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(query, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	var posts []*Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Avatar, &post.Content, &post.CreatedAt, &post.Image, &post.UserID, &post.Name, &post.Likes); err != nil {
			if err == sql.ErrNoRows {
				return posts, nil
			}
			return nil, err
		}
		post.IsUserLiked = 0
		reactRepo := NewReactionRepository()
		IsLiked, err := reactRepo.IsPostReactExist(userId, post.ID)
		if err != nil {
			return nil, err
		}
		if IsLiked {
			post.IsUserLiked = 1
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (r *PostRepository) IsPostExist(id int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM posts WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (r *PostRepository) GetPostById(id int64) (*Post, error) {
	query := `SELECT 
    p.id, 
    p.title, 
    p.userId, 
    p.content, 
    p.image, 
    p.createdAt, 
    COALESCE(NULLIF(u.nickname, ''), u.firstname || ' ' || u.lastname) AS username
FROM 
    posts p 
JOIN 
    users u ON p.userId = u.id 
WHERE 
    p.id = ?`
	var post Post
	row := r.db.QueryRow(query, id)
	err := row.Scan(&post.ID, &post.Title, &post.UserID, &post.Content, &post.Image, &post.CreatedAt, &post.Name)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) GetMyPosts(userId int64) ([]*Post, error) {
	query := `
		SELECT id, title, avatar, content, image, privacy, createdAt
		FROM posts
		WHERE userId = ?`
	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Avatar, &post.Content, &post.Image, &post.Privacy, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostRepository) GetPosts(loggedId, targetId int64) ([]*Post, error) {
	query := `
		SELECT p.id, p.avatar, p.content, p.image, p.privacy, p.createdAt, (SELECT COUNT(*) FROM post_reactions WHERE userId = $1 AND postId = p.id) FROM posts p
		JOIN users u ON p.userId = u.id
		LEFT JOIN follows f ON (f.followerId = $1 AND f.followingId = $2)
		WHERE (
			$1 = $2 OR (
				(u.is_public = 0 AND followerId IS NOT NULL
				AND (
					p.privacy = 0 OR p.privacy = 1 OR (
						p.privacy = 2 AND $1 in ( SELECT userId FROM post_privacy WHERE postId = p.id)
					)
				))
				OR
				(
					u.is_public = 1 AND (
						p.privacy = 0 OR (p.privacy = 1 AND followerId IS NOT NULL)OR (
						p.privacy = 2 AND $1 in ( SELECT userId FROM post_privacy WHERE postId = p.id)
						)
					)
				)
			)
		);
		`
	rows, err := r.db.Query(query, loggedId, targetId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Avatar, &post.Content, &post.Image, &post.Privacy, &post.CreatedAt, &post.IsUserLiked)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostRepository) GetProfilePosts(loggedId, targetId int64) ([]*Post, error) {
	query := `
		SELECT p.id, p.avatar, p.content, p.image, p.privacy, p.createdAt, (SELECT COUNT(*) FROM post_reactions WHERE userId = $1 AND postId = p.id) FROM posts p
		JOIN users u ON p.userId = u.id
		LEFT JOIN follows f ON (f.followerId = $1 AND f.followingId = $2)
		WHERE (
			p.userId = $2 AND
			($1 = $2 OR (
				(u.is_public = 0 AND followerId IS NOT NULL
				AND (
					p.privacy = 0 OR p.privacy = 1 OR (
						p.privacy = 2 AND $1 in ( SELECT userId FROM post_privacy WHERE postId = p.id)
					)
				))
				OR
				(
					u.is_public = 1 AND (
						p.privacy = 0 OR (p.privacy = 1 AND followerId IS NOT NULL)OR (
						p.privacy = 2 AND $1 in ( SELECT userId FROM post_privacy WHERE postId = p.id)
						)
					)
				)
			))
		);
		`
	rows, err := r.db.Query(query, loggedId, targetId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Avatar, &post.Content, &post.Image, &post.Privacy, &post.CreatedAt, &post.IsUserLiked)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostRepository) AddPrivacyUser(postId int64, userId int64) error {
	query := `
		INSERT INTO post_privacy (postId, userId) VALUES (?, ?)
	`

	prep, err := r.db.Prepare(query)
	if err != nil {
		return err
	}

	defer prep.Close()
	if _, err := prep.Exec(postId, userId); err != nil {
		return err
	}

	return nil
}
