package main

import (
	"context"
	"fmt"
	"log"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
)

func main() {
	// Create provider once, reuse for multiple requests.
	provider, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	response, err := provider.Completion(ctx, anyllm.CompletionParams{
		Model: "gpt-5-nano",
		Messages: []anyllm.Message{
			{Role: anyllm.RoleUser, Content: "What was a positive news story from today?"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Choices[0].Message.Content)
}
