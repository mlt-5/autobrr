-- Support for these download clients was removed, so their actions and clients can no longer run.
-- Release history is kept by unlinking it from the actions before they are deleted.
UPDATE release_action_status
SET action_id = NULL
WHERE action_id IN (
    SELECT id FROM action
    WHERE type IN ('QBITTORRENT', 'DELUGE_V1', 'DELUGE_V2', 'RTORRENT', 'TRANSMISSION', 'PORLA', 'ARIA2', 'SABNZBD', 'NZBGET')
);

DELETE FROM action
WHERE type IN ('QBITTORRENT', 'DELUGE_V1', 'DELUGE_V2', 'RTORRENT', 'TRANSMISSION', 'PORLA', 'ARIA2', 'SABNZBD', 'NZBGET');

DELETE FROM client
WHERE type IN ('QBITTORRENT', 'DELUGE_V1', 'DELUGE_V2', 'RTORRENT', 'TRANSMISSION', 'PORLA', 'ARIA2', 'SABNZBD', 'NZBGET');
