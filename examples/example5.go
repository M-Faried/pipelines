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

func splitter(token *StringToken) []*StringToken {
	// Split the Value field by comma
	parts := strings.Split(token.Value, ",")

	// Create a slice to hold the resulting StringToken pointers
	result := make([]*StringToken, len(parts))

	// Populate the result slice with new StringToken instances
	for i, part := range parts {
		result[i] = &StringToken{Value: part, Original: token.Value}
	}

	return result
}

func trimSpaces(token *StringToken) *StringToken {
	token.Value = strings.TrimSpace(token.Value)
	return token
}

func addStars(token *StringToken) *StringToken {
	token.Value = fmt.Sprintf("*%s*", token.Value)
	return token
}

func tokenPrinter(token *StringToken) {
	fmt.Println("Result:", token.Value, "\tOriginal:", token.Original)
}

// Example5 demonstrates a pipeline with a step that fragments tokens.
func Example5() {

	builder := &pip.Builder[*StringToken]{}

	splitter := builder.NewStep(pip.StepFragmenterConfig[*StringToken]{
		Label:    "fragmenter",
		Replicas: 1,
		Process:  splitter,
	})
	trim := builder.NewStep(pip.StepBasicConfig[*StringToken]{
		Label:    "trim",
		Replicas: 2,
		Process:  trimSpaces,
	})
	stars := builder.NewStep(pip.StepBasicConfig[*StringToken]{
		Label:    "stars",
		Replicas: 2,
		Process:  addStars,
	})
	result := builder.NewStep(pip.StepTerminalConfig[*StringToken]{
		Label:    "result",
		Replicas: 2,
		Process:  tokenPrinter,
	})

	pConfig := pip.PipelineConfig{
		DefaultStepChannelSize: 10,
		TrackTokensCount:       true,
	}
	pipeline := builder.NewPipeline(pConfig, splitter, trim, stars, result)
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
