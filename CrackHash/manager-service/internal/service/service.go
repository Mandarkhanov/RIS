package service

import (
	"bytes"
	"context"
	"crackhash/manager/internal/config"
	"crackhash/manager/internal/repository"
	"crackhash/pkg/domain"
	"encoding/xml"
	"log"
	"net/http"
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
	cfg      *config.Config
	taskRepo *repository.TaskRepository
}

func NewCrackManagerService(cfg *config.Config, taskRepo *repository.TaskRepository) *CrackManagerService {
	return &CrackManagerService{
		cfg:      cfg,
		taskRepo: taskRepo,
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
		TotalWorkers:    len(s.cfg.WorkerURLs),
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
	client := &http.Client{Timeout: s.cfg.ClientTimeout}

	for i, workerURL := range s.cfg.WorkerURLs {
		task := domain.WorkerTask{
			RequestID:  reqID,
			Hash:       req.Hash,
			MaxLength:  req.MaxLength,
			PartNumber: i + 1,
			PartCount:  len(s.cfg.WorkerURLs),
			Alphabet:   domain.Alphabet{Symbols: alphabet},
		}

		body, err := xml.Marshal(task)
		if err != nil {
			log.Printf("Failed to marshal task for worker %d: %v", i+1, err)
			continue
		}

		taskURL := workerURL + WorkerTaskPath

		go func(url string, body []byte) {
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Failed creating http request for worker %s: %v", url, err)
				return
			}
			req.Header.Set("Content-Type", "application/xml")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to send task to worker at %s: %v", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Worker %s return non-200 status code: %d", url, resp.StatusCode)
			}

		}(taskURL, body)
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
	cancelReq := domain.WorkerCancelRequest{RequestID: reqID}

	body, err := xml.Marshal(cancelReq)
	if err != nil {
		log.Printf("Failed marshalling cancel request: %v", err)
		return
	}

	client := &http.Client{Timeout: s.cfg.ClientTimeout}

	for _, workerURL := range s.cfg.WorkerURLs {
		cancelURL := workerURL + WorkerCancelPath

		go func(url string) {
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Failed creating http request for worker %s: %v", url, err)
				return
			}
			req.Header.Set("Content-Type", "application/xml")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to send cancel task to %s: %v", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Worker %s return non-200 status code: %d", url, resp.StatusCode)
			}
		}(cancelURL)
	}
}
