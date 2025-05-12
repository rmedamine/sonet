package models

import (
	"database/sql"
	"fmt"
	"log"

	"social/config"
)

type User struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	DateOfBirth string `json:"date_of_birth"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	About       string `json:"about"`
	IsPublic    bool   `json:"is_public"`
	CreatedAt   string `json:"created_at"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: config.DB}
}

func (r *UserRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (email, password, firstname, lastname, date_of_birth, nickname, avatar, about, is_public)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, user.Email, user.Password, user.Firstname, user.Lastname, user.DateOfBirth, user.Nickname, user.Avatar, user.About, user.IsPublic)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return fmt.Errorf("Error creating user: %v", err)
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
		WHERE email = ?
	`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Firstname, &user.Lastname, &user.DateOfBirth, &user.Nickname, &user.Avatar, &user.About, &user.IsPublic, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		log.Printf("Error querying database: %v", err)
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id int64) (*User, error) {
	var user User
	query := `
		SELECT id, email, password, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Password, &user.Firstname, &user.Lastname, &user.DateOfBirth, &user.Nickname, &user.Avatar, &user.About, &user.IsPublic, &user.CreatedAt)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		return nil, fmt.Errorf("Error getting user by ID: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET email = ?, firstname = ?, lastname = ?, date_of_birth = ?, nickname = ?, about = ?, is_public = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, user.Email, user.Firstname, user.Lastname, user.DateOfBirth, user.Nickname, user.About, user.IsPublic, user.ID)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return fmt.Errorf("Error updating user: %v", err)
	}
	return nil
}

func (r *UserRepository) UpdateAvatar(id int64, filename string) error {
	query := `
		UPDATE users
		SET avatar = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, filename, id)
	if err != nil {
		log.Printf("Error updating avatar: %v", err)
		return fmt.Errorf("Error updating avatar: %v", err)
	}
	return nil
}

func (r *UserRepository) SearchUsers(query string) ([]User, error) {
	var users []User
	sqlQuery := `
		SELECT id, email, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
		WHERE email LIKE ? OR firstname LIKE ? OR lastname LIKE ? OR nickname LIKE ?
	`
	rows, err := r.db.Query(sqlQuery, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, fmt.Errorf("Error searching users: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Firstname, &user.Lastname, &user.DateOfBirth, &user.Nickname, &user.Avatar, &user.About, &user.IsPublic, &user.CreatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			return nil, fmt.Errorf("Error scanning user: %v", err)
		}
		users = append(users, user)
	}
	return users, nil
}

// func (r *UserRepository) UpdatePassword(id int, password string) error {
// 	query := `
// 		UPDATE users
// 		SET password = ?
// 		WHERE id = ?
// 	`
// 	_, err := r.db.Exec(query, password, id)
// 	if err != nil {
// 		log.Printf("Error updating password: %v", err)
// 		return fmt.Errorf("Error updating password: %v", err)
// 	}
// 	return nil
// }

// func (r *UserRepository) DeleteUser(id int) error {
// 	query := `
// 		DELETE FROM users
// 		WHERE id = ?
// 	`
// 	_, err := r.db.Exec(query, id)
// 	if err != nil {
// 		log.Printf("Error deleting user: %v", err)
// 		return fmt.Errorf("Error deleting user: %v", err)
// 	}
// 	return nil
// }

func (r *UserRepository) GetUsers() ([]User, error) {
	var users []User
	query := `
		SELECT id, email, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
	`
	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return nil, fmt.Errorf("Error getting users: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Firstname, &user.Lastname, &user.DateOfBirth, &user.Nickname, &user.Avatar, &user.About, &user.IsPublic, &user.CreatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			return nil, fmt.Errorf("Error scanning user: %v", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) UserExistsByEmail(email string) (bool, error) {
	var count int
	query := `
    SELECT COUNT(*) FROM users 
    WHERE email = ?
    `
	err := r.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrUserNotFound
		}
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) UserExistsById(id int64) (bool, error) {
	fmt.Println(id)
	var count int
	query := `
    SELECT COUNT(*) FROM users 
    WHERE id = ?
    `
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrUserNotFound
		}
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) GetName(id int64) (string, error) {
	var name string
	query := `
		SELECT COALESCE(NULLIF(nickname, ''), firstname || ' ' || lastname) AS name
		FROM users
		WHERE id = ?
	`
	err := r.db.QueryRow(query, id).Scan(&name)
	if err != nil {
		log.Printf("Error getting name: %v", err)
		return "", fmt.Errorf("Error getting name: %v", err)
	}
	return name, nil
}

func (r *UserRepository) GetAvatar(id int64) (string, error) {
	var avatar string
	query := `
		SELECT avatar
		FROM users
		WHERE id = ?
	`
	err := config.DB.QueryRow(query, id).Scan(&avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		log.Printf("Error getting avatar: %v", err)
		return "", fmt.Errorf("Error getting avatar: %v", err)
	}
	return avatar, nil
}

func (r *UserRepository) SearchUser(query string) ([]User, error) {
	var users []User
	sqlQuery := `
		SELECT id, email, firstname, lastname, date_of_birth, nickname, avatar, about, is_public, created_at
		FROM users
		WHERE email LIKE ? OR firstname LIKE ? OR lastname LIKE ? OR nickname LIKE ?
	`
	rows, err := r.db.Query(sqlQuery, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, fmt.Errorf("Error searching users: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Firstname, &user.Lastname, &user.DateOfBirth, &user.Nickname, &user.Avatar, &user.About, &user.IsPublic, &user.CreatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			return nil, fmt.Errorf("Error scanning user: %v", err)
		}
		users = append(users, user)
	}
	return users, nil
}
