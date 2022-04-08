package main

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InputQN struct {
	QuestionNumber int
	MaxNumber      int
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

var GlobalDB []DatabaseQN

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/start", indexPage)
	router.POST("/api", getQuestions)
	router.POST("/answers/:id", getScore)

	router.Run("localhost:8080")

}
func indexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}
func getQuestions(c *gin.Context) {
	var newQN InputQN
	var newDB DatabaseQN

	MaxNumber, err := strconv.Atoi(c.PostForm("MaxNumber"))
	QuestionNumber, err2 := strconv.Atoi(c.PostForm("QuestionNumber"))
	if err != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newQN.MaxNumber = MaxNumber
	newQN.QuestionNumber = QuestionNumber

	rand.Seed(time.Now().UnixNano())

	newDB.SliceOne = randInt(newQN.MaxNumber, newQN.QuestionNumber)
	newDB.SliceTwo = randInt(newQN.MaxNumber, newQN.QuestionNumber)
	newDB.SliceOP = randOperator(newQN.QuestionNumber)

	id := uuid.New().String()
	newDB.Id = id

	c.HTML(http.StatusOK, "gaming.html", gin.H{
		"SliceOne": newDB.SliceOne,
		"SliceTwo": newDB.SliceTwo,
		"SliceOP":  newDB.SliceOP,
		"ID":       id,
	})

	newDB.StartTime = time.Now()
	newDB.SliceAnswers = ansInt(newDB.SliceOne, newDB.SliceTwo, newDB.SliceOP)
	newDB.MaxNumber = newQN.MaxNumber
	GlobalDB = append(GlobalDB, newDB)

}

func getScore(c *gin.Context) {

	id := c.Param("id")
	var score int
	var comp float64

	currentTime := time.Now()

	var game DatabaseQN
	var index int

	for i := range GlobalDB {
		if GlobalDB[i].Id == id {
			game = GlobalDB[i]
			index = i
			break
		}
	}

	for i, value := range game.SliceAnswers {
		answer, err := strconv.Atoi(c.PostForm("answer" + strconv.Itoa(i)))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if answer == value {
			score++
		}
	}

	diff := currentTime.Sub(game.StartTime)
	comp = (float64(score)/diff.Seconds() + 1) * float64(game.MaxNumber) * (float64(score) / float64(len(game.SliceTwo)))
	comp = math.Round(comp*100) / 100

	c.HTML(http.StatusOK, "end.html", gin.H{
		"Score": comp,
	})
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
