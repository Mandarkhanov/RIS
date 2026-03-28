package main

import (
	"bytes"
	"context"
	"crackhash/pkg/generator"
	"crackhash/pkg/models"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"log"
	"net/http"
	"strings"
)

func (a *WorkerApp) processTask(ctx context.Context, task models.WorkerTask) {
	defer func() {
		a.mu.Lock()
		if cancel, exists := a.activeTasks[task.RequestID]; exists {
			cancel()
			delete(a.activeTasks, task.RequestID)
		}
		a.mu.Unlock()
	}()

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
		a.sendResultToManager(task.RequestID, task.PartNumber, nil)
		return
	}

	targetHashBytes, err := hex.DecodeString(task.Hash)
	if err != nil {
		log.Printf("Worker %d error decoding target hash: %v", task.PartNumber, err)
		a.sendResultToManager(task.RequestID, task.PartNumber, nil)
		return
	}

	log.Printf("Worker %d START processing task [%d, %d)\n", task.PartNumber, startIdx, endIdx)
	gen := generator.NewGenerator(alphabetBytes, startIdx)
	wordBuf := make([]byte, len(gen.State))

	for i := startIdx; i < endIdx; i++ {
		if i%a.cfg.ContextCheckInterval == 0 && ctx.Err() != nil {
			log.Printf("Worker STOPPED task %s by Manager request", task.RequestID)
			if len(foundWords) > 0 {
				a.sendResultToManager(task.RequestID, task.PartNumber, foundWords)
			}
			return
		}

		wordBuf = gen.CurrentWordBytes(wordBuf)
		hash := md5.Sum(wordBuf)

		if bytes.Equal(hash[:], targetHashBytes) {
			foundWords = append(foundWords, string(wordBuf))
			log.Printf("Worker %d FOUND match: %s\n", task.PartNumber, foundWords[len(foundWords)-1])
			if a.cfg.StopOnFirstMatch {
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

	a.sendResultToManager(task.RequestID, task.PartNumber, foundWords)
}

func (a *WorkerApp) sendResultToManager(reqID string, partNumber int, foundWords []string) {
	resp := models.WorkerResponse{
		RequestID:  reqID,
		PartNumber: partNumber,
		Answers:    models.Answers{Words: foundWords},
	}

	body, err := xml.Marshal(resp)
	if err != nil {
		log.Printf("Worker error marshalling XML: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPatch, a.cfg.ManagerURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Worker error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{Timeout: a.cfg.ClientTimeout}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Worker error sending result to manager: %v", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("Manager returned non-200 status code: %d", res.StatusCode)
	} else {
		log.Printf("Worker successfully sent results for task %s", reqID)
	}
}
