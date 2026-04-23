package service

import (
	"bytes"
	"context"
	"crackhash/pkg/domain"
	"crackhash/worker/internal/config"
	"crackhash/worker/internal/generator"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"sync"
)

const ManagerRequestPath = "/internal/api/manager/hash/crack/request"

type CrackWorkerService struct {
	cfg         *config.Config
	activeTasks map[string]context.CancelFunc
	mu          sync.Mutex
}

func NewCrackWorkerService(cfg *config.Config) *CrackWorkerService {
	return &CrackWorkerService{
		cfg:         cfg,
		activeTasks: make(map[string]context.CancelFunc),
	}
}

func (s *CrackWorkerService) StartTask(task domain.WorkerTask) {
	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.activeTasks[task.RequestID] = cancel
	s.mu.Unlock()

	go s.processTask(ctx, task)
}

func (s *CrackWorkerService) CancelTask(reqID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancelFunc, exists := s.activeTasks[reqID]; exists {
		cancelFunc()
		delete(s.activeTasks, reqID)
	}
}

func (s *CrackWorkerService) processTask(ctx context.Context, task domain.WorkerTask) {
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

	// Воркеров больше количества комбинаций
	if startIdx >= totalCombinations {
		log.Printf("Worker %d has no work (startIdx >= totalCombinations)", task.PartNumber)
		s.sendResultToManager(task.RequestID, task.PartNumber, nil)
		return
	}

	targetHashBytes, err := hex.DecodeString(task.Hash)
	if err != nil {
		log.Printf("Worker %d failed decoding target hash: %v", task.PartNumber, err)
		s.sendResultToManager(task.RequestID, task.PartNumber, nil)
		return
	}

	log.Printf("Worker %d STARTED processing task [%d, %d)\n", task.PartNumber, startIdx, endIdx)
	gen := generator.NewGenerator(alphabetBytes, startIdx)
	wordBuf := make([]byte, len(gen.State))

	for i := startIdx; i < endIdx; i++ {
		if i%int64(s.cfg.ContextCheckInterations) == 0 && ctx.Err() != nil {
			log.Printf("Worker %d STOPPED task %s by Manager request", task.PartNumber, task.RequestID)
			if len(foundWords) > 0 {
				s.sendResultToManager(task.RequestID, task.PartNumber, foundWords)
			}
			return
		}

		wordBuf = gen.CurrentWordBytes(wordBuf)
		hash := md5.Sum(wordBuf)

		if bytes.Equal(hash[:], targetHashBytes) {
			foundWords = append(foundWords, string(wordBuf))
			log.Printf("Worker %d FOUND match: %s\n", task.PartNumber, foundWords[len(foundWords)-1])
			if s.cfg.StopOnFirstMatch {
				break
			}
		}

		gen.NextState()
	}

	if len(foundWords) > 0 {
		log.Printf("Worker %d FINISHED. Found %d words\n", task.PartNumber, len(foundWords))
	} else {
		log.Printf("Worker %d FINISHED. No words found.\n", task.PartNumber)
	}

	s.sendResultToManager(task.RequestID, task.PartNumber, foundWords)
}

func (s *CrackWorkerService) sendResultToManager(reqID string, partNumber int, foundWords []string) {
	resp := domain.WorkerResponse{
		RequestID:  reqID,
		PartNumber: partNumber,
		Answers:    domain.Answers{Words: foundWords},
	}

	body, err := xml.Marshal(resp)
	if err != nil {
		log.Printf("Worker failed marshalling XML: %v", err)
		return
	}

	managerRequestURL := s.cfg.ManagerURL + ManagerRequestPath
	req, err := http.NewRequest(http.MethodPatch, managerRequestURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Worker failed creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{Timeout: s.cfg.ClientTimeout}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Worker failed sending result to manager: %v", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("Manager returned non-200 status code: %d", res.StatusCode)
	} else {
		log.Printf("Worker successfully sent results for task %s", reqID)
	}
}
