CREATE TABLE ap_object_cache (
    ap_id VARCHAR(255) PRIMARY KEY,

    local_table VARCHAR(32),
    local_id INTEGER,

    type VARCHAR(32) NOT NULL,
    raw_json TEXT,
    inserted_at INTEGER DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_updated INTEGER,

    -- Last time the object was fetched. For local objects,
    -- last_fetched is null.
    last_fetched INTEGER,

    UNIQUE (local_table, local_id)
);

CREATE TABLE ap_collection_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_ap_id VARCHAR(255) NOT NULL,
    member_ap_id VARCHAR(255) NOT NULL,
    inserted_at INTEGER DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,

    UNIQUE (collection_ap_id, member_ap_id)
);

CREATE INDEX idx_collection_pagination
ON ap_collection_members (collection_ap_id, id DESC);

CREATE INDEX idx_reverse_collection_lookup
ON ap_collection_members (member_ap_id);