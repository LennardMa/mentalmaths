//ToDO: api rate limiter
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/robfig/cron"
	"golang.org/x/time/rate"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type InputQN struct {
	Id             string `json:"Id"`
	QuestionNumber int    `json:"questionNumber" validate:"required,min=1,max=200"`
	MaxNumber      int    `json:"maxNumber" validate:"required,min=2,max=9999"`
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
	Answers []int `json:"answers" validate:"required,min=1,max=200"`
}
type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var GlobalDB []DatabaseQN
var validate *validator.Validate

func main() {
	validate = validator.New()
	c := cron.New()
	c.AddFunc("@every 10s", delete)
	c.Start()

	//	rate := limiter.Rate{
	//		Period: 1 * time.Second,
	//		Limit:  1,
	//	}
	//	store := memory.NewStore()
	//	instance := limiter.New(store, rate)
	//	middleware := .NewMiddleware(limiter.New(store, rate))

	router := gin.Default()
	router.Use(limit())
	router.POST("/api", getQuestions)
	router.POST("/answers/:id", getScore)
	router.Run("localhost:8080")
}

func getQuestions(c *gin.Context) {
	var newQN InputQN
	var newDB DatabaseQN

	if err := c.ShouldBindJSON(&newQN); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	errs := validate.Struct(newQN)
	if errs != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.Error()})
		return
	}

	rand.Seed(time.Now().UnixNano())

	newDB.SliceOne = randInt(newQN.MaxNumber, newQN.QuestionNumber)
	newDB.SliceTwo = randInt(newQN.MaxNumber, newQN.QuestionNumber)
	newDB.SliceOP = randOperator(newQN.QuestionNumber)

	id := uuid.New().String()
	newDB.Id = id

	c.JSON(http.StatusCreated, newDB)

	newDB.StartTime = time.Now()
	newDB.SliceAnswers = ansInt(newDB.SliceOne, newDB.SliceTwo, newDB.SliceOP)
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
	errs := validate.Struct(newIA)
	if errs != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.Error()})
		return
	}

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
	for i := range game.SliceAnswers {
		if newIA.Answers[i] == game.SliceAnswers[i] {
			score++
		}
	}

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

func delete() {
	for i := range GlobalDB {
		currentTime := time.Now()
		diff := currentTime.Sub(GlobalDB[i].StartTime)
		if diff.Seconds() > 300 {
			GlobalDB = append(GlobalDB[:i], GlobalDB[i+1:]...)
		}
	}
}

var limiter = rate.NewLimiter(1, 3)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP
		limiter := getVisitor(ip)
		if limiter.Allow() == false {
			c.JSON(http.StatusTooManyRequests, "")
			return
		}

	}

}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 3)
		visitors[ip] = limiter
	}

	return limiter
}

//func Cleaner() chan bool {
//	ticker := time.NewTicker(2 * time.Minute)
//	quit := make(chan bool, 1)
//	go func() {
//		for {
//			select {
//			case <-ticker.C:
//				delete()
//			case <-quit:
//				ticker.Stop()
//			}
//		}
//	}()
//	return quit
//}
