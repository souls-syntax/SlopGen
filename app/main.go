package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/souls-syntax/SlopGen/app/internal/model"
)

func main() {
	// var NGROK_LINK string
	// flag.StringVar(&NGROK_LINK, )
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	// So we are taking the prompt here. Likely

	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	// apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://172.19.240.1:11434/v1"
	}
	// if apiKey == "" {
	// 	panic("Env variable OPENROUTER_API_KEY not found")
	// }
	// history := make([]string,0)
	readTool := model.Tool{
		Type: "function",
		Function: model.Function{
			Name:        "Read",
			Description: "Read and return the contents of a file",
			Parameters: model.Parameters{
				Type: "object",
				Properties: map[string]model.Property{
					"file_path": {
						Type:        "string",
						Description: "The path to the file to read",
					},
				},
				Required: []string{"file_path"},
			},
		},
	}
	res, _ := readTool.ConvertToOpenAITool()

	client := openai.NewClient(option.WithAPIKey("ollama"), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: "deepseek-coder:6.7b",
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							// We need to feed history slice here.
							// How to feel array to openAI go api?
							OfString: openai.String(prompt),
						},
					},
				},
			},
			Tools: []openai.ChatCompletionToolUnionParam{res},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	choice := resp.Choices[0]
	msg := choice.Message

	if len(msg.ToolCalls) > 0 {
		toolCall := msg.ToolCalls[0]
		var args model.Args

		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			fmt.Fprintln(os.Stderr, "failed to parse tool args:", err)
			os.Exit(1)
		}
	}

	// TODO: Uncomment the line below to pass the first stage
	fmt.Print(resp.Choices[0].Message.Content)
}
