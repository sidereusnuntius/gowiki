CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bot BOOLEAN DEFAULT FALSE NOT NULL,
    local BOOLEAN DEFAULT TRUE NOT NULL,
    ap_id VARCHAR(255) NOT NULL,
    url VARCHAR(255),
    username VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(64),
    summary TEXT,
    inbox TEXT,
    outbox VARCHAR(255),
    followers VARCHAR(255),
    public_key TEXT,
    private_key TEXT,
    trusted BOOLEAN NOT NULL,
    created INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_updated INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_fetched INT,

    UNIQUE (username, domain),
    UNIQUE (ap_id),
    UNIQUE (inbox),
    UNIQUE(outbox)
);

-- Store also the key the user used to sign up, if they signed up using an invitation
CREATE TABLE accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    password VARCHAR(60) NOT NULL,
    admin BOOLEAN DEFAULT false NOT NULL,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT false NOT NULL,
    approved BOOLEAN DEFAULT false NOT NULL,
    user_id INTEGER NOT NULL,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL,

    UNIQUE (email),
    UNIQUE (user_id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE invitations (
    id VARCHAR PRIMARY KEY,
    used BOOLEAN DEFAULT FALSE NOT NULL,
    used_by INTEGER,
    used_at TEXT,
    made_by INT NOT NULL,
    created TEXT NOT NULL,

    FOREIGN KEY (made_by) REFERENCES accounts (id),
    FOREIGN KEY (used_by) REFERENCES accounts (id),
    CHECK (made_by != used_by)
);

CREATE TABLE approval_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INT NOT NULL,
    reason TEXT NOT NULL,
    approved BOOLEAN,
    reviewer INT,
    created TEXT NOT NULL,
    
    UNIQUE (account_id),
    FOREIGN KEY (reviewer) REFERENCES accounts (id),
    CHECK (reviewer != account_id)
);

CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    local BOOLEAN DEFAULT TRUE NOT NULL,
    ap_id VARCHAR NOT NULL,
    url VARCHAR,
    instance_id INTEGER,
    language VARCHAR NOT NULL,
    media_type VARCHAR NOT NULL,
    title VARCHAR(255) NOT NULL,
    protected BOOLEAN DEFAULT FALSE NOT NULL,
    summary TEXT,
    content TEXT NOT NULL,
    created INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_updated INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_fetched INT,

    UNIQUE (ap_id),
    UNIQUE (title, instance_id),
    FOREIGN KEY (instance_id) REFERENCES instances (id)
);

CREATE TABLE revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ap_id VARCHAR(255),
    article_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    summary TEXT,
    diff TEXT NOT NULL,
    -- reviewed remains false until another user reads and either rejects or approves it.
    reviewed BOOLEAN DEFAULT FALSE NOT NULL,
    reviewer INTEGER,
    reviewed_at TEXT,
    -- The latest revision with published = true is the current revision of the article.
    published BOOLEAN DEFAULT FALSE NOT NULL,
    prev INTEGER,
    based_on INTEGER,
    created INTEGER DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,

    UNIQUE (ap_id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (reviewer) REFERENCES users (id),
    FOREIGN KEY (article_id) REFERENCES articles (id),
    FOREIGN KEY (prev) REFERENCES revisions (id),
    FOREIGN KEY (based_on) REFERENCES revisions (id)
);

CREATE TABLE instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hostname VARCHAR(255) NOT NULL,
    public_key TEXT,
    inbox VARCHAR(255),
    created INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    updated INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL
);

CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    digest CHAR(64) NOT NULL,
    path VARCHAR(255),
    ap_id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    filename VARCHAR(255),
    type VARCHAR(32) NOT NULL DEFAULT 'Document',
    mime_type VARCHAR(128) NOT NULL,
    size_bytes INTEGER,
    local BOOLEAN DEFAULT FALSE NOT NULL,
    uploaded_by INTEGER,
    url VARCHAR(255) NOT NULL,
    created INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,

    UNIQUE (digest),
    UNIQUE (ap_id),
    FOREIGN KEY (uploaded_by) REFERENCES users (id)
);

CREATE TABLE article_files (
    article_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,

    FOREIGN KEY (article_id) REFERENCES articles (id),
    FOREIGN KEY (file_id) REFERENCES files (id),
    PRIMARY KEY (article_id, file_id)
);

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