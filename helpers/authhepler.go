package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func Checkusertype(c *gin.Context, role string) (err error) {
	user_type := c.GetString("user_type")
	err = nil
	if user_type != role {
		err = errors.New("unauthorized access")
		return err
	}
	return err
}
func Matchusertypeid(c *gin.Context, userid string) (err error) {
	usertype := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil
	if usertype == "USER" && userid != uid {
		err = errors.New("unauthorized access")
		return err
	}
	err = Checkusertype(c, usertype)
	return err
}
