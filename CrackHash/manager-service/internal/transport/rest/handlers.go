package rest

import (
	"crackhash/manager/internal/service"
	"crackhash/pkg/domain"
	"encoding/json"
	"encoding/xml"
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
	if r.Method != http.MethodPost {
		http.Error(w, ErrMsgNonPostMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var req domain.CrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
		return
	}

	reqID, err := h.svc.CreateTask(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.CrackResponse{RequestID: reqID})
}

func (h *Handler) handleCrackStatusCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, ErrMsgNonGetMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	reqID := r.URL.Query().Get("requestId")
	if reqID == "" {
		http.Error(w, ErrMsgMissingReqID, http.StatusBadRequest)
		return
	}

	resp, exists := h.svc.GetTaskStatus(r.Context(), reqID)
	if !exists {
		http.Error(w, ErrMsgReqNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleWorkerResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, ErrMsgNonPatchMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var resp domain.WorkerResponse
	if err := xml.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, ErrMsgInvalidJSON, http.StatusBadRequest)
		return
	}

	err := h.svc.ProcessWorkerResult(r.Context(), resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
