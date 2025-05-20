-- name: CreateRefreshToken :one
INSERT INTO REFRESH_TOKEN (
  TOKEN,
  CREATED_AT,
  UPDATED_AT,
  USER_ID,
  EXPIRES_AT,
  REVOKED_AT
) VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  $3,
  $4
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT USER_ID
  FROM REFRESH_TOKEN
 WHERE TOKEN = $1
   AND EXPIRES_AT > NOW()
   AND REVOKED_AT IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE REFRESH_TOKEN
   SET REVOKED_AT = NOW(),
       UPDATED_AT = NOW()
 WHERE TOKEN = $1;