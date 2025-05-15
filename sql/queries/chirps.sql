-- name: CreateChirp :one
INSERT INTO CHIRPS(
  ID,
  CREATED_AT,
  UPDATED_AT,
  BODY,
  USER_ID
) VALUES (
  GEN_RANDOM_UUID(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;
