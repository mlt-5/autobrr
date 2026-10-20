UPDATE notification
SET events = array_remove(events, 'APP_UPDATE_AVAILABLE')
WHERE 'APP_UPDATE_AVAILABLE' = ANY (events);

DELETE
FROM notification_inbox
WHERE event = 'APP_UPDATE_AVAILABLE';
