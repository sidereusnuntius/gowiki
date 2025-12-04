CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    local BOOLEAN DEFAULT TRUE NOT NULL,
    ap_id VARCHAR NOT NULL,
    url VARCHAR,
    language VARCHAR NOT NULL,
    media_type VARCHAR NOT NULL,
    title VARCHAR(255) NOT NULL,
    host VARCHAR(255),
    type VARCHAR(32) NOT NULL,
    protected BOOLEAN DEFAULT FALSE NOT NULL,
    summary TEXT,
    content TEXT NOT NULL,
    created INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_updated INT DEFAULT (cast(strftime('%s','now') as int)) NOT NULL,
    last_fetched INT,

    UNIQUE (ap_id)
);

CREATE TABLE revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ap_id VARCHAR(255),
    article_id INTEGER NOT NULL,
    user_id INTEGER,
    diff TEXT NOT NULL,
    summary TEXT,
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
