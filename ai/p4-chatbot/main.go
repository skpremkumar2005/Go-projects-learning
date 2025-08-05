package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

// GeminiRequest is the request format for Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type Part struct {
	Text string `json:"text"`
}

// GeminiResponse is the response format from Gemini
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set")
	}

	e := echo.New()

	e.POST("/chat", func(c echo.Context) error {
		var req ChatRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		geminiReq := GeminiRequest{
			Contents: []Content{
				{
					Role: "user",
					Parts: []Part{
						{Text: req.Message},
					},
				},
			},
		}

		bodyBytes, err := json.Marshal(geminiReq)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to serialize request"})
		}

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to contact Gemini API"})
		}
		defer resp.Body.Close()

		resBody, _ := ioutil.ReadAll(resp.Body)
		var geminiResp GeminiResponse
		if err := json.Unmarshal(resBody, &geminiResp); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse Gemini response"})
		}

		if len(geminiResp.Candidates) == 0 {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "No response from Gemini"})
		}

		reply := geminiResp.Candidates[0].Content.Parts[0].Text
		return c.JSON(http.StatusOK, ChatResponse{Response: reply})
	})

	e.Logger.Fatal(e.Start(":8086"))
}
