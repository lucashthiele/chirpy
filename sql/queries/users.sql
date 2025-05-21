-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING ID, CREATED_AT, UPDATED_AT, EMAIL, IS_CHIRPY_RED;

-- name: DeleteAllUsers :exec
DELETE FROM USERS;

-- name: GetUserByEmail :one
SELECT ID,
       CREATED_AT,
       UPDATED_AT,
       EMAIL,
       HASHED_PASSWORD,
       IS_CHIRPY_RED
  FROM USERS
 WHERE EMAIL = $1;

-- name: UpdateUserEmailAndPassword :one
UPDATE USERS
   SET EMAIL = $1,
       HASHED_PASSWORD = $2
 WHERE ID = $3
 RETURNING ID, CREATED_AT, UPDATED_AT, EMAIL, IS_CHIRPY_RED;

-- name: UpgradeUser :exec
UPDATE USERS
   SET IS_CHIRPY_RED = TRUE
 WHERE ID = $1;