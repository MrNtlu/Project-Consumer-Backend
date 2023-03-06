package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var ErrUnauthorized = "Unauthorized access."

func bindJSONData(data interface{}, c *gin.Context) bool {
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return true
	}

	return false
}

func validatorErrorHandler(err error) string {
	var validator validator.ValidationErrors
	if errors.As(err, &validator) {
		for _, fieldError := range validator {
			switch fieldError.Tag() {
			case "required":
				return fmt.Sprintf("Missing field! %s field is required.", fieldError.Field())
			case "email":
				return "Invalid email."
			case "oneof":
				return fmt.Sprintf("Constraint validation failed on %s field.", fieldError.Field())
			}
		}
	}

	return err.Error()
}
