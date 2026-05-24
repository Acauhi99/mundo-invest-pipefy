package pipefy

type CreateCardInput struct {
	PipeID           int              `json:"pipe_id"`
	Title            string           `json:"title"`
	FieldsAttributes []FieldAttribute `json:"fields_attributes"`
}

type FieldAttribute struct {
	FieldID    string `json:"field_id"`
	FieldValue string `json:"field_value"`
}

type UpdateCardFieldInput struct {
	CardID   string `json:"card_id"`
	FieldID  string `json:"field_id"`
	NewValue string `json:"new_value"`
}
