package service

import (
	"context"
	"crackhash/manager/internal/config"
	"crackhash/manager/internal/repository"
	"crackhash/pkg/domain"
	"crackhash/pkg/rabbitmq"
	"encoding/xml"
	"log"
	"time"

	"github.com/google/uuid"
)

var alphabet = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
	"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

const (
	WorkerTaskPath   = "/internal/api/worker/hash/crack/task"
	WorkerCancelPath = "/internal/api/worker/hash/crack/cancel"
)

type CrackManagerService struct {
	cfg       *config.Config
	taskRepo  *repository.TaskRepository
	rmqClient *rabbitmq.Client
}

func NewCrackManagerService(cfg *config.Config, taskRepo *repository.TaskRepository, rmqClient *rabbitmq.Client) *CrackManagerService {
	return &CrackManagerService{
		cfg:       cfg,
		taskRepo:  taskRepo,
		rmqClient: rmqClient,
	}
}

func (s *CrackManagerService) CreateTask(ctx context.Context, req domain.CrackRequest) (string, error) {
	parts := make([]int, s.cfg.TaskPartsCount)
	for i := 0; i < s.cfg.TaskPartsCount; i++ {
		parts[i] = i + 1
	}

	newState := &repository.RequestState{
		ReqID:           uuid.New().String(),
		Hash:            req.Hash,
		MaxLength:       req.MaxLength,
		Status:          domain.StatusInProgress,
		Data:            []string{},
		WorkersFinished: 0,
		TotalWorkers:    s.cfg.TaskPartsCount,
		PendingParts:    parts,
		CreatedAt:       time.Now(),
	}

	finalReqID, err := s.taskRepo.StoreOrGetExists(ctx, newState)
	if err != nil {
		return "", err
	}

	if finalReqID == newState.ReqID {
		go s.DispatchTasks(newState.ReqID, req.Hash, req.MaxLength, parts)
	}
	return finalReqID, nil
}

func (s *CrackManagerService) DispatchTasks(reqID string, hash string, maxLength int, pendingParts []int) {
	for _, partNum := range pendingParts {
		task := domain.WorkerTask{
			RequestID:  reqID,
			Hash:       hash,
			MaxLength:  maxLength,
			PartNumber: partNum,
			PartCount:  s.cfg.TaskPartsCount,
			Alphabet:   domain.Alphabet{Symbols: alphabet},
		}

		body, err := xml.Marshal(task)
		if err != nil {
			log.Printf("Failed to marshal task part %d: %v", partNum, err)
			continue
		}

		err = s.rmqClient.PublishXML(context.Background(), s.cfg.TasksQueue, body)
		if err != nil {
			log.Printf("RMQ is down. Task [%s] (Part %d) stays in Outbox. Err: %v", reqID, partNum, err)
		} else {
			s.taskRepo.MarkPartDispatched(context.Background(), reqID, partNum)
			log.Printf("Task [%s] (part %d/%d) published and removed from Outbox", reqID, partNum, s.cfg.TaskPartsCount)
		}
	}
}

func (s *CrackManagerService) OutboxRelay() {
	for {
		time.Sleep(2 * time.Second)

		tasks, err := s.taskRepo.GetTasksWithPendingParts(context.Background())
		if err != nil || len(tasks) == 0 {
			continue
		}

		for _, t := range tasks {
			log.Printf("OutboxRelay: Found pending parts %v for task [%s], retrying...", t.PendingParts, t.ReqID)
			s.DispatchTasks(t.ReqID, t.Hash, t.MaxLength, t.PendingParts)
		}
	}
}

func (s *CrackManagerService) GetTaskStatus(ctx context.Context, reqID string) (domain.StatusResponse, bool) {
	state, exists := s.taskRepo.Get(ctx, reqID)
	if !exists {
		return domain.StatusResponse{}, false
	}

	var responseData []string
	if state.Status == domain.StatusReady || state.Status == domain.StatusPartialReady {
		responseData = state.Data
	}

	return domain.StatusResponse{
		Status: state.Status,
		Data:   responseData,
	}, true
}

func (s *CrackManagerService) ProcessWorkerResult(ctx context.Context, resp domain.WorkerResponse) error {
	updatedState, err := s.taskRepo.AddWorkerResult(ctx, resp.RequestID, resp.Answers.Words)
	if err != nil {
		log.Printf("Failed to update db for task [%s]: %v", resp.RequestID, err)
		return err
	}

	if updatedState.Status == domain.StatusReady {
		log.Printf("Task [%s] is READY. Found %d words.", resp.RequestID, len(updatedState.Data))
	}
	return nil
}

func (s *CrackManagerService) TimeoutWatcher() {
	for {
		time.Sleep(s.cfg.WatchInterval)
		ctx := context.Background()

		tasksToCancel, err := s.taskRepo.GetAndMarkTimedOutTasks(ctx, s.cfg.TaskTimeout)
		if err != nil {
			log.Printf("Error watching timeouts: %v", err)
			continue
		}

		for _, reqID := range tasksToCancel {
			log.Printf("Task [%s] timed out. Cancelling.", reqID)
			s.CancelTask(reqID)
		}
	}
}

func (s *CrackManagerService) CancelTask(reqID string) {
	cancelReq := domain.WorkerCancelRequest{RequestID: reqID}

	body, err := xml.Marshal(cancelReq)
	if err != nil {
		log.Printf("Failed marshalling cancel request: %v", err)
		return
	}

	err = s.rmqClient.PublishFanoutXML(context.Background(), "cancel_exchange", body)
	if err != nil {
		log.Printf("Failed to publish cancel message: %v", err)
	} else {
		log.Printf("Cancel signal for task [%s] sent to all workers via fanout", reqID)
	}
}
