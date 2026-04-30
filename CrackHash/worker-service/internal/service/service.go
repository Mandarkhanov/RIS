package service

import (
	"bytes"
	"context"
	"crackhash/pkg/domain"
	"crackhash/pkg/rabbitmq"
	"crackhash/worker/internal/config"
	"crackhash/worker/internal/generator"
	"crackhash/worker/internal/repository"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"log"
	"strings"
)

const ManagerRequestPath = "/internal/api/manager/hash/crack/request"

type CrackWorkerService struct {
	cfg             *config.Config
	activeTasksRepo *repository.ActiveTaskRepository
	rmqClient       *rabbitmq.Client
}

func NewCrackWorkerService(cfg *config.Config, activeTaskRepo *repository.ActiveTaskRepository, rmqClient *rabbitmq.Client) *CrackWorkerService {
	return &CrackWorkerService{
		cfg:             cfg,
		activeTasksRepo: activeTaskRepo,
		rmqClient:       rmqClient,
	}
}

func (s *CrackWorkerService) CancelTask(reqID string) {
	s.activeTasksRepo.Cancel(reqID)
}

func (s *CrackWorkerService) ProcessTask(ctx context.Context, task domain.WorkerTask) {
	childCtx, cancel := context.WithCancel(ctx)
	added, cancelled := s.activeTasksRepo.Add(task.RequestID, cancel)
	if cancelled {
		cancel()
		log.Printf("Task [%s] was previously cancelled. Skipping part %d.", task.RequestID, task.PartNumber)
		return
	}
	if !added {
		cancel()
		log.Printf("Task [%s] (Part %d) is already running.", task.RequestID, task.PartNumber)
		return
	}
	defer s.CancelTask(task.RequestID)

	var foundWords []string

	alphabetStr := strings.Join(task.Alphabet.Symbols, "")
	alphabetBytes := []byte(alphabetStr)
	aLen := len(alphabetBytes)
	totalCombinations := generator.TotalWords(aLen, task.MaxLength)

	chunkSize := (totalCombinations + int64(task.PartCount) - 1) / int64(task.PartCount)
	startIdx := int64(task.PartNumber-1) * chunkSize
	endIdx := startIdx + chunkSize
	if endIdx > totalCombinations {
		endIdx = totalCombinations
	}

	if startIdx >= totalCombinations {
		log.Println("Has no work (startIdx >= totalCombinations)")
		s.sendResultToRMQ(childCtx, task.RequestID, task.PartNumber, nil)
		return
	}

	targetHashBytes, err := hex.DecodeString(task.Hash)
	if err != nil {
		log.Printf("Failed decoding target hash: %v", err)
		s.sendResultToRMQ(childCtx, task.RequestID, task.PartNumber, nil)
		return
	}

	log.Printf("Started processing task [%s] part %d [%d, %d)\n", task.RequestID, task.PartNumber, startIdx, endIdx)
	gen := generator.NewGenerator(alphabetBytes, startIdx)
	wordBuf := make([]byte, len(gen.State))

	for i := startIdx; i < endIdx; i++ {
		if i%int64(s.cfg.ContextCheckIterations) == 0 && childCtx.Err() != nil {
			log.Printf("Stopped task %s", task.RequestID)
			if len(foundWords) > 0 {
				s.sendResultToRMQ(context.Background(), task.RequestID, task.PartNumber, foundWords)
			}
			return
		}

		wordBuf = gen.CurrentWordBytes(wordBuf)
		hash := md5.Sum(wordBuf)

		if bytes.Equal(hash[:], targetHashBytes) {
			foundWords = append(foundWords, string(wordBuf))
			log.Printf("Found match: %s\n", foundWords[len(foundWords)-1])
			if s.cfg.StopOnFirstMatch {
				break
			}
		}

		gen.NextState()
	}

	log.Printf("Finished task. Found %d words\n", len(foundWords))

	s.sendResultToRMQ(context.Background(), task.RequestID, task.PartNumber, foundWords)
}

func (s *CrackWorkerService) sendResultToRMQ(ctx context.Context, reqID string, partNumber int, foundWords []string) {
	resp := domain.WorkerResponse{
		RequestID:  reqID,
		PartNumber: partNumber,
		Answers:    domain.Answers{Words: foundWords},
	}

	body, err := xml.Marshal(resp)
	if err != nil {
		log.Printf("Failed marshalling XML: %v", err)
		return
	}

	err = s.rmqClient.PublishXML(ctx, s.cfg.ResultsQueue, body)
	if err != nil {
		log.Printf("Failed sending result to RMQ: %v", err)
		return
	}

	log.Printf("Successfully published results for task %s to RMQ", reqID)
}
