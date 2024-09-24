package examples

import (
	"context"
	"fmt"
	"strings"

	"github.com/m-faried/pipelines"
)

type StringToken struct {
	Value    string
	Original string
}

func splitter(token *StringToken) ([]*StringToken, error) {
	// Split the Value field by comma
	parts := strings.Split(token.Value, ",")

	// Create a slice to hold the resulting StringToken pointers
	result := make([]*StringToken, len(parts))

	// Populate the result slice with new StringToken instances
	for i, part := range parts {
		result[i] = &StringToken{Value: part, Original: token.Value}
	}

	return result, nil
}

func trimSpaces(token *StringToken) (*StringToken, error) {
	token.Value = strings.TrimSpace(token.Value)
	return token, nil
}

func addStars(token *StringToken) (*StringToken, error) {
	token.Value = fmt.Sprintf("*%s*", token.Value)
	return token, nil
}

func tokenPrinter(token *StringToken) error {
	fmt.Println("Result:", token.Value, "\tOriginal:", token.Original)
	return nil
}

func Example5() {
	splitter := pipelines.NewStepFragmenter("fragmenter", 1, splitter)
	trim := pipelines.NewStep("trim", 2, trimSpaces)
	stars := pipelines.NewStep("stars", 2, addStars)
	result := pipelines.NewStepResult("result", 2, tokenPrinter)

	pipe := pipelines.NewPipeline[*StringToken](2, splitter, trim, stars, result)
	pipe.Init()

	ctx := context.Background()
	pipe.Run(ctx)

	// feeding inputs
	pipe.FeedOne(&StringToken{Value: "Apples, Oranges , Bananas"})

	// waiting for all tokens to be processed
	pipe.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipe.Terminate()

	fmt.Println("Example 5 Done !!!")
}
