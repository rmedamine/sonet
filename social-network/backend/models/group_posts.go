package models

import (
	"database/sql"
	"time"

	"social/config"
)

type PostGroup struct {
	ID          int64     `json:"id"`
	GroupId     int64     `json:"group_id"`
	UserId      int64     `json:"user_id"`
	Content     string    `json:"content"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`   // For joining with user name
	Avatar      string    `json:"avatar"` // For joining with user avatar
	Likes       int       `json:"likes"`
	IsUserLiked int       `json:"is_user_liked"`
}

type PostGroupRepository struct {
	db *sql.DB
}

func NewPostGroupRepository() *PostGroupRepository {
	return &PostGroupRepository{db: config.DB}
}

func (r *PostGroupRepository) Create(postGroup *PostGroup) error {
	query := `INSERT INTO group_posts (group_id, user_id, content, image, created_at) VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, postGroup.GroupId, postGroup.UserId, postGroup.Content, postGroup.Image, time.Now())
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	postGroup.ID = id
	return nil
}

func (r *PostGroupRepository) GetPostsPerPage(groupId int64, page int, limit int, userId int64) ([]*PostGroup, error) {
	offset := (page - 1) * limit
	query := `SELECT 
    gp.id,
    gp.group_id,
    gp.user_id,
    gp.content,
    gp.image,
    gp.created_at,
    COALESCE(NULLIF(u.nickname, ''), u.firstname || ' ' || u.lastname) AS username,
    u.avatar,
    (SELECT COUNT(*) FROM group_post_reactions WHERE groupPostId = gp.id) AS likes
FROM 
    group_posts gp
LEFT JOIN 
    users u ON gp.user_id = u.id
WHERE 
    gp.group_id = ?
ORDER BY 
    gp.created_at DESC
LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, groupId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostGroup
	for rows.Next() {
		var post PostGroup
		if err := rows.Scan(&post.ID, &post.GroupId, &post.UserId, &post.Content, &post.Image, &post.CreatedAt, &post.Name, &post.Avatar, &post.Likes); err != nil {
			if err == sql.ErrNoRows {
				return posts, nil
			}
			return nil, err
		}
		post.IsUserLiked = 0
		reactRepo := NewGroupReactionRepository()
		isLiked, err := reactRepo.IsGroupPostReactionExist(userId, post.ID)
		if err != nil {
			return nil, err
		}
		if isLiked {
			post.IsUserLiked = 1
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostGroupRepository) Count(groupId int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM group_posts WHERE group_id = ?`, groupId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (r *PostGroupRepository) IsPostExist(id int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_posts WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (r *PostGroupRepository) GetPostById(id int64) (*PostGroup, error) {
	query := `SELECT 
    gp.id, 
    gp.group_id, 
    gp.user_id, 
    gp.content, 
    gp.image, 
    gp.created_at, 
    COALESCE(NULLIF(u.nickname, ''), u.firstname || ' ' || u.lastname) AS username,
    u.avatar
FROM 
    group_posts gp 
JOIN 
    users u ON gp.user_id = u.id 
WHERE 
    gp.id = ?`

	var post PostGroup
	row := r.db.QueryRow(query, id)
	err := row.Scan(&post.ID, &post.GroupId, &post.UserId, &post.Content, &post.Image, &post.CreatedAt, &post.Name, &post.Avatar)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostGroupRepository) GetPostsByGroupId(groupId int64) ([]*PostGroup, error) {
	query := `SELECT 
    gp.id, 
    gp.group_id, 
    gp.user_id, 
    gp.content, 
    gp.image, 
    gp.created_at,
    COALESCE(NULLIF(u.nickname, ''), u.firstname || ' ' || u.lastname) AS username,
    u.avatar
FROM 
    group_posts gp
JOIN 
    users u ON gp.user_id = u.id
WHERE 
    gp.group_id = ?
ORDER BY 
    gp.created_at DESC`

	rows, err := r.db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostGroup
	for rows.Next() {
		var post PostGroup
		err := rows.Scan(&post.ID, &post.GroupId, &post.UserId, &post.Content, &post.Image, &post.CreatedAt, &post.Name, &post.Avatar)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostGroupRepository) GetPostsByUserId(userId int64) ([]*PostGroup, error) {
	query := `SELECT id, group_id, user_id, content, image, created_at FROM group_posts WHERE user_id = ?`

	rows, err := r.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostGroup
	for rows.Next() {
		var post PostGroup
		err := rows.Scan(&post.ID, &post.GroupId, &post.UserId, &post.Content, &post.Image, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostGroupRepository) DeletePost(id int64) error {
	query := `DELETE FROM group_posts WHERE id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
