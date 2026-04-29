package rest

import (
	"crackhash/pkg/domain"
	"crackhash/pkg/util"
	"crackhash/worker/internal/service"
	"encoding/xml"
	"log"
	"net/http"
)

const (
	ErrMsgNonPostMethodNotAllowed = "Non-POST method not allowed"

	ErrMsgInvalidXML = "Invalid XML body"
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
	mux.HandleFunc("/internal/api/worker/hash/crack/task", h.handleTask)
	mux.HandleFunc("/internal/api/worker/hash/crack/cancel", h.handleCancel)
	return mux
}

func (h *Handler) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Failed to handleTask: %s", ErrMsgNonPostMethodNotAllowed)
		util.RespondWithXMLError(w, http.StatusMethodNotAllowed, ErrMsgNonPostMethodNotAllowed)
		return
	}

	var task domain.WorkerTask
	if err := xml.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("Failed to handleTask: %v", err)
		util.RespondWithXMLError(w, http.StatusBadRequest, ErrMsgInvalidXML)
		return
	}

	h.svc.StartTask(task)

	log.Printf("Started task [%s]", task.RequestID)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Failed to handleCancel: %s", ErrMsgNonPostMethodNotAllowed)
		util.RespondWithXMLError(w, http.StatusMethodNotAllowed, ErrMsgNonPostMethodNotAllowed)
		return
	}

	var req domain.WorkerCancelRequest
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to handleCancel: %v", err)
		util.RespondWithXMLError(w, http.StatusBadRequest, ErrMsgInvalidXML)
		return
	}

	h.svc.CancelTask(req.RequestID)

	log.Printf("Canceled task [%s]", req.RequestID)
	w.WriteHeader(http.StatusOK)
}
