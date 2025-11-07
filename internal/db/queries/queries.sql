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
    private_key
    )
VALUES ($1,$2,$3,$4,@inbox::text,@outbox::text,@followers::text,@publicKey::text, @privateKey::text) --Keys must not be null!
RETURNING id;

-- name: CreateAccount :exec
INSERT INTO accounts (password, admin, email, user_id)
VALUES ($1,$2,$3,$4);

-- name: GetAuthData :one
SELECT
    u.id as user_id,
    u.ap_id as ap_id,
    a.id as account_id,
    a.password,
    a.admin
FROM users as u
JOIN accounts as a on u.id = a.user_id
WHERE u.username = @username::text AND u.local;
-- where NOT a.blocked