CREATE TABLE follows2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follow_ap_id VARCHAR(255) NOT NULL,
    follower_ap_id VARCHAR(255) NOT NULL,
    followee_ap_id VARCHAR(255) NOT NULL,
    follower_inbox_url VARCHAR(255),
    approved BOOLEAN DEFAULT FALSE NOT NULL,
    created_at INTEGER DEFAULT (cast(strftime('%s','now') as int)),

    UNIQUE (follow_ap_id),
    UNIQUE (follower_ap_id, followee_ap_id),
    CHECK (follower_ap_id != followee_ap_id)
);

INSERT INTO follows2 SELECT * FROM follows;

DROP TABLE follows;
ALTER TABLE follows2 RENAME TO follows;