CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    digest CHAR(64) NOT NULL,
    path VARCHAR(255) NOT NULL,
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