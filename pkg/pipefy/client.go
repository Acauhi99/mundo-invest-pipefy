package pipefy

import (
	"fmt"
	"time"
)

type PipefyClient interface {
	SimulateSend(payload map[string]interface{}) string
	BuildCreateCardPayload(input CreateCardInput) map[string]interface{}
	BuildUpdateCardFieldPayload(input UpdateCardFieldInput) map[string]interface{}
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) BuildCreateCardPayload(input CreateCardInput) map[string]interface{} {
	return map[string]interface{}{
		"query": CreateCardMutation,
		"variables": map[string]interface{}{
			"input": input,
		},
	}
}

func (c *Client) BuildUpdateCardFieldPayload(input UpdateCardFieldInput) map[string]interface{} {
	return map[string]interface{}{
		"query": UpdateCardFieldMutation,
		"variables": map[string]interface{}{
			"input": input,
		},
	}
}

func (c *Client) SimulateSend(payload map[string]interface{}) string {
	cardID := fmt.Sprintf("card_sim_%d", time.Now().UnixNano())
	fmt.Printf("[Pipefy] Card ID simulado: %s\n", cardID)
	return cardID
}
