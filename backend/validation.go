package backend

import (
	"fmt"
	"regexp"
	"strings"
)

// validateUsername checks that the username is valid
func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return fmt.Errorf("username must contain at least 3 characters")
	}
	if len(username) > 30 {
		return fmt.Errorf("username cannot exceed 30 characters")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if !matched {
		return fmt.Errorf("username can only contain letters, numbers and underscores")
	}
	return nil
}

// validateFullName checks that the full name is valid
func validateFullName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return fmt.Errorf("full name must contain at least 2 characters")
	}
	if len(name) > 100 {
		return fmt.Errorf("full name cannot exceed 100 characters")
	}
	return nil
}

// validateEmail checks that the email address is valid
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) < 5 {
		return fmt.Errorf("email address is too short")
	}
	if len(email) > 254 {
		return fmt.Errorf("email address cannot exceed 254 characters")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	if !matched {
		return fmt.Errorf("email address is not valid")
	}
	return nil
}

// validatePassword checks that the password is strong enough
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must contain at least 8 characters")
	}
	if len(password) > 100 {
		return fmt.Errorf("password cannot exceed 100 characters")
	}
	return nil
}

// validatePostTitle checks that a post title is valid
func validatePostTitle(title string) error {
	title = strings.TrimSpace(title)
	if len(title) == 0 {
		return fmt.Errorf("title cannot be empty")
	}
	if len(title) > 200 {
		return fmt.Errorf("title cannot exceed 200 characters")
	}
	return nil
}

// validatePostContent checks that a post's content is valid
func validatePostContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil // content can be empty
	}
	if len(content) > 10000 {
		return fmt.Errorf("content cannot exceed 10,000 characters")
	}
	return nil
}

// validatePostTheme checks that a theme is valid
func validatePostTheme(theme string) error {
	if theme == "" {
		return nil // theme is optional
	}
	validThemes := map[string]bool{
		"tech":     true,
		"sport":    true,
		"gaming":   true,
		"culture":  true,
		"science":  true,
		"business": true,
		"others":   true,
	}
	if !validThemes[theme] {
		return fmt.Errorf("theme '%s' is not valid", theme)
	}
	return nil
}

// validateCommentContent checks that a comment is valid
func validateCommentContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return fmt.Errorf("comment cannot be empty")
	}
	if len(content) > 5000 {
		return fmt.Errorf("comment cannot exceed 5,000 characters")
	}
	return nil
}

// validatePositiveID checks that an ID is a positive integer
func validatePositiveID(id int) error {
	if id <= 0 {
		return fmt.Errorf("ID must be a positive number")
	}
	return nil
}