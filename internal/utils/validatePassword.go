package utils

import (
	"errors"
	"regexp"
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("Password must be at least 8 characters long")
	}

	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasLetter || !hasDigit || !hasSpecial {
		return errors.New("Password must contain letters, numbers and at least 1 special character")
	}

	return nil
}
