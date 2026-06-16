package backend

import (
	"fmt"
	"regexp"
	"strings"
)

// validateUsername vérifie que le nom d'utilisateur est valide
func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return fmt.Errorf("le nom d'utilisateur doit contenir au moins 3 caractères")
	}
	if len(username) > 30 {
		return fmt.Errorf("le nom d'utilisateur ne peut pas dépasser 30 caractères")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if !matched {
		return fmt.Errorf("le nom d'utilisateur ne peut contenir que des lettres, chiffres et underscores")
	}
	return nil
}

// validateFullName vérifie que le nom complet est valide
func validateFullName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return fmt.Errorf("le nom complet doit contenir au moins 2 caractères")
	}
	if len(name) > 100 {
		return fmt.Errorf("le nom complet ne peut pas dépasser 100 caractères")
	}
	return nil
}

// validateEmail vérifie que l'adresse email est valide
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) < 5 {
		return fmt.Errorf("l'adresse email est trop courte")
	}
	if len(email) > 254 {
		return fmt.Errorf("l'adresse email ne peut pas dépasser 254 caractères")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	if !matched {
		return fmt.Errorf("l'adresse email n'est pas valide")
	}
	return nil
}

// validatePassword vérifie que le mot de passe est assez fort
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("le mot de passe doit contenir au moins 8 caractères")
	}
	if len(password) > 100 {
		return fmt.Errorf("le mot de passe ne peut pas dépasser 100 caractères")
	}
	return nil
}

// validatePostTitle vérifie qu'un titre de post est valide
func validatePostTitle(title string) error {
	title = strings.TrimSpace(title)
	if len(title) == 0 {
		return fmt.Errorf("le titre ne peut pas être vide")
	}
	if len(title) > 200 {
		return fmt.Errorf("le titre ne peut pas dépasser 200 caractères")
	}
	return nil
}

// validatePostContent vérifie qu'un contenu de post est valide
func validatePostContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil // le contenu peut être vide
	}
	if len(content) > 10000 {
		return fmt.Errorf("le contenu ne peut pas dépasser 10 000 caractères")
	}
	return nil
}

// validatePostTheme vérifie qu'un thème est valide
func validatePostTheme(theme string) error {
	if theme == "" {
		return nil // le thème est optionnel
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
		return fmt.Errorf("le thème '%s' n'est pas valide", theme)
	}
	return nil
}

// validateCommentContent vérifie qu'un commentaire est valide
func validateCommentContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return fmt.Errorf("le commentaire ne peut pas être vide")
	}
	if len(content) > 5000 {
		return fmt.Errorf("le commentaire ne peut pas dépasser 5 000 caractères")
	}
	return nil
}

// validatePositiveID vérifie qu'un ID est un entier positif
func validatePositiveID(id int) error {
	if id <= 0 {
		return fmt.Errorf("l'identifiant doit être un nombre positif")
	}
	return nil
}