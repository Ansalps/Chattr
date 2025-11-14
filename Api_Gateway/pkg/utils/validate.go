package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) []string {

	var validationErrors []string

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			// Create user-friendly messages
			var msg string
			field := e.Field()
			switch e.Tag() {
			case "required":
				msg = field + " is required"
			case "email":
				msg = field + " must be a valid email address"
			case "min":
				msg = field + " must be at least " + e.Param() + " characters long"
			case "max":
				msg = field + " must be at most " + e.Param() + " characters long"
			case "eqfield":
				msg = field + " must be equal to " + e.Param()
			case "oneof":
				msg = fmt.Sprintf("%s must be one of [%s]", field, e.Param())
			case "username_valid":
				_, customMsg := IsValidUsername(e.Value().(string))
				msg = fmt.Sprintf("invalid username: %s", customMsg)
			case "password_strong":
				_, customMsg := IsValidPassword(e.Value().(string))
				msg = fmt.Sprintf("invalid password: %s", customMsg)
			default:
				msg = field + " is invalid"
			}
			validationErrors = append(validationErrors, msg)
		}
	} else {
		validationErrors = append(validationErrors, err.Error())
	}

	return validationErrors
}

func IsValidUsername(username string) (bool, string) {
	maxLength := 30
	validChars := `^[a-z0-9._]+$`

	if len(username) > maxLength {
		return false, "Username cannot exceed 30 characters"
	}
	if username == "" || !regexp.MustCompile(validChars).MatchString(username) {
		return false, "Username contains invalid characters, chracters allowed ['a-z' , '0-9' , '.' , '_' ]"
	}
	if username[0] == '.' || username[len(username)-1] == '.' {
		return false, "Username cannot start or end with a dot (.)"
	}
	if strings.Contains(username, "..") {
		return false, "Username cannot contain consecutive dots (..)"
	}
	return true, ""
}

func IsValidPassword(password string) (bool, string) {
	minLength := 8
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	if len(password) < minLength {
		return false, "Password must be at least 8 characters long"
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}

	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}

	if !hasDigit {
		return false, "Password must contain at least one digit"
	}

	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}
