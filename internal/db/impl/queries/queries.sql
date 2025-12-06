-- name: CreateLocalUser :one
INSERT INTO users (
    ap_id,
    username,
    name,
    host,
    trusted,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    private_key
    )
VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11) RETURNING id;

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

-- name: GetArticle :one
SELECT
    art.title,
    art.summary,
    art.content,
    art.protected,
    art.media_type,
    art.language,
    art.url,
    art.published,
    art.last_updated,
    art.host,
    att.name AS author
FROM
    articles AS art
JOIN (
    SELECT username AS name, ap_id AS id FROM users WHERE users.username = ?2
    UNION
    SELECT name AS name, url AS id FROM collectives WHERE collectives.name = ?2
) AS att ON att.id = art.attributed_to
where lower(art.title) = ?1 AND art.host = ?3
LIMIT 1;

-- name: GetAuthor :one
SELECT username AS name, host, 'person' AS actor_type FROM users WHERE ap_id = ?1
UNION
SELECT name AS name, host, 'collective' AS actor_type FROM collectives WHERE url = ?1
LIMIT 1;

-- name: IsUserTrusted :one
SELECT trusted FROM users where id = ?1 LIMIT 1;

-- name: CreateArticle :one
INSERT INTO articles (
    local,
    ap_id,
    author,
    attributed_to,
    url,
    language,
    media_type,
    title,
    host,
    type,
    protected,
    summary,
    content,
    published,
    last_updated,
    last_fetched
) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13, ?14, ?15, ?16)
ON CONFLICT (ap_id) DO UPDATE
SET
    url = ?5,
    language = ?6,
    media_type = ?7,
    title = ?8,
    host = ?9,
    type = ?10,
    protected = ?11,
    summary = ?12,
    content = ?13,
    published = ?14,
    last_updated = ?15,
    last_fetched = ?16
RETURNING id;

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

-- name: GetArticleIdByIri :one
SELECT id FROM articles WHERE ap_id = ? LIMIT 1;

-- name: ArticleTitleExists :one
SELECT EXISTS (
    SELECT TRUE
    FROM articles
    WHERE lower(title) = lower(?) AND author = ? AND host = ?
) AS BOOLEAN;

-- name: GetArticleIDS :one
SELECT
    art.id,
    art.ap_id,
    r.id AS rev_id
FROM articles art
JOIN (
    SELECT username AS name, ap_id AS id FROM users WHERE users.username = @author
    UNION
    SELECT name AS name, url AS id FROM collectives WHERE collectives.name = @author
) AS att ON att.id = art.attributed_to
JOIN revisions r ON r.article_id = art.id
WHERE LOWER(art.title) = LOWER(@title) AND art.host = @host
ORDER BY r.created DESC
LIMIT 1;

-- name: InsertRevision :one
INSERT INTO revisions (
    ap_id,
    article_id,
    user_id,
    summary,
    diff,
    published,
    prev
) VALUES (?1, ?2, ?3, ?4, ?5, true, ?6)
RETURNING id;

-- name: UpdateRevisionApId :exec
UPDATE revisions SET ap_id = ? WHERE id = ?;

-- name: UpdateArticle :exec
UPDATE articles
SET
    content = ?1,
    last_updated = (cast(strftime('%s','now') as int))
WHERE id = ?2;

-- name: UpdateArticleByIRI :exec
UPDATE articles
SET
    content = ?,
    last_updated = (cast(strftime('%s','now') as int))
WHERE ap_id = ?;

-- name: GetArticleContent :one
SELECT content FROM articles WHERE id = ?;

-- name: GetArticleContentByIRI :one
SELECT id, content FROM articles WHERE ap_id = ?;

-- name: GetRevisionList :many
SELECT
    r.id,
    r.reviewed,
    r.summary,
    a.title,
    u.username,
    r.created
FROM (
    SELECT
        art.id,
        art.title
    FROM articles art
    LEFT JOIN (
        SELECT username AS name, ap_id AS id FROM users WHERE users.username = @author
        UNION
        SELECT name AS name, url AS id FROM collectives WHERE collectives.name = @author
    ) AS att ON att.id = art.attributed_to
    WHERE LOWER(art.title) = LOWER(@title) AND art.host = @host
    LIMIT 1
) a
JOIN revisions r ON r.article_id = a.id
JOIN users u ON r.user_id = u.id
ORDER BY r.created DESC;

