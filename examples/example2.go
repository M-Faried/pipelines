package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/m-faried/pipelines"
)

type Token struct {
	values    []string
	currValue string
	count     int
}

func processToken(t *Token) (*Token, error) {
	t.currValue = fmt.Sprintf("*%s*", t.currValue)
	t.values = append(t.values, t.currValue)
	t.count++
	return t, nil
}

func printToken(t *Token) error {
	// for _, v := range t.values {
	// 	fmt.Println(v)
	// }
	// fmt.Println("Steps Count:", t.count)
	fmt.Println("Result:", t.currValue)
	return nil
}

func Example2() {
	ctx, cancel := context.WithCancel(context.Background())

	step1 := pipelines.NewStep("tokenStep1", 1, processToken)
	step2 := pipelines.NewStep("tokenStep2", 1, processToken)
	step3 := pipelines.NewStep("tokenStep3", 1, processToken)
	resultStep := pipelines.NewResultStep("printTokenStep", 1, printToken)
	pip := pipelines.NewPipeline(10, resultStep, step1, step2, step3)
	pip.Init()

	// Running
	go pip.Run(ctx)
	time.Sleep(1 * time.Second)

	// Feeding inputs
	pip.FeedOne(&Token{values: []string{}, currValue: "Hello"})
	pip.FeedOne(&Token{values: []string{}, currValue: "World"})
	pip.FeedMany([]*Token{
		{values: []string{}, currValue: "Welcome"},
		{values: []string{}, currValue: "All"},
	})

	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)
	fmt.Println("Example2 Done!!!")
}
