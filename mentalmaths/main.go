//ToDO: password input validation
//ToDo: migrate to PostgreSQL and redis
//ToDo: debug middleware
//ToDo: split up code into packages for better readability
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
	"sort"
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
	Highscore []float64
}

type session struct {
	username string
	expiry   time.Time
}

var sessions = map[string]session{}

var CredentialDB []CredDB
var GlobalDB []DatabaseQN

var validate *validator.Validate

func main() {
	validate = validator.New()

	c := cron.New()
	c.AddFunc("@every 10s", delete1)
	c.Start()

	router := gin.Default()
	router.Use(limit())
	router.POST("/api", getQuestions)
	router.POST("/answers/:id", getScore)
	router.POST("/signin", Signin)
	router.POST("/signup", Signup)
	router.GET("/highscore", Highscore)
	router.GET("/refresh", refresh)
	router.GET("/logout", logout)
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

	cookie, err := c.Cookie("sessiontoken")
	if err != nil {
		if err == http.ErrNoCookie {
			return
		}
		return
	}
	addScore(cookie, comp)

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

	if err := c.ShouldBindJSON(&cred); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var hashedPassword []byte
	for i := range CredentialDB {
		if CredentialDB[i].Username == cred.Username {
			hashedPassword = CredentialDB[i].Password
			break
		}
	}
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(cred.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "You Password or Username was wrong dummy"})
		return
	}
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(3600 * time.Second)

	sessions[sessionToken] = session{
		username: cred.Username,
		expiry:   expiresAt,
	}

	c.SetCookie("sessiontoken", sessionToken, 3600, "/", "localhost", true, true)
	c.JSON(http.StatusAccepted, getHighscore(sessionToken))
}

func Highscore(c *gin.Context) {
	cookie, err := c.Cookie("sessiontoken")
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
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	c.JSON(http.StatusAccepted, getHighscore(sessionToken))
}

func refresh(c *gin.Context) {
	cookie, err := c.Cookie("sessiontoken")
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
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(3600 * time.Second)

	sessions[newSessionToken] = session{
		username: userSession.username,
		expiry:   expiresAt,
	}
	delete(sessions, sessionToken)

	c.SetCookie("sessiontoken", newSessionToken, 3600, "/", "localhost", true, true)
	c.JSON(http.StatusAccepted, nil)
}

func logout(c *gin.Context) {

	cookie, err := c.Cookie("sessiontoken")
	if err != nil {
		if err == http.ErrNoCookie {
			c.JSON(http.StatusUnauthorized, nil)
			return
		}
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	sessionToken := cookie
	delete(sessions, sessionToken)

	c.SetCookie("sessiontoken", sessionToken, -1, "/", "localhost", true, true)
	c.JSON(http.StatusAccepted, nil)
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

func delete1() {
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

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func getHighscore(sessionToken string) []float64 {

	//ERROR HANDLING
	var hscore []float64

	userSession, exists := sessions[sessionToken]
	if !exists {
		return hscore
	}
	for i := range CredentialDB {
		//needs error handling
		if CredentialDB[i].Username == userSession.username {
			hscore = CredentialDB[i].Highscore
			break
		}
	}
	return hscore
}

func addScore(sessionToken string, score float64) {

	userSession, exists := sessions[sessionToken]
	if !exists {
		return
	}
	for i := range CredentialDB {
		//needs error handling
		if CredentialDB[i].Username == userSession.username {
			insertSorted(CredentialDB[i].Highscore, score)
			break
		}
	}

}

func insertSorted(data []float64, v float64) []float64 {
	i := sort.Search(len(data), func(i int) bool { return data[i] >= v })
	return insertAt(data, i, v)
}

func insertAt(data []float64, i int, v float64) []float64 {
	if i == len(data) {
		return append(data, v)
	}
	data = append(data[:i+1], data[i:]...)

	data[i] = v

	return data
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
