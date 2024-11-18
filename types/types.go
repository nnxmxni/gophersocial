package types

type APIResponseBody struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Token   string      `json:"token,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
