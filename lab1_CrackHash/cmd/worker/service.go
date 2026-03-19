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
	alphabetRune := []rune(alphabetStr)
	totalCombinations := generator.CalculateTotalWords(len(alphabetRune), task.MaxLength)

	chunkSize := totalCombinations / int64(task.PartCount)
	startIdx := int64(task.PartNumber-1) * chunkSize
	var endIdx int64
	if task.PartCount == task.PartNumber {
		endIdx = totalCombinations
	} else {
		endIdx = startIdx + chunkSize
	}

	log.Printf("Worker processing task [%s, %s)",
		generator.GetWordAtIndex(startIdx, alphabetRune),
		generator.GetWordAtIndex(endIdx, alphabetRune),
	)

	for i := startIdx; i < endIdx; i++ {
		if ctx.Err() != nil {
			log.Printf("Worker STOPPED task %s by Manager request", task.RequestID)

			if len(foundWords) > 0 {
				a.sendResultToManager(task.RequestID, task.PartNumber, foundWords)
			}
			return
		}

		word := generator.GetWordAtIndex(i, alphabetRune)
		hash := md5.Sum([]byte(word))
		hashStr := hex.EncodeToString(hash[:])

		if hashStr == task.Hash {
			foundWords = append(foundWords, word)
		}
	}

	a.sendResultToManager(task.RequestID, task.PartNumber, foundWords)
}

func (a *WorkerApp) sendResultToManager(reqID string, partNumber int, foundWords []string) {
	resp := models.WorkerResponse{
		RequestID:  reqID,
		PartNumber: partNumber,
		Answers: models.Answers{
			Words: foundWords,
		},
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

	client := &http.Client{}
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
