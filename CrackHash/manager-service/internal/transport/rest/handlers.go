package rest

import (
	"crackhash/manager/internal/service"
	"crackhash/pkg/domain"
	"crackhash/pkg/util"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
)

const (
	ErrMsgNonPostMethodNotAllowed  = "Non-POST method not allowed"
	ErrMsgNonGetMethodNotAllowed   = "Non-GET method not allowed"
	ErrMsgNonPatchMethodNotAllowed = "Non-PATCH method not allowed"

	ErrMsgMissingReqID = "Missing requestId parameter"
	ErrMsgReqNotFound  = "RequestNotFound"

	ErrMsgInvalidJSON         = "Invalid JSON body"
	ErrMsgInvalidXML          = "Invalid XML body"
	ErrMsgInternalServerError = "Internal server error"
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
		log.Printf("Failed to handleCrackRequest: %s", ErrMsgNonPostMethodNotAllowed)
		util.RespondWithJSONError(w, http.StatusMethodNotAllowed, ErrMsgNonPostMethodNotAllowed)
		return
	}

	var req domain.CrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed decode JSON: %v", err)
		util.RespondWithJSONError(w, http.StatusBadRequest, ErrMsgInvalidJSON)
		return
	}

	reqID, err := h.svc.CreateTask(r.Context(), req)
	if err != nil {
		log.Printf("Failed to CreateTask: %v", err)
		util.RespondWithJSONError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
		return
	}

	log.Printf("Created task [%s]", reqID)
	util.RespondWithJSON(w, http.StatusOK, domain.CrackResponse{RequestID: reqID})
}

func (h *Handler) handleCrackStatusCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Failed to handleCrackStatusCheck: %s", ErrMsgNonGetMethodNotAllowed)
		util.RespondWithJSONError(w, http.StatusMethodNotAllowed, ErrMsgNonGetMethodNotAllowed)
		return
	}

	reqID := r.URL.Query().Get("requestId")
	if reqID == "" {
		log.Printf("Failed to check status of task [%s]: %s", reqID, ErrMsgMissingReqID)
		util.RespondWithJSONError(w, http.StatusBadRequest, ErrMsgMissingReqID)
		return
	}

	resp, exists := h.svc.GetTaskStatus(r.Context(), reqID)
	if !exists {
		log.Printf("Failed to handleCrackStatusCheck: %s", ErrMsgReqNotFound)
		util.RespondWithJSONError(w, http.StatusNotFound, ErrMsgReqNotFound)
		return
	}

	log.Printf("Returned task [%s] status", reqID)
	util.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleWorkerResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		log.Printf("Failed to handleWorkerResponse: %s", ErrMsgNonPatchMethodNotAllowed)
		util.RespondWithXMLError(w, http.StatusMethodNotAllowed, ErrMsgNonPatchMethodNotAllowed)
		return
	}

	var resp domain.WorkerResponse
	if err := xml.NewDecoder(r.Body).Decode(&resp); err != nil {
		log.Printf("Failed to handleWorkerResponse: %v", err)
		util.RespondWithXMLError(w, http.StatusBadRequest, ErrMsgInvalidXML)
		return
	}

	err := h.svc.ProcessWorkerResult(r.Context(), resp)
	if err != nil {
		log.Printf("Failed to handleWorkerResponse: %v", err)
		util.RespondWithXMLError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
		return
	}

	log.Printf("Processed worker response about task [%s]", resp.RequestID)
	w.WriteHeader(http.StatusOK)
}
