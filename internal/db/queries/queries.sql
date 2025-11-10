-- name: CreateLocalUser :one
INSERT INTO users (
    ap_id,
    username,
    name,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    private_key,
    created,
    last_updated
    )
VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11) RETURNING id;

-- name: CreateAccount :exec
INSERT INTO 
    accounts (password, admin, email, user_id, created, last_updated)
VALUES 
    (?, ?, ?, ?, ?, ?);

-- name: AuthUserByEmail :one
SELECT
    u.id AS user_id,
    a.id AS account_id,
    u.username,
    a.password,
    a.admin
FROM users AS u
JOIN accounts AS a
ON a.user_id = u.id
WHERE u.local AND a.email = ?1
LIMIT 1;

-- name: AuthUserByUsername :one
SELECT
    u.id AS user_id,
    a.id AS account_id,
    u.username,
    a.password,
    a.admin
FROM users AS u
JOIN accounts AS a
ON a.user_id = u.id
WHERE u.local AND u.username = ?1
LIMIT 1;