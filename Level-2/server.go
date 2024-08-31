// server.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Request struct {
	ClientID string `json:"ClientID"`
	Prompt   string `json:"Prompt"`
}

type Response struct {
	ClientID  string        `json:"ClientID"`
	Prompt    string        `json:"Prompt"`
	Message   genai.Content `json:"Message"`
	TimeSent  int64         `json:"TimeSent"`
	TimeRecvd int64         `json:"TimeRecvd"`
	Source    string        `json:"Source"`
}

func handleConnection(conn net.Conn, model *genai.GenerativeModel) {
	defer conn.Close()

	var req Request
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&req); err != nil {
		log.Println("Failed to decode request:", err)
		return
	}

	timeSent := time.Now().Unix()
	ctx := context.Background()
	resp, err := model.GenerateContent(ctx, genai.Text(req.Prompt))
	if err != nil {
		log.Println("Failed to generate content:", err)
		return
	}
	timeRecvd := time.Now().Unix()

	var content genai.Content
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		content = *resp.Candidates[0].Content
	}

	response := Response{
		ClientID:  req.ClientID,
		Prompt:    req.Prompt,
		Message:   content,
		TimeSent:  timeSent,
		TimeRecvd: timeRecvd,
		Source:    "Gemini",
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		log.Println("Failed to encode response:", err)
		return
	}
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

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Server listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn, model)
	}
}
