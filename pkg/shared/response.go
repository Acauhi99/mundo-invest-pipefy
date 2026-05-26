package shared

type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(data any) APIResponse {
	return APIResponse{Success: true, Data: data}
}

func ValidationError(message string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   &APIError{Code: "VALIDATION_ERROR", Message: message},
	}
}

func NotFoundError(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	}
}

func ConflictError(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	}
}

func InternalError(message string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   &APIError{Code: "INTERNAL_ERROR", Message: message},
	}
}
