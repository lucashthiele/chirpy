package chirps

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucashthiele/chirpy/internal/config"
	"github.com/lucashthiele/chirpy/internal/database"
	"github.com/lucashthiele/chirpy/pkg/parser"
	"github.com/lucashthiele/chirpy/pkg/response"
)

type params struct {
	Body string `json:"body"`
}

type responseData struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    string    `json:"user_id"`
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

func HandleCreateChirp(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	params := &params{}

	err = parser.ParseBody(req.Body, params)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	userId, ok := req.Context().Value(config.UserIDKey).(uuid.UUID)
	if !ok {
		response.RespondWithInternalServerError(res, fmt.Errorf("omg you're so bad at this"))
	}

	err = validateChirp(params.Body)
	if err != nil {
		response.RespondWithError(res, http.StatusBadRequest, err.Error())
		return
	}

	cleanedBody := removeProfaneWords(params.Body)

	chirp := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	}

	createdChirp, err := cfg.Db.CreateChirp(req.Context(), chirp)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	bodyResp := responseData{
		Id:        createdChirp.ID.String(),
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body:      createdChirp.Body,
		UserId:    createdChirp.UserID.String(),
	}

	response.RespondWithJSON(res, http.StatusCreated, bodyResp)
}

func HandleGetAllChirps(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	chirps, err := cfg.Db.ListAllChirps(req.Context())
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	bodyResp := make([]responseData, len(chirps))

	for i, chirp := range chirps {
		bodyResp[i] = responseData{
			Id:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID.String(),
		}
	}

	response.RespondWithJSON(res, http.StatusOK, bodyResp)
}

func HandlerGetChirpByID(res http.ResponseWriter, req *http.Request) {
	cfg, err := config.New()
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	chirpID := req.PathValue("chirpID")

	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		response.RespondWithInternalServerError(res, err)
	}

	chirp, err := cfg.Db.GetChirpByID(req.Context(), chirpUUID)
	if err == sql.ErrNoRows {
		response.RespondWithError(res, http.StatusNotFound, "Not found")
		return
	}
	if err != nil {
		response.RespondWithInternalServerError(res, err)
		return
	}

	bodyResp := responseData{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
	}

	response.RespondWithJSON(res, http.StatusOK, bodyResp)
}
