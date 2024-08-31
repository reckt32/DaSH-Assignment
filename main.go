package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Response struct {
	Prompt    string        `json:"Prompt"`
	Message   genai.Content `json:"Message"`
	TimeSent  int64         `json:"TimeSent"`
	TimeRecvd int64         `json:"TimeRecvd"`
	Source    string        `json:"Source"`
}

func main() {
	godotenv.Load()
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	// Read prompts from input.txt
	prompts, err := readLines("input.txt")
	if err != nil {
		log.Fatal(err)
	}

	var responses []Response

	for _, prompt := range prompts {
		timeSent := time.Now().Unix()

		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			log.Fatal(err)
		}

		timeRecvd := time.Now().Unix()

		for _, c := range resp.Candidates {
			if c.Content != nil {
				response := Response{
					Prompt:    prompt,
					Message:   *c.Content,
					TimeSent:  timeSent,
					TimeRecvd: timeRecvd,
					Source:    "Gemini",
				}
				responses = append(responses, response)
			}
		}
	}

	outputFile, err := os.Create("output.json")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(responses)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Responses written to output.json")
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
