CREATE TABLE users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bot BOOLEAN DEFAULT FALSE NOT NULL,
    local BOOLEAN DEFAULT TRUE NOT NULL,
    ap_id VARCHAR(255) NOT NULL,
    username VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(64),
    summary TEXT,
    inbox VARCHAR(255),
    outbox VARCHAR(255),
    followers VARCHAR(255),
    public_key TEXT,
    private_key TEXT,
    created TIMESTAMP DEFAULT NOW() NOT NULL,
    last_updated TIMESTAMP DEFAULT NOW() NOT NULL,
    last_fetched TIMESTAMP,

    UNIQUE (username, domain),
    UNIQUE (ap_id),
    UNIQUE (inbox),
    UNIQUE(outbox)
);

-- Store also the key the user used to sign up, if they signed up using an invitation
CREATE TABLE accounts (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    password BYTEA NOT NULL,
    admin BOOLEAN DEFAULT false NOT NULL,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT false NOT NULL,
    approved BOOLEAN DEFAULT false NOT NULL,
    user_id INTEGER NOT NULL,
    created TIMESTAMP DEFAULT NOW() NOT NULL,
    last_updated TIMESTAMP NOT NULL,

    UNIQUE (email),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE invitations (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    used BOOLEAN DEFAULT FALSE NOT NULL,
    used_by INT,
    used_at TIMESTAMP,
    made_by INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,

    FOREIGN KEY (made_by) REFERENCES accounts (id),
    FOREIGN KEY (used_by) REFERENCES accounts (id),
    CHECK (made_by != used_by)
);

CREATE TABLE approval_requests (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id INT NOT NULL,
    reason TEXT NOT NULL,
    approved BOOLEAN,
    reviewer INT,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    
    UNIQUE (account_id),
    FOREIGN KEY (reviewer) REFERENCES accounts (id),
    CHECK (reviewer != account_id)
);