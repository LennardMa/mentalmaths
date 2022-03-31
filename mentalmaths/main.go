package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type InputQN struct {
	Id             string `json:"Id"`
	QuestionNumber int    `json:"questionNumber" binding:"required"`
	MaxNumber      int    `json:"maxNumber" binding:"required"`
}

type DatabaseQN struct {
	Id           string
	SliceOne     []int
	SliceTwo     []int
	SliceOP      []string
	SliceAnswers []int
	MaxNumber    int
	StartTime    time.Time
}
type InputAn struct {
	Answers []int `json:"answers" binding:"required"`
}

var GlobalDB []DatabaseQN

func main() {

	router := gin.Default()
	router.POST("/api", postQN)
	router.POST("/answers/:id", getScore)
	router.Run("localhost:8080")

}

func postQN(c *gin.Context) {
	var newQN InputQN
	var newDB DatabaseQN
	if err := c.ShouldBindJSON(&newQN); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rand.Seed(time.Now().UnixNano())
	newDB.SliceOne = randInt(newQN.MaxNumber, newQN.QuestionNumber)
	newDB.SliceTwo = randInt(newQN.MaxNumber, newQN.QuestionNumber)
	newDB.SliceOP = randOperator(newQN.QuestionNumber)
	id := uuid.New().String()
	newDB.Id = id
	c.JSON(http.StatusCreated, newDB)

	newDB.SliceAnswers = ansInt(newDB.SliceOne, newDB.SliceTwo, newDB.SliceOP)
	newDB.StartTime = time.Now()
	newDB.MaxNumber = newQN.MaxNumber
	GlobalDB = append(GlobalDB, newDB)

}

func getScore(c *gin.Context) {
	id := c.Param("id")
	var newIA InputAn
	var score int
	var comp float64
	if err := c.ShouldBindJSON(&newIA); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var game DatabaseQN
	var index int
	for i := range GlobalDB {
		if GlobalDB[i].Id == id {
			game = GlobalDB[i]
			index = i
			break
		}
	}
	for i := range game.SliceAnswers {
		if newIA.Answers[i] == game.SliceAnswers[i] {
			score++
		}
	}
	currentTime := time.Now()
	diff := currentTime.Sub(game.StartTime)
	comp = (float64(score)/diff.Seconds() + 1) * float64(game.MaxNumber) * (float64(score) / float64(len(game.SliceTwo)))
	comp = math.Round(comp*100) / 100
	c.JSON(http.StatusAccepted, comp)
	GlobalDB = append(GlobalDB[:index], GlobalDB[index+1:]...)

}

func randInt(max, n int) []int {
	resSlice := make([]int, n)
	for i := range resSlice {
		resSlice[i] = rand.Intn(max) + 1
	}
	return resSlice
}
func randOperator(n int) []string {
	opSlice := make([]string, n)
	var op = []string{"+", "-", "*"}
	for i := range opSlice {
		opSlice[i] = op[rand.Intn(3)]
	}
	return opSlice
}

func ansInt(a []int, b []int, c []string) []int {
	ans := make([]int, len(b))
	for i := range ans {
		if c[i] == "+" {
			ans[i] = a[i] + b[i]
		} else if c[i] == "-" {
			ans[i] = a[i] - b[i]
		} else {
			ans[i] = a[i] * b[i]
		}
	}
	return ans
}
