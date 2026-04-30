package amqp

import (
	"crackhash/pkg/domain"
	"crackhash/worker/internal/service"
	"encoding/xml"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type CancelConsumer struct {
	svc *service.CrackWorkerService
}

func NewCancelConsumer(svc *service.CrackWorkerService) *CancelConsumer {
	return &CancelConsumer{svc: svc}
}

func (c *CancelConsumer) HandleMessage(msg amqp.Delivery) {
	var req domain.WorkerCancelRequest
	if err := xml.Unmarshal(msg.Body, &req); err != nil {
		log.Printf("Failed to unmarshal cancel XML: %v", err)
		return
	}

	log.Printf("Received CANCEL signal for task [%s]", req.RequestID)
	c.svc.CancelTask(req.RequestID)
}
