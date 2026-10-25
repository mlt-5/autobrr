CREATE TABLE release_vocab
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    kind         TEXT                  NOT NULL,
    tag          TEXT                  NOT NULL,
    title        TEXT    DEFAULT ''    NOT NULL,
    aliases      TEXT [] DEFAULT '{}'  NOT NULL,
    pattern      TEXT    DEFAULT ''    NOT NULL,
    release_type TEXT    DEFAULT ''    NOT NULL,
    exclusive    BOOLEAN DEFAULT FALSE NOT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (kind, tag)
);

CREATE TABLE release_vocab_removed
(
    kind TEXT NOT NULL,
    tag  TEXT NOT NULL,
    PRIMARY KEY (kind, tag)
);
