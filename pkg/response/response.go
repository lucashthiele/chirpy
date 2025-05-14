package response

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lucashthiele/chirpy/internal/model"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	resp, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)

	errResp := model.ErrorResponse{
		Error: msg,
	}
	resp, err := json.Marshal(errResp)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}
