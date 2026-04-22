package domain

type WorkerTask struct {
	RequestID  string   `xml:"RequestId"`
	Hash       string   `xml:"Hash"`
	MaxLength  int      `xml:"MaxLength"`
	PartNumber int      `xml:"PartNumber"`
	PartCount  int      `xml:"PartCount"`
	Alphabet   Alphabet `xml:"Alphabet"`
}

type Alphabet struct {
	Symbols []string `xml:"symbols"`
}

type WorkerResponse struct {
	RequestID  string  `xml:"RequestId"`
	PartNumber int     `xml:"PartNumber"`
	Answers    Answers `xml:"Answers"`
}

type Answers struct {
	Words []string `xml:"words"`
}

type WorkerCancelRequest struct {
	RequestID string `xml:"requestId"`
}
