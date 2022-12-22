package middleware

import (
	"net/http"

	helper "gojwt/helpers"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		clientToken := c.Request.Header.Get("token")
		if clientToken == " " {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "No authorization"})
			c.Abort()
			return
		}
		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.User_id)
		c.Set("user_type", claims.User_type)
		c.Next()

	}
}
