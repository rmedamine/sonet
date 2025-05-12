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

// CreateNewGroupPost handles the creation of a new post in a group
func CreateNewGroupPost(post *models.PostGroup, file multipart.File, fileHeader *multipart.FileHeader) error {
	postGroupRepo := models.NewPostGroupRepository()

	// Validate content length
	if !utils.IsBetween(post.Content, 0, 3000) {
		return errors.New("content has exceeded the limits")
	}

	// Check if content is empty
	if strings.TrimSpace(post.Content) == "" {
		return errFieldsEmpty
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
		imagePath := filepath.Join("uploads/group_posts", uniqueFileName)

		// Save file to server
		if err := os.MkdirAll("uploads/group_posts", os.ModePerm); err != nil {
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
	err := postGroupRepo.Create(post)
	if err != nil {
		return err
	}

	return nil
}

// GetGroupPosts retrieves posts for a specific group with pagination
func GetGroupPosts(groupId int64, page, limit int, userId int64) ([]*models.PostGroup, int, error) {
	postGroupRepo := models.NewPostGroupRepository()

	// Check if page is valid
	if page < 1 {
		page = 1
	}

	// Get posts with pagination
	posts, err := postGroupRepo.GetPostsPerPage(groupId, page, limit, userId)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	totalCount, err := postGroupRepo.Count(groupId)
	if err != nil {
		return nil, 0, err
	}

	return posts, totalCount, nil
}

// GetGroupPostById retrieves a single post by ID
func GetGroupPostById(postId int64) (*models.PostGroup, error) {
	postGroupRepo := models.NewPostGroupRepository()

	// Check if post exists
	exists, err := postGroupRepo.IsPostExist(postId)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("post not found")
	}

	// Get the post
	return postGroupRepo.GetPostById(postId)
}
