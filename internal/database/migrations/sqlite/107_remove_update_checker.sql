UPDATE notification
SET events = '{' || trim(replace(replace(',' || substr(events, 2, length(events) - 2) || ',', ',APP_UPDATE_AVAILABLE,', ','), ',"APP_UPDATE_AVAILABLE",', ','), ',') || '}'
WHERE events LIKE '%APP_UPDATE_AVAILABLE%';

DELETE
FROM notification_inbox
WHERE event = 'APP_UPDATE_AVAILABLE';
