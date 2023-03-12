package utils

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	const cost = 12
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), cost)

	return string(hashedPassword)
}

func CheckPassword(hashedPassword, databasePassword []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, databasePassword)
	if err != nil {
		logrus.Error("failed to compare hash ", err)
	}

	return err
}
