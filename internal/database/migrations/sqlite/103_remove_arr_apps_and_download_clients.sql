UPDATE release_action_status
SET action_id = NULL
WHERE action_id IN (SELECT id
                    FROM action
                    WHERE type IN ('RADARR', 'SONARR', 'LIDARR', 'WHISPARR', 'WHISPARR_V3', 'READARR', 'SPORTARR'));

DELETE
FROM action
WHERE type IN ('RADARR', 'SONARR', 'LIDARR', 'WHISPARR', 'WHISPARR_V3', 'READARR', 'SPORTARR');

DELETE
FROM list
WHERE type IN ('RADARR', 'SONARR', 'LIDARR', 'WHISPARR', 'WHISPARR_V3', 'READARR', 'SPORTARR');

-- SQLite cannot drop a column that is part of a foreign key, so action and list are rebuilt.
-- The rows are copied aside and inserted back into a table that has the same name again, because
-- release_action_status and list_filter refer to these tables by name. Deferring the foreign key
-- checks lets the release_action_status references dangle until the action rows are back.
PRAGMA defer_foreign_keys = ON;

CREATE TEMP TABLE action_backup AS
SELECT id,
       name,
       type,
       enabled,
       exec_cmd,
       exec_args,
       watch_folder,
       webhook_host,
       webhook_method,
       webhook_type,
       webhook_data,
       webhook_headers,
       filter_id
FROM action;

DROP TABLE action;

CREATE TABLE action
(
    id              INTEGER PRIMARY KEY,
    name            TEXT,
    type            TEXT,
    enabled         BOOLEAN,
    exec_cmd        TEXT,
    exec_args       TEXT,
    watch_folder    TEXT,
    webhook_host    TEXT,
    webhook_method  TEXT,
    webhook_type    TEXT,
    webhook_data    TEXT,
    webhook_headers TEXT [] DEFAULT '{}',
    filter_id       INTEGER,
    FOREIGN KEY (filter_id) REFERENCES filter (id)
);

INSERT INTO action (id, name, type, enabled, exec_cmd, exec_args, watch_folder, webhook_host, webhook_method,
                    webhook_type, webhook_data, webhook_headers, filter_id)
SELECT id,
       name,
       type,
       enabled,
       exec_cmd,
       exec_args,
       watch_folder,
       webhook_host,
       webhook_method,
       webhook_type,
       webhook_data,
       webhook_headers,
       filter_id
FROM action_backup;

DROP TABLE action_backup;

-- Dropping list cascades to list_filter only when foreign keys are on, so the filter links of the
-- remaining lists are copied aside and list_filter is emptied before they go back in.
CREATE TEMP TABLE list_backup AS
SELECT id,
       name,
       enabled,
       type,
       url,
       headers,
       api_key,
       match_release,
       include_year,
       skip_clean_sanitize,
       last_refresh_time,
       last_refresh_status,
       last_refresh_data,
       created_at,
       updated_at
FROM list;

CREATE TEMP TABLE list_filter_backup AS
SELECT list_id, filter_id
FROM list_filter
WHERE list_id IN (SELECT id FROM list_backup);

DROP TABLE list;

CREATE TABLE list
(
    id                  INTEGER PRIMARY KEY,
    name                TEXT                   NOT NULL,
    enabled             BOOLEAN,
    type                TEXT                   NOT NULL,
    url                 TEXT,
    headers             TEXT []   DEFAULT '{}' NOT NULL,
    api_key             TEXT,
    match_release       BOOLEAN,
    include_year        BOOLEAN   DEFAULT FALSE,
    skip_clean_sanitize BOOLEAN   DEFAULT FALSE,
    last_refresh_time   TIMESTAMP,
    last_refresh_status TEXT,
    last_refresh_data   TEXT,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO list (id, name, enabled, type, url, headers, api_key, match_release, include_year, skip_clean_sanitize,
                  last_refresh_time, last_refresh_status, last_refresh_data, created_at, updated_at)
SELECT id,
       name,
       enabled,
       type,
       url,
       headers,
       api_key,
       match_release,
       include_year,
       skip_clean_sanitize,
       last_refresh_time,
       last_refresh_status,
       last_refresh_data,
       created_at,
       updated_at
FROM list_backup;

DELETE
FROM list_filter;

INSERT INTO list_filter (list_id, filter_id)
SELECT list_id, filter_id
FROM list_filter_backup;

DROP TABLE list_backup;

DROP TABLE list_filter_backup;

DROP TABLE client;
