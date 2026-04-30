package amqp

import (
	"context"
	"crackhash/manager/internal/service"
	"crackhash/pkg/domain"
	"encoding/xml"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	svc *service.CrackManagerService
}

func NewConsumer(svc *service.CrackManagerService) *Consumer {
	return &Consumer{svc: svc}
}

func (c *Consumer) HandleMessage(msg amqp.Delivery) {
	var resp domain.WorkerResponse
	if err := xml.Unmarshal(msg.Body, &resp); err != nil {
		log.Printf("Failed to unmarshal result XML: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("Received result for task %s (Part %d)", resp.RequestID, resp.PartNumber)

	err := c.svc.ProcessWorkerResult(context.Background(), resp)
	if err != nil {
		log.Printf("Failed to process worker result: %v", err)
		msg.Nack(false, true)
		return
	}

	if err := msg.Ack(false); err != nil {
		log.Printf("Failed to ACK result message: %v", err)
	}
}
