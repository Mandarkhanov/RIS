package domain

type CrackRequest struct {
	Hash      string `json:"hash"`
	MaxLength int    `json:"maxLength"`
}

type CrackResponse struct {
	RequestID string `json:"requestId"`
}

type RequestStatus string

const (
	StatusInProgress   RequestStatus = "IN_PROGRESS"
	StatusReady        RequestStatus = "READY"
	StatusPartialReady RequestStatus = "PARTIAL_READY"
	StatusError        RequestStatus = "ERROR"
)

type StatusResponse struct {
	Status RequestStatus `json:"status"`
	Data   []string      `json:"data"`
}
