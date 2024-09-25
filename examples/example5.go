package examples

import (
	"context"
	"fmt"
	"strings"

	pip "github.com/m-faried/pipelines"
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

// Example5 demonstrates a pipeline with a step that fragments tokens.
func Example5() {

	splitter := pip.NewStep[*StringToken](&pip.StepFragmenterConfig[*StringToken]{
		Label:    "fragmenter",
		Replicas: 1,
		Process:  splitter,
	})
	trim := pip.NewStep[*StringToken](&pip.StepConfig[*StringToken]{
		Label:    "trim",
		Replicas: 2,
		Process:  trimSpaces,
	})
	stars := pip.NewStep[*StringToken](&pip.StepConfig[*StringToken]{
		Label:    "stars",
		Replicas: 2,
		Process:  addStars,
	})
	result := pip.NewStep[*StringToken](&pip.StepResultConfig[*StringToken]{
		Label:    "result",
		Replicas: 2,
		Process:  tokenPrinter,
	})

	pipeline := pip.NewPipeline[*StringToken](10, splitter, trim, stars, result)
	pipeline.Init()

	ctx := context.Background()
	pipeline.Run(ctx)

	// feeding inputs
	pipeline.FeedOne(&StringToken{Value: "Apples, Oranges , Bananas"})

	// waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 5 Done !!!")
}
