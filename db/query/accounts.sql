-- name: CreateAccount :one
INSERT INTO accounts (
  owner, 
  balance, 
  currency
) VALUES (
  $1, $2, $3
)
RETURNING *; 
/* '*' so that id is also returned, along with other important details */

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: ListAccount :many
SELECT * FROM accounts
ORDER BY owner
LIMIT $1
OFFSET $2;

/* exec since it doesn't return any data */
-- name: UpdateAccount :exec
UPDATE accounts
set balance = $2
WHERE id = $1;