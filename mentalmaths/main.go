package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type inputQN struct {
	Id             string  `json:"Id"`
	QuestionNumber float64 `json:"questionNumber" binding:"required"`
	MaxNumber      float64 `json:"maxNumber" binding:"required"`
}

var qn []inputQN

func main() {

	router := gin.Default()
	router.POST("/api", postQN)
	router.GET("/question", getQuestions)
	router.Run("localhost:8080")

}

func postQN(c *gin.Context) {
	var newQN inputQN
	if err := c.ShouldBindJSON(&newQN); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.New().String()
	newQN.Id = id
	qn = append(qn, newQN)
	c.JSON(http.StatusCreated, id)
}

func getQuestions(c *gin.Context) {

}
