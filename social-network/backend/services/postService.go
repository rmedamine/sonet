package services

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social/models"
	"social/utils"
)

var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

func CreateNewPost(post *models.Post, file multipart.File, fileHeader *multipart.FileHeader) error {
	postRepo := models.NewPostRepository()

	// Validate title length
	if !utils.IsBetween(post.Title, 0, 200) {
		return errors.New("title has exceeded the limits")
	}

	// Validate content length
	if !utils.IsBetween(post.Content, 0, 3000) {
		return errors.New("content has exceeded the limits")
	}

	// Check if content is empty
	if strings.TrimSpace(post.Content) == "" {
		return errFieldsEmpty
	}


	if post.Privacy < 0 || post.Privacy > 2 {
		return errInvalidPrivacy
	}

	// Handle file upload if an image is provided
	if file != nil && fileHeader != nil {
		// Validate file extension
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !AllowedExtensions[ext] {
			return errors.New("unsupported file type, only images are allowed")
		}

		// Validate MIME type
		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		if err != nil {
			return errors.New("error reading file for validation")
		}
		file.Seek(0, io.SeekStart) // Reset file pointer after reading

		mimeType := http.DetectContentType(buffer)
		if !strings.HasPrefix(mimeType, "image/") {
			return errors.New("invalid file type, only images are allowed")
		}

		// Generate unique filename
		uniqueFileName := utils.GenerateRandomFilename(ext)
		imagePath := filepath.Join("uploads/posts", uniqueFileName)

		// Save file to server
		if err := os.MkdirAll("uploads/posts", os.ModePerm); err != nil {
			return errors.New("error creating the image directory")
		}
		outFile, err := os.Create(imagePath)
		if err != nil {
			return errors.New("error saving the image")
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			return errors.New("error writing the image file")
		}

		// Set image path in the post struct
		post.Image = imagePath
	}

	// Set creation time
	post.CreatedAt = time.Now()

	// Save post to the database
	err := postRepo.Create(post)
	if err != nil {
		return err
	}
	userRepo := models.NewUserRepository()

	for _, user := range post.Users {
		exist, err := userRepo.UserExistsById(user)
		if err != nil {
			return err
		}
		if !exist {
			return errors.New("invalid user")
		}

		if err = postRepo.AddPrivacyUser(post.ID, user); err != nil {
			return err
		}
	}

	return nil
}


