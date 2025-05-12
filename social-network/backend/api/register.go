package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"social/models"
	"social/services"
	"social/utils"
)

// RegisterHandler handles the user registration process
func RegisterApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var user models.User
	// Decode the request body into the user struct
	user.Email = strings.TrimSpace(r.FormValue("email"))
	user.Password = r.FormValue("password")
	user.Firstname = strings.TrimSpace(r.FormValue("firstname"))
	user.Lastname = strings.TrimSpace(r.FormValue("lastname"))
	user.DateOfBirth = r.FormValue("date_of_birth")
	user.Nickname = strings.TrimSpace(r.FormValue("nickname"))
	user.About = strings.TrimSpace(r.FormValue("about"))
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "File too large", nil)
		return
	}
	user.Avatar = r.FormValue("avatar")
	user.IsPublic = true
	// Validate input
	if err := validateUser(&user, false); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	file, handler, err := r.FormFile("avatar")
	if err != nil && file != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid file upload", nil)
		return
	}

	if file != nil {
		defer file.Close()
		AvatarPath, status := services.Upload(file, handler, "avatar")
		if status != http.StatusOK {
			utils.WriteJSON(w, status, AvatarPath, nil)
			return
		}
		user.Avatar = AvatarPath
	}

	// Save user to database
	err = services.RegisterUser(&user)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Error creating user", nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, "User created successfully", nil)
}

// validateUser validates the user input for registration
func validateUser(user *models.User, isUpdateProfile bool) error {
	// Validate email format
	minAge := 16
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(user.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password length
	if !isUpdateProfile && len(user.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	// Validate required fields
	if user.Firstname == "" || user.Lastname == "" {
		return fmt.Errorf("firstname and Lastname are required")
	}

	if user.DateOfBirth == "" {
		return fmt.Errorf("date of birth is required")
	}

	dob, err := time.Parse("2006-01-02", user.DateOfBirth)
	if err != nil {
		return fmt.Errorf("invalid date format. Use YYYY-MM-DD")
	}

	// Validate that DOB is not in the future
	if dob.After(time.Now()) {
		return fmt.Errorf("date of birth cannot be in the future")
	}

	// Calculate the user's age
	today := time.Now()
	age := today.Year() - dob.Year()
	if today.YearDay() < dob.YearDay() {
		age-- // Adjust if birthday hasn't occurred yet this year
	}

	// Check if the user is at least 16 years old
	if age < minAge {
		return fmt.Errorf("user must be at least 16 years old")
	}
	return nil
}
