package main

import (
	"bytes"
	"crackhash/pkg/models"
	"encoding/xml"
	"log"
	"net/http"
	"time"
)

var alphabet = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
	"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

func (a *ManagerApp) dispatchTasks(reqID string, req models.CrackRequest) {
	client := &http.Client{Timeout: a.cfg.ClientTimeout}

	for i, workerURL := range a.cfg.WorkerURLs {
		task := models.WorkerTask{
			RequestID:  reqID,
			Hash:       req.Hash,
			MaxLength:  req.MaxLength,
			PartNumber: i + 1,
			PartCount:  len(a.cfg.WorkerURLs),
			Alphabet:   models.Alphabet{Symbols: alphabet},
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

	body, err := xml.Marshal(cancelReq)
	if err != nil {
		log.Printf("Failed marshalling cancel request: %v", err)
		return
	}

	client := &http.Client{Timeout: a.cfg.ClientTimeout}

	for _, workerURL := range a.cfg.WorkerURLs {
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
