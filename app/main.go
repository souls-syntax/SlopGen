package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"bufio"
	"os/exec"
	"runtime"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/souls-syntax/SlopGen/app/internal/model"
)
type ReadArgs struct {
    FilePath string `json:"file_path"`
}
func main() {
	// var NGROK_LINK string
	// flag.StringVar(&NGROK_LINK, )
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	var ngrok_url string
	flag.StringVar(&ngrok_url, "n", "", "Link of the superior intelligence")
	// So we are taking the prompt here. Likely

	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	// apiKey := os.Getenv("OPENROUTER_API_KEY")
	var model string
	baseUrl := ngrok_url + "/v1"
	if baseUrl == "" {
		baseUrl = "http://localhost:11434/v1"
		model = "qwen2.5:7b"
	} else {
		model = "gpt-oss:20b"
	}
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

	writeTool := model.Tool{
		Type: "function",
		Function: model.Function{
			Name:        "Write",
			Description: "Write content to a file. Creates the file if it doesn't exist",
			Parameters: model.Parameters{
				Type: "object",
				Properties: map[string]model.Property{
					"file_path": {
						Type:        "string",
						Description: "The path to the file to write",
					},
					"content" : {
						Type:        "string",
						Description: "Content to write to the file",
					},
				},
				Required: []string{"file_path", "content"},
			},
		},
	}

	executeTool := model.Tool{
		Type: "function",
		Function: model.Function{
			Name:        "Execute",
			Description: "Execute a shell command and return it's output (stdout + stderr)",
			Parameters: model.Parameters{
				Type: "object",
				Properties: map[string]model.Property{
					"command": {
						Type:        "string",
						Description: "The shell command to execute",
					},
				},
				Required: []string{"file_path", "content"},
			},
		},
	}
	res, _ := readTool.ConvertToOpenAITool()
	writ, _ := writeTool.ConvertToOpenAITool()
	exect, _ := executeTool.ConvertToOpenAITool()
	cwd, _ := os.Getwd()
	systemPrompt := fmt.Sprintf(
    "You are a helpful assistant. The current working directory is: %s. When asked to read a file, use the Read tool with paths relative to this directory.",
    cwd,
	)
	client := openai.NewClient(option.WithAPIKey("ollama"), option.WithBaseURL(baseUrl))
	messages := []openai.ChatCompletionMessageParamUnion{
				{
					OfSystem: &openai.ChatCompletionSystemMessageParam{
    				Content: openai.ChatCompletionSystemMessageParamContentUnion{
        			OfString: openai.String(systemPrompt),
    				},
					},
				},
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							// We need to feed history slice here.
							// How to feel array to openAI go api?
							OfString: openai.String(prompt),
						},
					},	
				},
			}
		tools :=  []openai.ChatCompletionToolUnionParam{res, writ, exect}
	reader := bufio.NewReader(os.Stdin)
	for {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model: model,
				Messages:messages ,
				Tools:tools,
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
		messages = append(messages, msg.ToParam())
		// fmt.Fprintf(os.Stderr, "ToolCalls: %+v\n", msg.ToolCalls)
		// fmt.Fprintf(os.Stderr, "Content: %s\n", msg.Content)
		if len(msg.ToolCalls) > 0 {

			for _, tc := range msg.ToolCalls{
				switch tc.Function.Name {
				case "Read":
					var args ReadArgs
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
							messages = append(messages, openai.ToolMessage(fmt.Sprintf("error parsing args: %v", err), tc.ID))
							continue
					}
					content, err := os.ReadFile(args.FilePath)
					if err != nil {
							messages = append(messages, openai.ToolMessage(fmt.Sprintf("error: %v", err), tc.ID))
							continue
					}
					messages = append(messages, openai.ToolMessage(string(content), tc.ID))
				case "Write":
					var wargs model.WriteArgs
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &wargs); err != nil {
							messages = append(messages, openai.ToolMessage(fmt.Sprintf("error parsing args: %v", err), tc.ID))
							continue
					}
					err = os.WriteFile(wargs.FilePath, []byte(wargs.Content), 0644)
					if err != nil {
							messages = append(messages, openai.ToolMessage(fmt.Sprintf("error writing content: %v", err), tc.ID))
							continue
					} else {
						messages = append(messages, openai.ToolMessage("File written successfully", tc.ID))
					}
				case "Execute":
					var xargs model.ExecuteArgs 
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &xargs); err != nil {
							messages = append(messages, openai.ToolMessage(fmt.Sprintf("error parsing args: %v", err), tc.ID))
							continue
					}
					fmt.Printf("Execute: %q — confirm? [y/N]: ", xargs.Command)
					input, err := reader.ReadString('\n')
					input = strings.TrimSpace(input)
					if input != "y" {
						messages = append(messages, openai.ToolMessage(fmt.Sprintf("Command was rejected by admin"), tc.ID))
						continue
					}
					var cmd *exec.Cmd 
					if runtime.GOOS == "windows" {
  			  	cmd = exec.Command("cmd", "/C", xargs.Command)
					} else {
    				cmd = exec.Command("sh", "-c", xargs.Command)
					}
					out, err := cmd.CombinedOutput()
					fmt.Println(string(out))
					if err != nil {
							messages = append(messages, openai.ToolMessage(fmt.Sprintf("error executing the command: %v", err), tc.ID))
							continue
					} else {
						messages = append(messages, openai.ToolMessage(string(out), tc.ID))
					}

				}
			}
		} else {
			fmt.Print(msg.Content)
			fmt.Println(" >>> ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			messages = append(messages, openai.UserMessage(string(input)))
		}

		// fmt.Print(resp.Choices[0].Message.Content)
	}
}
