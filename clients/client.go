package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Request struct {
	ClientID string `json:"ClientID"`
	Prompt   string `json:"Prompt"`
}

type Response struct {
	ClientID  string          `json:"ClientID"`
	Prompt    string          `json:"Prompt"`
	Message   json.RawMessage `json:"Message"`
	TimeSent  int64           `json:"TimeSent"`
	TimeRecvd int64           `json:"TimeRecvd"`
	Source    string          `json:"Source"`
}

func main() {
	godotenv.Load()

	if len(os.Args) < 3 {
		log.Fatal("Usage: go run client.go <ClientID> <NumClients>")
	}
	clientID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid ClientID")
	}
	numClients, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("Invalid NumClients")
	}

	prompts, err := readLines("/input.txt")
	if err != nil {
		log.Fatal(err)
	}

	var clientPrompts []string
	for i, prompt := range prompts {
		if i%numClients == clientID {
			clientPrompts = append(clientPrompts, prompt)
		}
	}

	outputFile := fmt.Sprintf("output_client%d.json", clientID)
	var responses []Response

	for _, prompt := range clientPrompts {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			log.Fatal(err)
		}

		req := Request{ClientID: fmt.Sprintf("client%d", clientID), Prompt: prompt}
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(req); err != nil {
			log.Println("Failed to encode request:", err)
			conn.Close()
			continue
		}

		var resp Response
		decoder := json.NewDecoder(conn)
		if err := decoder.Decode(&resp); err != nil {
			log.Println("Failed to decode response:", err)
			conn.Close()
			continue
		}
		conn.Close()

		if resp.ClientID != fmt.Sprintf("client%d", clientID) {
			resp.Source = "user"
		}

		responses = append(responses, resp)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(responses); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client %d responses written to %s\n", clientID, outputFile)
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
