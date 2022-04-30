package test

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Score struct {
	sco int
}

func test() {
	score := Score{}
	fmt.Printf("Welcome to mental maths training. \n")
	var numberTo int
	var questionNumber int
	fmt.Printf("Enter the highest number you wish to encounter: ")
	fmt.Scan(&numberTo)

	fmt.Printf("\nEnter the number of questions you wish to solve: ")
	fmt.Scan(&questionNumber)

	fmt.Println("Ready in:")

	for i := 3; i > 0; i-- {
		time.Sleep(1 * time.Second)
		var b = i
		fmt.Printf("%v \n", b)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("Go!")

	rand.Seed(time.Now().UnixNano())

	var sliceOne []int = randInt(numberTo, questionNumber)
	var sliceTwo []int = randInt(numberTo, questionNumber)
	var sliceOP []string = randOperator(questionNumber)
	var sliceIntAnswers []int = ansInt(sliceOne, sliceTwo, sliceOP)

	oldTime := time.Now()

	for i := range sliceOP {
		fmt.Printf("%v %v %v = ", sliceOne[i], sliceOP[i], sliceTwo[i])
		var input int
		fmt.Scan(&input)
		if input == sliceIntAnswers[i] {
			score.sco++
		}
	}
	currentTime := time.Now()

	diff := currentTime.Sub(oldTime)

	fmt.Printf("You scored %v out of %v in %v seconds!", score.sco, len(sliceOP), diff.Seconds())

	comp := (float64(score.sco)/diff.Seconds() + 1) * float64(numberTo) * (float64(score.sco) / float64(questionNumber))

	fmt.Printf(" \nThis results in a composite score of %v.", math.Round(comp*100)/100)

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
