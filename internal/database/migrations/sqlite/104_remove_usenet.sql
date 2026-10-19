DELETE
FROM feed_cache
WHERE feed_id IN (SELECT id
                  FROM feed
                  WHERE type = 'NEWZNAB'
                     OR indexer_id IN (SELECT id FROM indexer WHERE implementation = 'newznab' OR identifier = 'curupira'));

DELETE
FROM feed
WHERE type = 'NEWZNAB'
   OR indexer_id IN (SELECT id FROM indexer WHERE implementation = 'newznab' OR identifier = 'curupira');

DELETE
FROM filter_indexer
WHERE indexer_id IN (SELECT id FROM indexer WHERE implementation = 'newznab' OR identifier = 'curupira');

DELETE
FROM indexer
WHERE implementation = 'newznab'
   OR identifier = 'curupira';

DELETE
FROM irc_channel
WHERE name = '#curupira-anuncios'
  AND network_id IN (SELECT id FROM irc_network WHERE server = 'irc.synirc.net');

DELETE
FROM irc_network
WHERE server = 'irc.synirc.net'
  AND NOT EXISTS (SELECT 1 FROM irc_channel WHERE irc_channel.network_id = irc_network.id);

UPDATE feed
SET settings = REPLACE(settings, '"download_type":"NZB"', '"download_type":"TORRENT"')
WHERE settings LIKE '%"download_type":"NZB"%';
