package utils

import (

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
