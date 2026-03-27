package main

import (
	"bytes"
	"crackhash/pkg/models"
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"
)

var alphabet = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
	"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

func (a *ManagerApp) dispatchTasks(reqID string, req models.CrackRequest) {
	for i, url := range a.cfg.WorkersURLs {
		task := models.WorkerTask{
			RequestID:  reqID,
			Hash:       req.Hash,
			MaxLength:  req.MaxLength,
			PartNumber: i + 1,
			PartCount:  len(a.cfg.WorkersURLs),
			Alphabet:   models.Alphabet{Symbols: alphabet},
		}

		body, err := xml.Marshal(task)
		if err != nil {
			log.Printf("Failed to marshal task for worker %d: %v", i+1, err)
			continue
		}

		go func(workerURL string, body []byte) {
			resp, err := http.Post(workerURL, "application/xml", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Failed to send task to worker at %s: %v", workerURL, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Worker %s return non-200 status code: %d", workerURL, resp.StatusCode)
			}

		}(url, body)
	}
}

func (a *ManagerApp) timeoutWatcher() {
	for {
		time.Sleep(a.cfg.WatchInterval)
		now := time.Now()

		var tasksToCancel []string

		a.storage.mu.RLock()
		for key, state := range a.storage.tasks {
			state.mu.Lock()
			if state.Status == models.StatusInProgress && now.Sub(state.CreatedAt) > a.cfg.TaskTimeout {
				if len(state.Data) > 0 {
					state.Status = models.StatusPartialReady
					log.Printf("Task %s timed out. Status set to PARTIAL (has data).", key)
				} else {
					state.Status = models.StatusError
					log.Printf("Task %s timed out. Status set to ERROR.", key)
				}
				tasksToCancel = append(tasksToCancel, key)
			}
			state.mu.Unlock()
		}
		a.storage.mu.RUnlock()

		for _, reqID := range tasksToCancel {
			a.cancelTask(reqID)
		}
	}
}

func (a *ManagerApp) cancelTask(reqID string) {
	cancelReq := models.WorkerCancelRequest{RequestID: reqID}
	body, _ := xml.Marshal(cancelReq)

	for _, taskURL := range a.cfg.WorkersURLs {
		cancelURL := strings.Replace(taskURL, "/task", "/cancel", 1)

		go func(url string) {
			resp, err := http.Post(url, "application/xml", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Failed to cancel task to worker at %s: %v", taskURL, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Worker %s return non-200 status code: %d", taskURL, resp.StatusCode)
			}
		}(cancelURL)
	}
}
