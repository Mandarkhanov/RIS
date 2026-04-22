package rest

import (
	"crackhash/manager/internal/service"
	"net/http"
)

const (
	ErrMsgNonPostMethodNotAllowed  = "Non-POST method not allowed"
	ErrMsgNonGetMethodNotAllowed   = "Non-GET method not allowed"
	ErrMsgNonPatchMethodNotAllowed = "Non-PATCH method not allowed"

	ErrMsgMissingReqID = "Missing requestId parameter"
	ErrMsgReqNotFound  = "RequestNotFound"

	ErrMsgInvalidJSON = "Invalid JSON body"
)

const (
	WorkerTaskPath   = "/internal/api/worker/hash/crack/task"
	WorkerCancelPath = "/internal/api/worker/hash/crack/cancel"
)

type Handler struct {
	svc *service.CrackManagerService
}

func NewHandler(svc *service.CrackManagerService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/hash/crack", h.handleCrackRequest)
	mux.HandleFunc("/api/hash/status", h.handleCrackStatusCheck)
	mux.HandleFunc("/internal/api/manager/hash/crack/request", h.handleWorkerResponse)
	return mux
}

func (h *Handler) handleCrackRequest(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, ErrMsgNonPostMethodNotAllowed, http.StatusMethodNotAllowed)
	// 	return
	// }

	// var req domain.CrackRequest
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
	// 	return
	// }

	// reqID := uuid.New().String()
	// newState := &repository.RequestState{
	// 	ReqID:           reqID,
	// 	Hash:            req.Hash,
	// 	MaxLength:       req.MaxLength,
	// 	Status:          domain.StatusInProgress,
	// 	Data:            []string{},
	// 	WorkersFinished: 0,
	// 	TotalWorkers:    len(h.cfg.WorkerURLs),
	// 	CreatedAt:       time.Now(),
	// }

	// finalReqID, err := h.storage.StoreOrGetExists(reqID, newState)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusTooManyRequests)
	// 	return
	// }

	// if finalReqID == reqID {
	// 	go h.svc.DispatchTasks(reqID, req)
	// } else {
	// 	log.Printf("Task with [hash:%s | maxLength:%d] already exists with ID: %s", req.Hash, req.MaxLength, finalReqID)
	// }

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(domain.CrackResponse{RequestID: finalReqID})
}

func (h *Handler) handleCrackStatusCheck(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodGet {
	// 	http.Error(w, ErrMsgNonGetMethodNotAllowed, http.StatusMethodNotAllowed)
	// 	return
	// }

	// reqID := r.URL.Query().Get("requestId")
	// if reqID == "" {
	// 	http.Error(w, ErrMsgMissingReqID, http.StatusBadRequest)
	// 	return
	// }

	// state, exists := h.storage.Get(reqID)
	// if !exists {
	// 	http.Error(w, ErrMsgReqNotFound, http.StatusNotFound)
	// 	return
	// }

	// state.mu.Lock()
	// defer state.mu.Unlock()

	// var responseData []string
	// if state.Status == domain.StatusReady || state.Status == domain.StatusPartialReady {
	// 	responseData = state.Data
	// }

	// resp := domain.StatusResponse{
	// 	Status: state.Status,
	// 	Data:   responseData,
	// }

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleWorkerResponse(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPatch {
	// 	http.Error(w, ErrMsgNonPatchMethodNotAllowed, http.StatusMethodNotAllowed)
	// 	return
	// }

	// var resp domain.WorkerResponse
	// if err := xml.NewDecoder(r.Body).Decode(&resp); err != nil {
	// 	http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
	// 	return
	// }

	// state, exists := h.storage.Get(resp.RequestID)
	// if !exists {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	// state.mu.Lock()
	// defer state.mu.Unlock()
	// if state.Status == domain.StatusReady {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	// if len(resp.Answers.Words) > 0 {
	// 	state.Data = append(state.Data, resp.Answers.Words...)
	// 	if state.Status == domain.StatusError {
	// 		state.Status = domain.StatusPartialReady
	// 	}
	// }
	// state.WorkersFinished++

	// if state.WorkersFinished == state.TotalWorkers {
	// 	state.Status = domain.StatusReady
	// 	log.Printf("Task %s is READY. Found %d words.", resp.RequestID, len(state.Data))
	// }
	// w.WriteHeader(http.StatusOK)
}
