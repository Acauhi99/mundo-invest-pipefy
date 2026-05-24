package pipefy

import "fmt"

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

func (c *Client) SimulateSend(payload map[string]interface{}) {
	fmt.Printf("[Pipefy] Simulando envio GraphQL: %+v\n", payload)
}
