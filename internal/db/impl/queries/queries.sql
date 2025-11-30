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
    attributed_to,
    url,
    instance_id,
    language,
    media_type,
    title,
    content,
    published
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

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
    u.id,
    u.username,
    u.name,
    i.hostname,
    u.url,
    u.local,
    u.summary
FROM users u
JOIN instances i ON u.instance_id = i.id
WHERE u.username = lower(?) AND NOT u.local AND i.hostname = ?;

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

-- name: ActorIdByOutbox :one
SELECT ap_id from users u where u.outbox = ?1
UNION
SELECT url as ap_id FROM instances i WHERE i.outbox = ?1;

-- name: ActorIdByInbox :one
SELECT ap_id FROM users u WHERE u.inbox = ?1
UNION
SELECT url AS ap_id FROM instances i WHERE i.inbox = ?1;

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
    i.hostname
FROM files f
LEFT JOIN users u ON u.id = f.uploaded_by
LEFT JOIN instances i ON u.instance_id = i.id
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
    attributed_to,
    url,
    instance_id,
    language,
    media_type,
    title,
    protected,
    summary,
    content,
    published,
    inserted_at,
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

-- name: Follow :one
INSERT INTO follows (
    follow_ap_id,
    follower_ap_id,
    followee_ap_id,
    follower_inbox_url
) VALUES (?, ?, ?, ?)
RETURNING id;

-- name: GetUserKeys :one
SELECT ap_id, private_key FROM users WHERE local AND id = ?;

-- name: GetPrivateKeyByID :one
SELECT private_key FROM users WHERE local AND ap_id = ?1
UNION
SELECT private_key FROM instances WHERE private_key IS NOT NULL AND url = ?1
LIMIT 1;

-- name: InsertOrUpdateUser :one
INSERT INTO users (
    local,
    ap_id,
    url,
    username,
    name,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    trusted,
    last_updated,
    last_fetched
) VALUES (false, ?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, false, (cast(strftime('%s','now') as int)), ?10)
ON CONFLICT (ap_id) DO UPDATE
SET url = ?2,
    username = ?3,
    name = ?4,
    summary = ?5,
    inbox = ?6,
    outbox = ?7,
    followers = ?8,
    public_key = ?9,
    last_updated = cast(strftime('%s','now') as int),
    last_fetched = ?10
RETURNING id;

-- name: InsertOrUpdateApObject :exec
INSERT INTO ap_object_cache (ap_id, local_table, local_id, type, raw_json, last_fetched)
VALUES (?1, ?2, ?3, ?4, ?5, ?6)
ON CONFLICT (ap_id) DO UPDATE
SET type = ?4,
    raw_json = ?5,
    last_updated = cast(strftime('%s','now') as int),
    last_fetched = ?6;

-- name: GetInboxByActorId :one
SELECT u.inbox as inbox FROM users u WHERE u.ap_id = ?1
UNION
SELECT i.inbox as inbox FROM instances i WHERE i.url = ?1
LIMIT 1;

-- name: UpdateFollowInbox :exec
UPDATE follows SET follower_inbox_url = ? WHERE follower_ap_id = ?;

-- name: GetCollectiveByID :one
SELECT
    cache.type,
    i.name,
    i.hostname,
    i.url,
    i.public_key,
    i.inbox,
    i.outbox,
    i.followers
FROM instances i
JOIN ap_object_cache cache ON cache.ap_id = i.url
WHERE id = ?
LIMIT 1;

-- name: GetUserApId :one
SELECT ap_id FROM users WHERE local AND username = lower(?1)
UNION
SELECT url AS ap_id FROM INSTANCES WHERE name = lower(?1);

-- name: GetFollowers :many
SELECT follower_ap_id FROM follows WHERE followee_ap_id = ?;

-- name: GetUserUriById :one
SELECT ap_id FROM users WHERE id = ? LIMIT 1;