package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var ollamaURL string
var llmModel string

type Request struct {
	Text string `json:"text"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaChunk represents a single chunk from the LLM response.
type OllamaChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Failed to write JSON response:", err)
	}
}

func generateResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Invalid request method"})
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	var prompt string = "Please provide a concise and accurate summary of the text below. Return only the summary in the same language as the original text, with no extra commentary or formatting.\n\n" + req.Text

	ollamaReq := OllamaRequest{
		Model:  llmModel,	
		Prompt: prompt,
	}
	fmt.Println("Ollama request:", ollamaReq)
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		log.Println("Error marshalling request:", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println("Error contacting Ollama:", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to contact LLM"})
		return
	}
	defer resp.Body.Close()
	fmt.Println("Ollama response:", resp.Body)

	// Stream-decode each JSON chunk and accumulate the response.
	var fullResponse string
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk OllamaChunk
		if err := decoder.Decode(&chunk); err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error decoding chunk from Ollama:", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Invalid response from LLM"})
			return
		}
		fullResponse += chunk.Response
		if chunk.Done {
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"response": fullResponse})
}

func main() {
	ollamaURL = os.Getenv("OLLAMA_API_URL")
	if ollamaURL == "" {
		log.Fatal("OLLAMA_API_URL environment variable is not set")
	}
	llmModel = os.Getenv("LLM_MODEL")
	if llmModel == "" {  // Corrected check here
		log.Fatal("LLM_MODEL environment variable is not set")
	}

	http.HandleFunc("/summurize", generateResponse)
	fmt.Println("Go API is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
