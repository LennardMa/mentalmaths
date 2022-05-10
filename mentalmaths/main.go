//ToDO: password input validation
//ToDo: migrate to PostgreSQL
//ToDo: debug middleware
//ToDo: split up code into packages for better readability
//TODO: Session token
//TODO: improve error handling

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/robfig/cron"
	"golang.org/x/crypto/bcrypt"
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
type Credentials struct {
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required"`
}

type CredDB struct {
	Password  []byte
	Username  string
	Highscore float64
}

type session struct {
	username string
}

var sessions = map[string]session{}

var CredentialDB []CredDB
var GlobalDB []DatabaseQN

var validate *validator.Validate

func main() {
	validate = validator.New()

	c := cron.New()
	c.AddFunc("@every 10s", delete)
	c.Start()

	router := gin.Default()
	router.Use(limit())
	router.POST("/api", getQuestions)
	router.POST("/answers/:id", getScore)
	router.POST("/signin", Signin)
	router.POST("/signup", Signup)
	router.GET("/welcome", welcome)
	router.Run("localhost:8080")

	//	rate := limiter.Rate{
	//		Period: 1 * time.Second,
	//		Limit:  1,
	//	}
	//	store := memory.NewStore()
	//	instance := limiter.New(store, rate)
	//	middleware := .NewMiddleware(limiter.New(store, rate))

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

func Signup(c *gin.Context) {
	var cred Credentials
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var dummy CredDB
	dummy.Username = cred.Username
	dummy.Password = hashedPassword
	CredentialDB = append(CredentialDB, dummy)
	c.JSON(http.StatusAccepted, nil)
}

func Signin(c *gin.Context) {
	var cred Credentials
	var hscore float64
	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var hashedPassword []byte
	for i := range CredentialDB {
		if CredentialDB[i].Username == cred.Username {
			hashedPassword = CredentialDB[i].Password
			hscore = CredentialDB[i].Highscore
			break
		}
	}
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(cred.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Your Password was wrong dipshit": err.Error()})
		return
	}
	sessionToken := uuid.NewString()

	sessions[sessionToken] = session{
		username: cred.Username,
	}
	c.SetCookie("session token", sessionToken, 99999, "/", "localhost", true, true)
	c.JSON(http.StatusAccepted, hscore)
}

func welcome(c *gin.Context) {
	var hscore float64
	cookie, err := c.Cookie("session token")
	if err != nil {
		if err == http.ErrNoCookie {
			c.JSON(http.StatusUnauthorized, nil)
			return
		}
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	sessionToken := cookie
	userSession, exists := sessions[sessionToken]
	if !exists {
		c.JSON(http.StatusBadRequest, nil)
	}
	for i := range CredentialDB {
		if CredentialDB[i].Username == userSession.username {
			hscore = CredentialDB[i].Highscore
			break
		}
	}
	c.JSON(http.StatusAccepted, hscore)
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
		ip := c.ClientIP()
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
