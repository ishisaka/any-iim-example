// 例: ツール/関数呼び出し
//
// このサンプルでは、any-llm-go でツール（関数呼び出し）を使用する方法を示しています。
//
// 実行方法:
//
//	export OPENAI_API_KEY="sk-..."
//	go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
)

// モデルが呼び出せるツールを定義します。
var tools = []anyllm.Tool{
	{
		Type: "function",
		Function: anyllm.Function{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The city name, e.g., 'Paris' or 'New York'",
					},
					"unit": map[string]any{
						"type":        "string",
						"enum":        []string{"celsius", "fahrenheit"},
						"description": "Temperature unit",
					},
				},
				"required": []string{"location"},
			},
		},
	},
}

// simulateWeather は指定された位置と温度単位の天気情報をシミュレートします。
//
// 引数:
//   - location: 天気情報を取得する場所を指定します。
//   - unit: 温度の単位を指定します。空文字列の場合はデフォルトで "celsius" が使用されます。
//
// 戻り値:
//   - JSON形式の天気情報文字列を返します。含まれる情報は位置情報、気温(22度固定)、
//     温度単位、天気条件("sunny"固定)です。
//
// 注記:
//   実際のアプリケーションではここで実際の天気APIを呼び出すようにしてください。
func simulateWeather(location, unit string) string {
	if unit == "" {
		unit = "celsius"
	}
	// 実際のアプリではここで実際の天気API を呼び出します。
	return fmt.Sprintf(`{"location": "%s", "temperature": 22, "unit": "%s", "condition": "sunny"}`, location, unit)
}

func main() {
	// OpenAIプロバイダーを初期化します
	provider, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 天気について質問する初期メッセージを設定します
	messages := []anyllm.Message{
		{Role: anyllm.RoleUser, Content: "What's the weather like in Paris?"},
	}

	fmt.Println("User: What's the weather like in Paris?")
	fmt.Println()

	// 最初のリクエスト - モデルがツールを呼び出すことがあります
	response, err := provider.Completion(ctx, anyllm.CompletionParams{
		Model:      "gpt-5-nano",
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "auto",
	})
	if err != nil {
		log.Fatal(err)
	}

	// モデルがツールを呼び出したいかどうかを確認します
	if response.Choices[0].FinishReason == anyllm.FinishReasonToolCalls {
		fmt.Println("Model is calling tools...")

		// アシスタントのメッセージ(ツール呼び出しを含む)を会話に追加します
		messages = append(messages, response.Choices[0].Message)

		// 各ツール呼び出しを処理します
		for _, tc := range response.Choices[0].Message.ToolCalls {
			fmt.Printf("  Tool: %s\n", tc.Function.Name)
			fmt.Printf("  Arguments: %s\n", tc.Function.Arguments)

			// 引数をパースします
			var args struct {
				Location string `json:"location"`
				Unit     string `json:"unit"`
			}
			if unmarshalErr := json.Unmarshal([]byte(tc.Function.Arguments), &args); unmarshalErr != nil {
				log.Fatal(unmarshalErr)
			}

			// 関数を実行します
			result := simulateWeather(args.Location, args.Unit)
			fmt.Printf("  Result: %s\n", result)
			fmt.Println()

			// ツール結果を会話に追加します
			messages = append(messages, anyllm.Message{
				Role:       anyllm.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		// ツール結果を含めて会話を続けます
		response, err = provider.Completion(ctx, anyllm.CompletionParams{
			Model:    "gpt-5-nano",
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// 最終的なレスポンスを出力します
	fmt.Printf("Assistant: %s\n", response.Choices[0].Message.Content)
}
