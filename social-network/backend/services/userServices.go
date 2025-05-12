package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"social/config"
	"social/models"
	"social/utils"
)

var (
	errInvalidUserPass  = errors.New("invalid Email or password")
	errUserOrEmailExist = errors.New("email already used")
	errFieldsEmpty      = errors.New("all fields must be completed")
	errInvalidPrivacy   = errors.New("invalid post privacy")
)

func RegisterUser(user *models.User) error {
	userRepo := models.NewUserRepository()

	// check if the username or email alread yexist
	isUserExist, err := userRepo.UserExistsByEmail(user.Email)
	if err != nil {
		return err
	}
	if isUserExist {
		return errUserOrEmailExist
	}
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return userRepo.CreateUser(user)
}

func LoginUser(email, password string) (*models.User, error) {
	if len(strings.TrimSpace(email)) == 0 || len(strings.TrimSpace(password)) == 0 {
		return nil, errFieldsEmpty
	}
	userRepo := models.NewUserRepository()
	user, err := userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	err = utils.CheckPassword(user.Password, password)
	if err != nil {
		fmt.Println(err)
		return nil, errInvalidUserPass
	}
	return user, nil
}

func Upload(file multipart.File, handler *multipart.FileHeader, path string) (string, int) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if !AllowedExtensions[ext] {
		return "Unsupported file type (only JPG, PNG, GIF allowed)", http.StatusBadRequest
	}

	// Generate a random filename
	randomFilename := utils.GenerateRandomFilename(ext)
	uploadPath := filepath.Join("uploads/"+path, randomFilename)

	// Ensure the upload directory exists
	if err := os.MkdirAll("uploads/"+path, os.ModePerm); err != nil {
		return "Error creating upload directory", http.StatusInternalServerError
	}

	// Save the file
	if err := utils.SaveFile(file, uploadPath); err != nil {
		return "Error saving file", http.StatusInternalServerError
	}
	return uploadPath, http.StatusOK
}

func GetUserData(token string) (*models.Profile, error) {
	// Check if token is empty
	if token == "" {
		return nil, fmt.Errorf("invalid user session")
	}

	// Get session
	session, err := config.SESSION.GetSession(token)
	if err != nil {
		return nil, fmt.Errorf("invalid user session")
	}

	// Validate session
	if session == nil || session.UserId == 0 {
		return nil, fmt.Errorf("invalid user session")
	}

	profileRepo := models.NewProfileRepository()
	profile, err := profileRepo.GetMyProfile(session.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}
