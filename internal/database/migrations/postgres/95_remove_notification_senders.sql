DELETE
FROM filter_notification
WHERE notification_id IN (SELECT id FROM notification WHERE type NOT IN ('WEBHOOK', 'BUILTIN'));

DELETE
FROM notification
WHERE type NOT IN ('WEBHOOK', 'BUILTIN');

ALTER TABLE notification
    DROP COLUMN token,
    DROP COLUMN api_key,
    DROP COLUMN title,
    DROP COLUMN icon,
    DROP COLUMN host,
    DROP COLUMN username,
    DROP COLUMN password,
    DROP COLUMN channel,
    DROP COLUMN rooms,
    DROP COLUMN targets,
    DROP COLUMN devices,
    DROP COLUMN topic,
    DROP COLUMN sound,
    DROP COLUMN event_sounds,
    DROP COLUMN priority;
