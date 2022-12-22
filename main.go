package main

import (
	routes "gojwt/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error in .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.New()
	router.Use(gin.Logger())
	routes.AuthRouter(router)
	routes.UserRouter(router)
	router.GET("/example", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "permission granted"})
	})
	router.Run(":" + port)
}
