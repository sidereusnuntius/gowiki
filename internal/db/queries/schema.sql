CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bot BOOLEAN DEFAULT FALSE NOT NULL,
    local BOOLEAN DEFAULT TRUE NOT NULL,
    ap_id VARCHAR(255) NOT NULL,
    username VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(64),
    summary TEXT,
    inbox TEXT,
    outbox VARCHAR(255),
    followers VARCHAR(255),
    public_key TEXT,
    private_key TEXT,
    created TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    last_fetched TEXT,

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
    created_at TEXT NOT NULL,

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
    created_at TEXT NOT NULL,
    
    UNIQUE (account_id),
    FOREIGN KEY (reviewer) REFERENCES accounts (id),
    CHECK (reviewer != account_id)
);