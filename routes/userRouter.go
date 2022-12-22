package routes

import (
	controller "gojwt/controllers"
	middleware "gojwt/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(incomingRouter *gin.Engine) {
	incomingRouter.Use(middleware.Authenticate())
	incomingRouter.GET("/users/:user_id", controller.GetUser())
	incomingRouter.GET("/users", controller.GetUsers())
}