-- name: GetActorData :one
SELECT
    id,
    username as name,
    host,
    ap_id,
    url,
    summary,
    'user' AS type
FROM users u WHERE u.username = lower(?1) AND u.host = ?2
UNION
SELECT
    id,
    name,
    host,
    url as ap_id,
    url,
    summary,
    'group' AS type
FROM collectives c WHERE c.name = lower(?1) AND c.host = ?2;

-- name: GetArticlesByActorId :many
SELECT
    title,
    published
FROM articles
WHERE attributed_to = ?;

-- name: GetForeignUserData :one
SELECT
    u.id,
    u.username,
    u.name,
    u.host,
    u.url,
    u.local,
    u.summary
FROM users u
WHERE u.username = lower(?) AND NOT u.local AND u.host = ?;

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
SELECT url as ap_id FROM collectives i WHERE i.outbox = ?1;

-- name: ActorIdByInbox :one
SELECT ap_id FROM users u WHERE u.inbox = ?1
UNION
SELECT url AS ap_id FROM collectives i WHERE i.inbox = ?1;

-- name: OutboxForInbox :one
SELECT outbox from users where inbox = ?;

-- name: GetInstanceId :one
SELECT id from collectives where host = ?;

-- name: InsertInstance :one
INSERT INTO collectives (host, public_key, inbox) VALUES (?, ?, ?) RETURNING id;

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
    f.host,
    u.username
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
    attributed_to,
    url,
    host,
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

-- name: GetArticleByIRI :one
SELECT
    ap_id,
    attributed_to,
    url,
    host,
    language,
    media_type,
    title,
    protected,
    summary,
    content,
    published,
    inserted_at,
    last_updated
FROM articles where ap_id = ?;

-- name: ApExists :one
SELECT EXISTS(SELECT TRUE FROM ap_object_cache WHERE ap_id = ?) AS BOOLEAN;

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
SELECT private_key FROM collectives WHERE private_key IS NOT NULL AND url = ?1
LIMIT 1;

-- name: InsertOrUpdateUser :one
INSERT INTO users (
    local,
    ap_id,
    url,
    username,
    name,
    host,
    summary,
    inbox,
    outbox,
    followers,
    public_key,
    trusted,
    last_updated,
    last_fetched
) VALUES (false, ?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, false, (cast(strftime('%s','now') as int)), ?11)
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
SELECT i.inbox as inbox FROM collectives i WHERE i.url = ?1
LIMIT 1;

-- name: UpdateFollowInbox :exec
UPDATE follows SET follower_inbox_url = ? WHERE follower_ap_id = ?;

-- name: GetCollectiveByID :one
SELECT
    cache.type,
    i.name,
    i.host,
    i.url,
    i.public_key,
    i.inbox,
    i.outbox,
    i.followers
FROM collectives i
JOIN ap_object_cache cache ON cache.ap_id = i.url
WHERE i.id = ?
LIMIT 1;

-- name: GetUserApId :one
SELECT ap_id FROM users WHERE local AND username = lower(?1)
UNION
SELECT url AS ap_id FROM collectives WHERE name = lower(?1);

-- name: GetFollowers :many
SELECT follower_ap_id FROM follows WHERE followee_ap_id = ?;

-- name: GetUserUriById :one
SELECT ap_id FROM users WHERE id = ? LIMIT 1;

-- name: CollectionMembersIRIs :many
SELECT member_ap_id FROM ap_collection_members WHERE collection_ap_id = ?;

-- name: GetCollectionActivitiesPage :many
SELECT
    cache.ap_id,
    cache.raw_json
FROM
    ap_collection_members col
JOIN
    ap_object_cache cache
ON cache.ap_id = col.member_ap_id
WHERE col.collection_ap_id = @collection_id AND (
    cache.id < sqlc.narg('last_id')
    OR
    sqlc.narg('last_id') IS NULL
)
ORDER BY cache.id DESC
LIMIT @page_size;

-- name: GetCollectionStart :one
SELECT
    CAST(MAX(cache.id) AS BIGINT) AS start,
    CAST(COUNT(cache.id) AS BIGINT) AS size
FROM
    ap_collection_members col
JOIN
    ap_object_cache cache
ON cache.ap_id = col.member_ap_id
WHERE
    col.collection_ap_id = @collection_id
ORDER BY cache.id DESC;

-- name: GetUserId :one
SELECT id FROM users where ap_id = ? LIMIT 1;

-- name: IsAdmin :one
SELECT admin FROM accounts a WHERE a.id = ?; 