package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	if err := r.RunTLS(":8080", "/etc/certs/tls.crt", "/etc/certs/tls.key"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Starting server https://localhost:8080/")
}
