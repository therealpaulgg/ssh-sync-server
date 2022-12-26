package live

type ServerMessage struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}
