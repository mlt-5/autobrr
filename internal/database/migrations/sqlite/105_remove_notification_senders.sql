DELETE
FROM filter_notification
WHERE notification_id IN (SELECT id FROM notification WHERE type NOT IN ('WEBHOOK', 'BUILTIN'));

DELETE
FROM notification
WHERE type NOT IN ('WEBHOOK', 'BUILTIN');

ALTER TABLE notification DROP COLUMN token;
ALTER TABLE notification DROP COLUMN api_key;
ALTER TABLE notification DROP COLUMN title;
ALTER TABLE notification DROP COLUMN icon;
ALTER TABLE notification DROP COLUMN host;
ALTER TABLE notification DROP COLUMN username;
ALTER TABLE notification DROP COLUMN password;
ALTER TABLE notification DROP COLUMN channel;
ALTER TABLE notification DROP COLUMN rooms;
ALTER TABLE notification DROP COLUMN targets;
ALTER TABLE notification DROP COLUMN devices;
ALTER TABLE notification DROP COLUMN topic;
ALTER TABLE notification DROP COLUMN sound;
ALTER TABLE notification DROP COLUMN event_sounds;
ALTER TABLE notification DROP COLUMN priority;
