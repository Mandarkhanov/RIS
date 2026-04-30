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
	newState := &repository.RequestState{
		ReqID:           uuid.New().String(),
		Hash:            req.Hash,
		MaxLength:       req.MaxLength,
		Status:          domain.StatusInProgress,
		Data:            []string{},
		WorkersFinished: 0,
		TotalWorkers:    s.cfg.TaskPartsCount,
		CreatedAt:       time.Now(),
	}

	finalReqID, err := s.taskRepo.StoreOrGetExists(ctx, newState)
	if err != nil {
		return "", err
	}

	if finalReqID == newState.ReqID {
		go s.DispatchTasks(newState.ReqID, req)
	}
	return finalReqID, nil
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
		log.Printf("Failed to update db for task %s: %v", resp.RequestID, err)
		return err
	}

	if updatedState.Status == domain.StatusReady {
		log.Printf("Task %s is READY. Found %d words.", resp.RequestID, len(updatedState.Data))
	}
	return nil
}

func (s *CrackManagerService) DispatchTasks(reqID string, req domain.CrackRequest) {
	for i := 0; i < s.cfg.TaskPartsCount; i++ {
		task := domain.WorkerTask{
			RequestID:  reqID,
			Hash:       req.Hash,
			MaxLength:  req.MaxLength,
			PartNumber: i + 1,
			PartCount:  s.cfg.TaskPartsCount,
			Alphabet:   domain.Alphabet{Symbols: alphabet},
		}

		body, err := xml.Marshal(task)
		if err != nil {
			log.Printf("Failed to marshal task part %d: %v", i+1, err)
			continue
		}

		err = s.rmqClient.PublishXML(context.Background(), s.cfg.TasksQueue, body)
		if err != nil {
			log.Printf("Failed to publish task part %d to RMQ: %v", i+1, err)
		} else {
			log.Printf("Task %s (part %d/%d) published to RMQ", reqID, i+1, s.cfg.TaskPartsCount)
		}
	}
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
			log.Printf("Task %s timed out. Cancelling.", reqID)
			s.CancelTask(reqID)
		}
	}
}

func (s *CrackManagerService) CancelTask(reqID string) {
	log.Printf("Cancellation logic for task %s will be implemented via RMQ Fanout later", reqID)
}
