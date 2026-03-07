package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Model       string    `json:"model,omitempty"`
	TopP        int       `json:"top_p,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type ChatHistory struct {
	Messages []Message
}

// AddMessage adds a message to history
func (h *ChatHistory) AddMessage(role, content string) {
	h.Messages = append(h.Messages, Message{
		Role:    role,
		Content: content,
	})
}

// MCP JSON-RPC types
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallShopifyMCP calls the Shopify MCP server to get relevant information
func callShopifyMCP(query string) (string, error) {
	mcpURL := "https://your-custom-mcp-url.com/api/mcp"

	// First, list available tools
	listToolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	jsonData, err := json.Marshal(listToolsReq)
	if err != nil {
		return "", fmt.Errorf("error marshaling tools/list request: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating MCP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")


	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling MCP server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading MCP response: %v", err)
	}

	var toolsResp MCPResponse
	if err := json.Unmarshal(body, &toolsResp); err != nil {
		return "", fmt.Errorf("error parsing MCP response: %v", err)
	}

	if toolsResp.Error != nil {
		return "", fmt.Errorf("MCP error: %s", toolsResp.Error.Message)
	}

	// Parse tools list
	var toolsList struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(toolsResp.Result, &toolsList); err != nil {
		return "", fmt.Errorf("error parsing tools list: %v", err)
	}

	// Determine which tool to call based on the query
	var toolName string
	var toolArgs map[string]interface{}

	queryLower := strings.ToLower(query)

	// Check if query is about products
	if strings.Contains(queryLower, "product") || strings.Contains(queryLower, "item") ||
		strings.Contains(queryLower, "buy") || strings.Contains(queryLower, "price") ||
		strings.Contains(queryLower, "search") || strings.Contains(queryLower, "find") {
		// Try to find search_shop_catalog tool
		for _, tool := range toolsList.Tools {
			if tool.Name == "search_shop_catalog" || strings.Contains(tool.Name, "search") {
				toolName = tool.Name
				toolArgs = map[string]interface{}{
					"query":   query,
					"context": "User is asking about products in the store",
					"limit":   10,
				}
				break
			}
		}
	}

	// Check if query is about policies, shipping, returns, etc.
	if strings.Contains(queryLower, "policy") || strings.Contains(queryLower, "shipping") ||
		strings.Contains(queryLower, "return") || strings.Contains(queryLower, "refund") ||
		strings.Contains(queryLower, "faq") || strings.Contains(queryLower, "question") {
		for _, tool := range toolsList.Tools {
			if tool.Name == "search_shop_policies_and_faqs" || strings.Contains(tool.Name, "policy") ||
				strings.Contains(tool.Name, "faq") {
				toolName = tool.Name
				toolArgs = map[string]interface{}{
					"query":   query,
					"context": "User is asking about store policies or FAQs",
				}
				break
			}
		}
	}

	// If no specific tool found, try the first available tool
	if toolName == "" && len(toolsList.Tools) > 0 {
		toolName = toolsList.Tools[0].Name
		toolArgs = map[string]interface{}{
			"query":   query,
			"context": "User query about the store",
		}
	}

	if toolName == "" {
		return "", fmt.Errorf("no tools available from MCP server")
	}

	// Call the selected tool
	toolCallReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: MCPToolCallParams{
			Name:      toolName,
			Arguments: toolArgs,
		},
	}

	jsonData, err = json.Marshal(toolCallReq)
	if err != nil {
		return "", fmt.Errorf("error marshaling tool call request: %v", err)
	}

	req, err = http.NewRequest("POST", mcpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating tool call request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling MCP tool: %v", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading tool response: %v", err)
	}

	var toolResp MCPResponse
	if err := json.Unmarshal(body, &toolResp); err != nil {
		return "", fmt.Errorf("error parsing tool response: %v", err)
	}

	if toolResp.Error != nil {
		return "", fmt.Errorf("MCP tool error: %s", toolResp.Error.Message)
	}

	// Format the result as a string to include in context
	var resultContent string
	if err := json.Unmarshal(toolResp.Result, &resultContent); err != nil {
		// If it's not a string, try to format it nicely
		var resultObj interface{}
		if err := json.Unmarshal(toolResp.Result, &resultObj); err == nil {
			formatted, _ := json.MarshalIndent(resultObj, "", "  ")
			resultContent = string(formatted)
		} else {
			resultContent = string(toolResp.Result)
		}
	}

	fmt.Println(resultContent)
	return resultContent, nil
}

// isShopifyQuery checks if the query is related to Shopify/store
func isShopifyQuery(query string) bool {
	queryLower := strings.ToLower(query)
	shopifyKeywords := []string{
		"product", "item", "buy", "purchase", "price", "cart", "checkout",
		"shipping", "delivery", "return", "refund", "policy", "faq",
		"store", "shop", "catalog", "inventory", "order", "payment",
	}
	for _, keyword := range shopifyKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}
	return false
}

// GetChatbotResponse sends a request to GitHub Models API
func getChatbotResponse(input string, context string, history *ChatHistory) (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	curr := time.Now()
	fmt.Printf("Request to chatbot at time: %s\n", curr)

	// Build messages array
	messages := []Message{}

	// Check if query is Shopify-related and fetch data from MCP
	shopifyContext := ""
	if isShopifyQuery(input) {
		fmt.Printf("Detected Shopify-related query, calling MCP server...\n")
		mcpData, err := callShopifyMCP(input)
		if err != nil {
			fmt.Printf("Warning: Failed to get Shopify MCP data: %v\n", err)
			// Continue without MCP data rather than failing
		} else {
			shopifyContext = fmt.Sprintf("\n\n[Shopify Store Information]\n%s\n\nUse this information to answer the user's question about the store. If the information is not relevant, you can ignore it.", mcpData)
		}
	}

	// Add system context if provided
	combinedContext := context
	if shopifyContext != "" {
		combinedContext = context + shopifyContext
	}

	if combinedContext != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: combinedContext,
		})
	}

	// Add conversation history
	if history != nil {
		messages = append(messages, history.Messages...)
	}

	// Add current user input
	messages = append(messages, Message{
		Role:    "user",
		Content: input,
	})

	// Prepare request
	reqBody := ChatRequest{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
		Model:       "openai/gpt-4o-mini",
		TopP:        1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	url := "https://models.github.ai/inference/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+githubToken)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	response := chatResp.Choices[0].Message.Content

	// Add to history
	if history != nil {
		history.AddMessage("user", input)
		history.AddMessage("assistant", response)
	}

	fmt.Printf("Chatbot response time: %v\n", time.Since(curr))
	return response, nil
}
