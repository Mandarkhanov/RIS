package main

import (
	"crackhash/pkg/models"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	ErrMsgNonPostMethodNotAllowed  = "Non-POST method not allowed"
	ErrMsgNonGetMethodNotAllowed   = "Non-GET method not allowed"
	ErrMsgNonPatchMethodNotAllowed = "Non-PATCH method not allowed"

	ErrMsgMissingReqID = "Missing requestId parameter"
	ErrMsgReqNotFound  = "RequestNotFound"

	ErrMsgInvalidJSON = "Invalid JSON body"
)

func (a *ManagerApp) handleCrackRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMsgNonPostMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var req models.CrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
		return
	}

	reqID := uuid.New().String()
	newState := &RequestState{
		ReqID:           reqID,
		Hash:            req.Hash,
		MaxLength:       req.MaxLength,
		Status:          models.StatusInProgress,
		Data:            []string{},
		WorkersFinished: 0,
		TotalWorkers:    len(a.cfg.WorkerURLs),
		CreatedAt:       time.Now(),
	}

	finalReqID, err := a.storage.StoreOrGetExists(reqID, newState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	if finalReqID == reqID {
		go a.dispatchTasks(reqID, req)
	} else {
		log.Printf("Task with [hash:%s | maxLength:%d] already exists with ID: %s", req.Hash, req.MaxLength, finalReqID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CrackResponse{RequestID: finalReqID})
}

func (a *ManagerApp) handleCrackStatusCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, ErrMsgNonGetMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	reqID := r.URL.Query().Get("requestId")
	if reqID == "" {
		http.Error(w, ErrMsgMissingReqID, http.StatusBadRequest)
		return
	}

	state, exists := a.storage.Get(reqID)
	if !exists {
		http.Error(w, ErrMsgReqNotFound, http.StatusNotFound)
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	var responseData []string
	if state.Status == models.StatusReady || state.Status == models.StatusPartialReady {
		responseData = state.Data
	}

	resp := models.StatusResponse{
		Status: state.Status,
		Data:   responseData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *ManagerApp) handleWorkerResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, ErrMsgNonPatchMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var resp models.WorkerResponse
	if err := xml.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
		return
	}

	state, exists := a.storage.Get(resp.RequestID)
	if !exists {
		w.WriteHeader(http.StatusOK)
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	if state.Status == models.StatusReady {
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(resp.Answers.Words) > 0 {
		state.Data = append(state.Data, resp.Answers.Words...)
		if state.Status == models.StatusError {
			state.Status = models.StatusPartialReady
		}
	}
	state.WorkersFinished++

	if state.WorkersFinished == state.TotalWorkers {
		state.Status = models.StatusReady
		log.Printf("Task %s is READY. Found %d words.", resp.RequestID, len(state.Data))
	}
	w.WriteHeader(http.StatusOK)
}
