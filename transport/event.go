package transport

type EventMessage struct {
	Id      string
	ReplyTo string
	Type    string
	Data    string
}

func (this EventMessage) GetType() string {
	return "EventMessage"
}

func (this EventMessage) ToJson() string {
	return ""
}
