package transport

type Message struct {
	Id      string
	Type    string
	Owner   string
	Scope   string
	Payload MessageType
}

type MessageType interface {
	GetType() string
	ToJson() string
}
