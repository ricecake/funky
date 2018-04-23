package transport

type HttpMessage struct {
	Headers map[string][]string
	Method  string
	Url     string
	Host    string
	Body    []byte
	Code    uint
}

func (this HttpMessage) GetType() string {
	return "HttpMessage"
}

func (this HttpMessage) ToJson() string {
	return ""
}
