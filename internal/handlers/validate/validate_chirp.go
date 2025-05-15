package validate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/lucashthiele/chirpy/pkg/response"
)

type params struct {
	Body string `json:"body"`
}

type responseData struct {
	Body string `json:"cleaned_body"`
}

func validateChirp(str string) error {
	if len(str) > 140 {
		return fmt.Errorf("chirp is too long")
	}

	return nil
}

func removeProfaneWords(str string) string {
	badWords := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}

	splittedWords := strings.Split(str, " ")
	for i, word := range splittedWords {
		if replace, found := badWords[strings.ToLower(word)]; found {
			splittedWords[i] = replace
		}
	}

	return strings.Join(splittedWords, " ")
}

func HandleValidateChirp(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := params{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding body: %s", err)
		response.RespondWithInternalServerError(res, err)
		return
	}

	err = validateChirp(params.Body)
	if err != nil {
		response.RespondWithError(res, http.StatusBadRequest, err.Error())
		return
	}

	cleanedBody := removeProfaneWords(params.Body)

	bodyResp := responseData{
		Body: cleanedBody,
	}

	response.RespondWithJSON(res, http.StatusOK, bodyResp)
}
