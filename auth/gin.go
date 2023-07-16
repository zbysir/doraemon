package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
)

var AuthErr = errors.New("auth error")

func Auth(secret string) gin.HandlerFunc {
	if secret == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		t, _ := c.Cookie("token")
		if t == "" {
			c.Error(AuthErr)
			c.Abort()
			return
		}
		if !CheckToken(secret, t) {
			c.Error(AuthErr)
			c.Abort()
			return
		}

		c.Next()
	}
}

func CreateAndSaveToken(c *gin.Context, secret string) {
	t := CreateToken(secret)
	c.SetCookie("token", t, 7*24*3600, "", c.Request.Host, false, true)
}
