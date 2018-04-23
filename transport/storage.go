package transport

type StorageMessage struct {
	Id     string
	Owner  string
	Scope  string
	Key    string
	Value  string
	Action string
}

func (this StorageMessage) GetType() string {
	return "StorageMessage"
}

func (this StorageMessage) ToJson() string {
	return ""
}
