package rest

import (
	"crackhash/pkg/domain"
	"crackhash/worker/internal/service"
	"encoding/xml"
	"net/http"
)

const (
	ErrMsgNonPostMethodNotAllowed = "Non-POST method not allowed"

	ErrMsgInvalidXML = "Invalid XML body"
)

const (
	WorkerTaskPath   = "/internal/api/worker/hash/crack/task"
	WorkerCancelPath = "/internal/api/worker/hash/crack/cancel"
)

type Handler struct {
	svc *service.CrackWorkerService
}

func NewHandler(svc *service.CrackWorkerService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(WorkerTaskPath, h.handleTask)
	mux.HandleFunc(WorkerCancelPath, h.handleCancel)
	return mux
}

func (h *Handler) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMsgNonPostMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var task domain.WorkerTask
	if err := xml.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, ErrMsgInvalidXML, http.StatusBadRequest)
		return
	}

	h.svc.StartTask(task)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMsgNonPostMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var req domain.WorkerCancelRequest
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrMsgInvalidXML, http.StatusBadRequest)
		return
	}

	h.svc.CancelTask(req.RequestID)
	w.WriteHeader(http.StatusOK)
}
