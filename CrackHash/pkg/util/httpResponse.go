package util

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func RespondWithJSONError(w http.ResponseWriter, code int, message string) {
	payload := map[string]string{
		"error": message,
	}
	RespondWithJSON(w, code, payload)
}

func RespondWithXML(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	if err := xml.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func RespondWithXMLError(w http.ResponseWriter, code int, message string) {
	payload := struct {
		XMLName xml.Name `xml:"error"`
		Message string   `xml:"message"`
	}{
		Message: message,
	}
	RespondWithXML(w, code, payload)
}
