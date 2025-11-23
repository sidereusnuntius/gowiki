-- name: CreateLocalUser :one
INSERT INTO users (
    ap_id,
    username,
    name,
    trusted,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    private_key
    )
VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10) RETURNING id;

-- name: CreateAccount :exec
INSERT INTO 
    accounts (password, admin, email, user_id)
VALUES 
    (?, ?, ?, ?);

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

-- name: GetLocalArticleByTitle :one
SELECT
    title,
    summary,
    content,
    protected,
    media_type,
    language
FROM
    articles
where local AND title = ?1
LIMIT 1;

-- name: IsUserTrusted :one
SELECT trusted FROM users where id = ?1 LIMIT 1;

-- name: CreateArticle :one
INSERT INTO articles (
    ap_id,
    url,
    instance_id,
    language,
    media_type,
    title,
    content
) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: EditArticle :one
INSERT INTO revisions (
    ap_id,
    article_id,
    user_id,
    summary,
    diff,
    reviewed,
    published,
    prev,
    based_on
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: GetArticleIDS :one
SELECT
    a.ap_id,
    a.id AS article_id,
    r.id AS rev_id
FROM articles a JOIN revisions r ON r.article_id = a.id
WHERE lower(a.title) = lower(@title)
ORDER BY r.created DESC
LIMIT 1;

-- name: InsertRevision :exec
INSERT INTO revisions (
    ap_id,
    article_id,
    user_id,
    summary,
    diff,
    published,
    prev
) VALUES (?1, ?2, ?3, ?4, ?5, true, ?6);

-- name: UpdateArticle :exec
UPDATE articles
SET
    content = ?1,
    last_updated = (cast(strftime('%s','now') as int))
WHERE id = ?2;

-- name: GetArticleContent :one
SELECT content FROM articles WHERE id = ?;

-- name: GetRevisionList :many
SELECT
    r.id,
    r.reviewed,
    r.summary,
    a.title,
    u.username,
    r.created
FROM (
    SELECT id, title from articles WHERE lower(title) = lower(@title) LIMIT 1
) a
JOIN revisions r ON r.article_id = a.id
JOIN users u ON r.user_id = u.id
ORDER BY r.created DESC;

-- name: GetLocalUserData :one
SELECT
    id,
    username,
    name,
    url,
    summary
FROM users
WHERE local AND username = lower(?);

-- name: GetForeignUserData :one
SELECT
    id,
    username,
    name,
    domain,
    url,
    local,
    summary
FROM users
WHERE username = lower(?) AND NOT local AND domain = ?;

-- name: GetRevisionsByUserId :many
SELECT
    r.id,
    a.title,
    r.summary,
    r.reviewed,
    r.created
FROM revisions r
JOIN articles a
ON a.id = r.article_id
WHERE r.user_id = ?;

-- name: GetUserFull :one
SELECT
    ap_id,
    url,
    username,
    name,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    created,
    last_updated
FROM users
WHERE local AND ap_id = ?;

-- name: GetUserFullByID :one
-- name: GetUserFull :one
SELECT
    ap_id,
    url,
    username,
    name,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    created,
    last_updated
FROM users
WHERE id = ?;

-- name: UserExists :one
SELECT COUNT(id) == 1 FROM users WHERE ap_id = ?;

-- name: UserIdByOutbox :one
SELECT ap_id from users where outbox = ?;

-- name: UserIdByInbox :one
SELECT ap_id from users where inbox = ?;

-- name: OutboxForInbox :one
SELECT outbox from users where inbox = ?;

-- name: GetInstanceId :one
SELECT id from instances where hostname = ?;

-- name: InsertInstance :one
INSERT INTO instances (hostname, public_key, inbox) VALUES (?, ?, ?) RETURNING id;

-- name: InsertFile :one
INSERT INTO files (
    local,
    digest,
    path,
    ap_id,
    name,
    filename,
    type,
    mime_type,
    size_bytes,
    uploaded_by,
    url
) VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: FileExists :one
SELECT COUNT(id) == 1 FROM files WHERE digest = ?;

-- name: GetFile :one
SELECT
    f.id,
    f.digest,
    f.path,
    f.ap_id,
    f.name,
    f.filename,
    f.type,
    f.mime_type,
    f.size_bytes,
    f.local,
    f.url,
    f.created,
    u.username,
    u.domain
FROM files f
LEFT JOIN users u ON u.id = f.uploaded_by
WHERE f.digest = ?;

-- name: InsertApObject :exec
INSERT INTO ap_object_cache (ap_id, local_table, local_id, type, raw_json, last_updated, last_fetched)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: AddToCollection :one
INSERT INTO ap_collection_members (collection_ap_id, member_ap_id) VALUES (?, ?) RETURNING id;

-- name: GetApObject :one
SELECT ap_id, local_table, local_id, type, raw_json, last_updated, last_fetched
FROM ap_object_cache
WHERE ap_id = ?;

-- name: GetArticleByID :one
SELECT
    ap_id,
    url,
    instance_id,
    language,
    media_type,
    title,
    protected,
    summary,
    content,
    created,
    last_updated
FROM articles where id = ?;

-- name: ApExists :one
SELECT EXISTS(SELECT TRUE FROM ap_object_cache WHERE ap_id = ?);

-- name: UpdateAp :exec
UPDATE ap_object_cache
SET
    raw_json = ?,
    last_updated = (cast(strftime('%s','now') as int))
WHERE ap_id = ?;

-- name: DeleteAp :exec
DELETE FROM ap_object_cache
WHERE ap_id = ?;

-- name: CollectionContains :one
SELECT EXISTS(SELECT TRUE FROM ap_collection_members WHERE collection_ap_id = ? AND member_ap_id = ?);

-- name: GetCollectionFirstPage :many
SELECT member_ap_id FROM ap_collection_members WHERE collection_ap_id = ? ORDER BY id DESC;