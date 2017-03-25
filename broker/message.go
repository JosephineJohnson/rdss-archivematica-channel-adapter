package broker

// Message ...
type Message struct {
	Header MessageHeader `json:"messageHeader"`
	Body   MessageBody   `json:"messageBody"`
}

// MessageHeader ...
type MessageHeader struct {
	ID            string `json:"messageId"`
	CorrelationID string ``
	Type          string `json:"messageType"`
	ReturnAddress string
}

// MessageBody ...
type MessageBody struct {
	Payload string `json:"payload"`
}
