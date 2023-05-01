package helpers

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func ExtractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	jwtToken := strings.Split(header, " ")
	if len(jwtToken) != 2 {
		return "", false
	}

	return jwtToken[1], true
}

func ParseToken(jwtToken string) (*jwt.Token, bool) {
	token, _ := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, OK := token.Method.(*jwt.SigningMethodHMAC); !OK {
			return nil, errors.New("bad signed method received")
		}

		return []byte("supersaucysecret"), nil
	})

	return token, true
}

func OptionalTokenCheck(c *gin.Context) {
	jwtToken, isHeaderExist := ExtractBearerToken(c.GetHeader("Authorization"))
	if isHeaderExist {

		token, isTokenParsed := ParseToken(jwtToken)
		if !isTokenParsed {
			c.Next()
		} else if token != nil {
			claims, OK := token.Claims.(jwt.MapClaims)
			if !OK {
				c.Next()
			} else {
				claimedUID, OK := claims["id"].(string)
				if !OK {
					c.Next()
				} else {
					c.Set("uuid", claimedUID)
				}
			}
		}
	}

	c.Next()
}
