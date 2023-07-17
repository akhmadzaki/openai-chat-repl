package main

import (
	"net/http"
	"log"
	"fmt"
	"io/ioutil"
	"bytes"
	"encoding/json"
	"time"
	_"flag"
	"os"
	"strings"
	"bufio"

	"github.com/joho/godotenv"
)

const (
	BASE_URL = "https://api.openai.com/v1"
	MODEL = "gpt-3.5-turbo"
)

type Timestamp struct {
    time.Time
}

func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
    var raw int64
    err := json.Unmarshal(bytes, &raw)

    if err != nil {
        fmt.Printf("error decoding timestamp: %s\n", err)
        return err
    }

    p.Time = time.Unix(raw, 0)
    return nil
}

type ChatRequest struct {
	Model string `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Id string `json:"id"`
	Object string `json:"object"`
	Created Timestamp `json:"created"`
	Model string `json:"model"`
	Choices []Choice `json:"choices"`
	Usage Usage `json:"usage"`
}

type Choice struct {
	Index int `json:"index"`
	Message Message `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens int `json:"total_tokens"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Message string `json:"message"`
	Type string `json:"type"`
	Param string `json:"param"`
	Code string `json:"code"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}
}

func main() {
	var token []string
	var command string

	for {
		fmt.Printf("$> ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		if text == "\r\n" || text == "\n" {
			continue
		}

		token = strings.Fields(text)
		command = token[0]

		switch command {
			case "exit":
				os.Exit(0)
			case "help":
				showHelp()
			// case "input":
			// 	if len(token) == 1 {
			// 		fmt.Println("Please provide input file as prompt.\n")
			// 		continue
			// 	}

			// 	if len(token) == 1 {
			// 		fmt.Println("Too many input file given.\n")
			// 		continue
			// 	}

			// 	parseInputText(token[1])
			case "prompt":
				inputPrompt := strings.Join(token[1:], " ")

				chatRequest := ChatRequest{
					Model: MODEL,
					Messages : []Message{
						Message{
							Role: "user",
							Content: inputPrompt,
						},
					},
				}

				marshalled, err := json.Marshal(chatRequest)
				if err != nil {
					log.Fatal(err)
				}

				postChat(marshalled)
			default:
				fmt.Println("Please provide correct input. Type help to show available command.\n")
				continue
		}
	}	
}

func showHelp() {
	fmt.Println("Usage: <command> [optional argument]\n")
	fmt.Println("where available commands are:")
	// fmt.Println("  input: Read prompt from input file")
	fmt.Println("  prompt: Read prompt from standard input")
	fmt.Println("  help: Show available command")
	fmt.Println("  exit: Exit program")
	fmt.Println()
}

func parseInputText(text string) (string, error) {
	return "", nil
}

func postChat(chatRequest []byte) {
	client := http.Client{}

	req , err := http.NewRequest("POST", BASE_URL + "/chat/completions", bytes.NewReader(chatRequest))
	if err != nil {
		log.Fatal(err)
	}

	SECRET_KEY := os.Getenv("SECRET_KEY")

	req.Header.Set("Authorization", "Bearer "+SECRET_KEY)
	req.Header.Set("Content-Type", "application/json")

	res , err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	body, error := ioutil.ReadAll(res.Body)
	if error != nil {
		log.Fatal(err)
	}

	var chatResponse ChatResponse
	var errorResponse ErrorResponse

	if(strings.Contains(string(body), "chat.completion")) {
		err = json.Unmarshal(body, &chatResponse)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(chatResponse.Choices[0].Message.Content)
	} else {
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(errorResponse.Error.Message)
	}
	fmt.Println()
}