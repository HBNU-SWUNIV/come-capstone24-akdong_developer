package controllers

import (
	"carte/models"
	"net/http"

	"github.com/gin-gonic/gin"
)


func CreateContainer(c *gin.Context) {
	err := models.CreateContainer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container created successfully"})
}

func BuildImage(c *gin.Context) {
	err := models.BuildImage("image.tar.gz")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image built successfully"})
}
