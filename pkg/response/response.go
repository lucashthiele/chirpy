package response

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lucashthiele/chirpy/internal/model"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, err := json.Marshal(payload)
	if err != nil {
		RespondWithInternalServerError(w, err)
		return
	}

	w.WriteHeader(code)
	w.Write(resp)
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	errResp := model.ErrorResponse{
		Error: msg,
	}

	resp, err := json.Marshal(errResp)
	if err != nil {
		RespondWithInternalServerError(w, err)
		return
	}

	w.WriteHeader(code)
	w.Write(resp)
}

func RespondWithInternalServerError(w http.ResponseWriter, err error) {
	log.Printf("Internal Server Error: %s", err)

	errResp := model.ErrorResponse{
		Error: "something went wrong",
	}

	resp, err := json.Marshal(errResp)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write(resp)
}
