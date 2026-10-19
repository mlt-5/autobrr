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

ALTER TABLE action
    DROP COLUMN external_client_id,
    DROP COLUMN external_client,
    DROP COLUMN client_id;

ALTER TABLE list
    DROP COLUMN client_id,
    DROP COLUMN tags_included,
    DROP COLUMN tags_excluded,
    DROP COLUMN include_unmonitored,
    DROP COLUMN include_alternate_titles;

DROP TABLE client;
