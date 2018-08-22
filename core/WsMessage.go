package core

type Message struct {
	Body   []byte `json:"body"`
	granulation float64
	patern string
}