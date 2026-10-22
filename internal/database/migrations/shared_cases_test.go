// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const insertIrcNetwork = `INSERT INTO irc_network (
	id, enabled, name, server, port, tls, tls_skip_verify, pass, nick,
	auth_mechanism, auth_account, auth_password, invite_command,
	use_bouncer, bouncer_addr, bot_mode, use_proxy, proxy_id, created_at, updated_at
) VALUES `

func setupNordicBytesNotMigrated(db *sql.DB) error {
	if _, err := db.Exec(insertIrcNetwork + `
		(1, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_a',
		 'SASL_PLAIN', 'acct_a', 'secret_a', '', FALSE, '', TRUE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00'),
		(2, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6667, FALSE, FALSE, '', 'bot_b',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00'),
		(3, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_b',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00'),
		(4, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_c',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00')`); err != nil {
		return err
	}

	_, err := db.Exec(`INSERT INTO irc_channel (id, enabled, name, password, detached, network_id) VALUES
		(1, TRUE, '#nordicbytes', 'chan_pw', FALSE, 1),
		(2, TRUE, '#bithdtv', '', FALSE, 1),
		(3, FALSE, '#NordicBytes', '', TRUE, 2),
		(4, TRUE, '#nordicbytes', '', FALSE, 3),
		(5, TRUE, '#ncore', '', FALSE, 4)`)

	return err
}

func validateNordicBytesNotMigrated(db *sql.DB, t *testing.T) {
	var port int
	var tls, botMode bool
	var mechanism, account, password string

	err := db.QueryRow(`SELECT port, tls, bot_mode, auth_mechanism, auth_account, auth_password FROM irc_network WHERE server = 'irc.nordicbytes.org' AND nick = 'bot_a'`).
		Scan(&port, &tls, &botMode, &mechanism, &account, &password)
	require.NoError(t, err)
	assert.Equal(t, 6697, port)
	assert.True(t, tls)
	assert.True(t, botMode)
	assert.Equal(t, "SASL_PLAIN", mechanism)
	assert.Equal(t, "acct_a", account)
	assert.Equal(t, "secret_a", password)

	var chanPassword string
	err = db.QueryRow(`SELECT password FROM irc_channel WHERE network_id = (SELECT id FROM irc_network WHERE server = 'irc.nordicbytes.org' AND nick = 'bot_a') AND name = '#announce'`).Scan(&chanPassword)
	require.NoError(t, err)
	assert.Equal(t, "chan_pw", chanPassword)

	// bot_b had two p2p-network rows; only the lowest id is carried over so UNIQUE (server, port, nick) holds.
	var enabled, detached bool
	err = db.QueryRow(`SELECT n.port, n.tls, c.enabled, c.detached FROM irc_network n JOIN irc_channel c ON c.network_id = n.id WHERE n.server = 'irc.nordicbytes.org' AND n.nick = 'bot_b' AND c.name = '#announce'`).
		Scan(&port, &tls, &enabled, &detached)
	require.NoError(t, err)
	assert.Equal(t, 6697, port)
	assert.True(t, tls, "new network must use TLS even if the p2p-network row did not")
	assert.False(t, enabled)
	assert.True(t, detached)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM irc_network WHERE server = 'irc.nordicbytes.org'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	err = db.QueryRow(`SELECT COUNT(*) FROM irc_network WHERE id IN (2, 3)`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "p2p-network rows only used by NordicBytes must be removed")

	var name string
	err = db.QueryRow(`SELECT name FROM irc_channel WHERE network_id = 1`).Scan(&name)
	require.NoError(t, err, "shared p2p-network row must keep its other channels and nothing else")
	assert.Equal(t, "#bithdtv", name)

	err = db.QueryRow(`SELECT COUNT(*) FROM irc_channel WHERE network_id = 4 AND name = '#ncore'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "other p2p-network rows must not be touched")
}

func setupNordicBytesAlreadyMigrated(db *sql.DB) error {
	if _, err := db.Exec(insertIrcNetwork + `
		(1, TRUE, 'NordicBytes', 'irc.nordicbytes.org', 6697, TRUE, FALSE, '', 'bot_x',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00'),
		(2, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_a',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00'),
		(3, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_b',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00')`); err != nil {
		return err
	}

	_, err := db.Exec(`INSERT INTO irc_channel (id, enabled, name, password, detached, network_id) VALUES
		(1, TRUE, '#announce', '', FALSE, 1),
		(2, TRUE, '#nordicbytes', '', FALSE, 2),
		(3, TRUE, '#bithdtv', '', FALSE, 2),
		(4, TRUE, '#nordicbytes', '', FALSE, 3)`)

	return err
}

func validateNordicBytesAlreadyMigrated(db *sql.DB, t *testing.T) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM irc_network WHERE server = 'irc.nordicbytes.org'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "an existing NordicBytes network must not get a duplicate")

	err = db.QueryRow(`SELECT COUNT(*) FROM irc_channel WHERE network_id = 1`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var name string
	err = db.QueryRow(`SELECT name FROM irc_channel WHERE network_id = 2`).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "#bithdtv", name)

	err = db.QueryRow(`SELECT COUNT(*) FROM irc_network WHERE id = 3`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "stale p2p-network row must be removed")
}

func setupNordicBytesNotUsed(db *sql.DB) error {
	if _, err := db.Exec(insertIrcNetwork + `
		(1, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_a',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00'),
		(2, TRUE, 'P2P-Network', 'irc.p2p-network.net', 6697, TRUE, FALSE, '', 'bot_b',
		 'NONE', '', '', '', FALSE, '', FALSE, FALSE, NULL,
		 '2025-01-01 00:00:00', '2025-01-01 00:00:00')`); err != nil {
		return err
	}

	_, err := db.Exec(`INSERT INTO irc_channel (id, enabled, name, password, detached, network_id) VALUES
		(1, TRUE, '#bithdtv', '', FALSE, 1)`)

	return err
}

func validateNordicBytesNotUsed(db *sql.DB, t *testing.T) {
	var networks, channels int
	err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM irc_network), (SELECT COUNT(*) FROM irc_channel)`).Scan(&networks, &channels)
	require.NoError(t, err)
	assert.Equal(t, 2, networks, "nothing must change without NordicBytes, including empty p2p-network rows")
	assert.Equal(t, 1, channels)
}

func setupBuiltinNotificationExisting(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO notification (name, type, enabled, events, webhook)
		VALUES ('Hook', 'WEBHOOK', TRUE, '{PUSH_APPROVED}', 'https://hook.example/notify')`)

	return err
}

func validateBuiltinNotification(db *sql.DB, t *testing.T, wantEvents []string) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM notification WHERE type = 'BUILTIN'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var name string
	var enabled bool
	var events []string
	err = db.QueryRow(`SELECT name, enabled, events FROM notification WHERE type = 'BUILTIN'`).Scan(&name, &enabled, pq.Array(&events))
	require.NoError(t, err)
	assert.Equal(t, "Built-in", name)
	assert.True(t, enabled)
	assert.Equal(t, wantEvents, events)

	_, err = db.Exec(`INSERT INTO notification (name, type, enabled) VALUES ('Second', 'BUILTIN', TRUE)`)
	assert.Error(t, err, "the unique index must reject a second built-in notification")

	_, err = db.Exec(`INSERT INTO notification (name, type, enabled) VALUES ('Another hook', 'WEBHOOK', TRUE)`)
	assert.NoError(t, err, "the unique index must only cover the built-in type")
}

func validateBuiltinNotificationKeepsExisting(db *sql.DB, t *testing.T) {
	validateBuiltinNotification(db, t, []string{"PUSH_ERROR", "IRC_DISCONNECTED", "APP_UPDATE_AVAILABLE"})

	var webhook string
	err := db.QueryRow(`SELECT webhook FROM notification WHERE name = 'Hook'`).Scan(&webhook)
	require.NoError(t, err)
	assert.Equal(t, "https://hook.example/notify", webhook)
}

func setupBuiltinNotificationPresent(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO notification (name, type, enabled, events)
		VALUES ('Existing', 'BUILTIN', FALSE, '{PUSH_REJECTED}')`)

	return err
}

func validateBuiltinNotificationNotDuplicated(db *sql.DB, t *testing.T) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM notification WHERE type = 'BUILTIN'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var name string
	err = db.QueryRow(`SELECT name FROM notification WHERE type = 'BUILTIN'`).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Existing", name, "the seed must not replace a built-in row that already exists")
}

func execAll(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func queryIDs(t *testing.T, db *sql.DB, query string) []int {
	t.Helper()

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())

	return ids
}

func setupDownloadClientData(db *sql.DB) error {
	return execAll(db,
		`INSERT INTO filter (id, name) VALUES (1, 'movies')`,
		`INSERT INTO client (id, name, type, enabled, host) VALUES (1, 'qbit', 'QBITTORRENT', TRUE, 'http://qbit')`,
		`INSERT INTO client (id, name, type, enabled, host) VALUES (2, 'radarr', 'RADARR', TRUE, 'http://radarr')`,
		`INSERT INTO action (id, name, type, enabled, client_id, filter_id) VALUES (1, 'qbit', 'QBITTORRENT', TRUE, 1, 1)`,
		`INSERT INTO action (id, name, type, enabled, watch_folder, filter_id) VALUES (2, 'watch', 'WATCH_FOLDER', TRUE, '/watch', 1)`,
		`INSERT INTO "release" (id, torrent_name) VALUES (1, 'Test.Release-GRP')`,
		`INSERT INTO release_action_status (id, status, action, action_id, type, filter_id, release_id) VALUES (1, 'PUSH_APPROVED', 'qbit', 1, 'QBITTORRENT', 1, 1)`,
	)
}

func validateDownloadClientsRemoved(db *sql.DB, t *testing.T) {
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT id FROM action ORDER BY id`))
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT id FROM client ORDER BY id`))

	var actionID sql.NullInt64
	require.NoError(t, db.QueryRow(`SELECT action_id FROM release_action_status WHERE id = 1`).Scan(&actionID))
	assert.False(t, actionID.Valid, "release history is kept but unlinked from the deleted action")
}

func setupTorrentClientColumnData(db *sql.DB) error {
	return execAll(db,
		`INSERT INTO filter (id, name) VALUES (1, 'movies')`,
		`INSERT INTO action (id, name, type, enabled, watch_folder, category, filter_id) VALUES (1, 'watch', 'WATCH_FOLDER', TRUE, '/watch', 'movies', 1)`,
	)
}

func validateTorrentClientColumnsDropped(db *sql.DB, t *testing.T) {
	var watchFolder string
	require.NoError(t, db.QueryRow(`SELECT watch_folder FROM action WHERE id = 1`).Scan(&watchFolder))
	assert.Equal(t, "/watch", watchFolder)

	_, err := db.Exec(`SELECT category FROM action`)
	assert.Error(t, err, "action.category must be dropped")
	_, err = db.Exec(`SELECT reannounce_max_attempts FROM action`)
	assert.Error(t, err, "action.reannounce_max_attempts must be dropped")
}

func setupUsenetData(db *sql.DB) error {
	return execAll(db,
		`INSERT INTO filter (id, name) VALUES (1, 'movies')`,
		`INSERT INTO indexer (id, identifier, implementation, name, enabled) VALUES (1, 'generic-newznab', 'newznab', 'Newznab', TRUE)`,
		`INSERT INTO indexer (id, identifier, implementation, name, enabled) VALUES (2, 'curupira', 'irc', 'Curupira', TRUE)`,
		`INSERT INTO indexer (id, identifier, implementation, name, enabled) VALUES (3, 'torznab', 'torznab', 'Torznab', TRUE)`,
		`INSERT INTO feed (id, indexer_id, name, type, enabled) VALUES (1, 1, 'newznab', 'NEWZNAB', TRUE)`,
		`INSERT INTO feed (id, indexer_id, name, type, enabled) VALUES (2, 3, 'torznab', 'TORZNAB', TRUE)`,
		`INSERT INTO feed (id, indexer_id, name, type, enabled, settings) VALUES (3, 3, 'rss', 'RSS', TRUE, '{"download_type":"NZB"}')`,
		`INSERT INTO feed_cache (feed_id, key, value, ttl) VALUES (1, 'a', 'b', CURRENT_TIMESTAMP)`,
		`INSERT INTO feed_cache (feed_id, key, value, ttl) VALUES (2, 'a', 'b', CURRENT_TIMESTAMP)`,
		`INSERT INTO filter_indexer (filter_id, indexer_id) VALUES (1, 1)`,
		`INSERT INTO filter_indexer (filter_id, indexer_id) VALUES (1, 3)`,
		`INSERT INTO irc_network (id, name, server, port) VALUES (1, 'SynIRC', 'irc.synirc.net', 6697)`,
		`INSERT INTO irc_channel (id, name, network_id) VALUES (1, '#curupira-anuncios', 1)`,
	)
}

func validateUsenetRemoved(db *sql.DB, t *testing.T) {
	assert.Equal(t, []int{3}, queryIDs(t, db, `SELECT id FROM indexer ORDER BY id`))
	assert.Equal(t, []int{2, 3}, queryIDs(t, db, `SELECT id FROM feed ORDER BY id`))
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT feed_id FROM feed_cache ORDER BY feed_id`))
	assert.Equal(t, []int{3}, queryIDs(t, db, `SELECT indexer_id FROM filter_indexer ORDER BY indexer_id`))
	assert.Empty(t, queryIDs(t, db, `SELECT id FROM irc_channel`))
	assert.Empty(t, queryIDs(t, db, `SELECT id FROM irc_network`))

	var settings string
	require.NoError(t, db.QueryRow(`SELECT settings FROM feed WHERE id = 3`).Scan(&settings))
	assert.Equal(t, `{"download_type":"TORRENT"}`, settings)
}

func setupNotificationSenderData(db *sql.DB) error {
	return execAll(db,
		`INSERT INTO filter (id, name) VALUES (1, 'movies')`,
		`INSERT INTO notification (id, name, type, enabled, webhook) VALUES (10, 'discord', 'DISCORD', TRUE, 'https://discord.example')`,
		`INSERT INTO notification (id, name, type, enabled, webhook) VALUES (11, 'hook', 'WEBHOOK', TRUE, 'https://hook.example')`,
		`INSERT INTO filter_notification (filter_id, notification_id) VALUES (1, 10)`,
		`INSERT INTO filter_notification (filter_id, notification_id) VALUES (1, 11)`,
	)
}

func validateNotificationSendersRemoved(db *sql.DB, t *testing.T) {
	assert.NotContains(t, queryIDs(t, db, `SELECT id FROM notification ORDER BY id`), 10)
	assert.Contains(t, queryIDs(t, db, `SELECT id FROM notification ORDER BY id`), 11)
	assert.Equal(t, []int{11}, queryIDs(t, db, `SELECT notification_id FROM filter_notification ORDER BY notification_id`))

	_, err := db.Exec(`SELECT token FROM notification`)
	assert.Error(t, err, "notification.token must be dropped")
}

func setupListSourceData(db *sql.DB) error {
	return execAll(db,
		`INSERT INTO filter (id, name) VALUES (1, 'movies')`,
		`INSERT INTO list (id, name, enabled, type, url) VALUES (1, 'mdblist', TRUE, 'MDBLIST', 'https://mdblist.example')`,
		`INSERT INTO list (id, name, enabled, type, url) VALUES (2, 'trakt', TRUE, 'TRAKT', 'https://trakt.example')`,
		`INSERT INTO list_filter (list_id, filter_id) VALUES (1, 1)`,
		`INSERT INTO list_filter (list_id, filter_id) VALUES (2, 1)`,
	)
}

func validateListSourcesRemoved(db *sql.DB, t *testing.T) {
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT id FROM list ORDER BY id`))
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT list_id FROM list_filter ORDER BY list_id`))

	_, err := db.Exec(`SELECT include_year FROM list`)
	assert.Error(t, err, "list.include_year must be dropped")
}

func setupArrData(db *sql.DB) error {
	return execAll(db,
		`INSERT INTO filter (id, name) VALUES (1, 'movies')`,
		`INSERT INTO client (id, name, type, enabled, host) VALUES (1, 'radarr', 'RADARR', TRUE, 'http://radarr')`,
		`INSERT INTO action (id, name, type, enabled, client_id, filter_id) VALUES (1, 'radarr', 'RADARR', TRUE, 1, 1)`,
		`INSERT INTO action (id, name, type, enabled, webhook_host, filter_id) VALUES (2, 'hook', 'WEBHOOK', TRUE, 'https://hook.example', 1)`,
		`INSERT INTO "release" (id, torrent_name) VALUES (1, 'Test.Release-GRP')`,
		`INSERT INTO release_action_status (id, status, action, action_id, type, filter_id, release_id) VALUES (1, 'PUSH_APPROVED', 'radarr', 1, 'RADARR', 1, 1)`,
		`INSERT INTO release_action_status (id, status, action, action_id, type, filter_id, release_id) VALUES (2, 'PUSH_APPROVED', 'hook', 2, 'WEBHOOK', 1, 1)`,
		`INSERT INTO list (id, name, enabled, type, client_id) VALUES (1, 'radarr', TRUE, 'RADARR', 1)`,
		`INSERT INTO list (id, name, enabled, type, url) VALUES (2, 'trakt', TRUE, 'TRAKT', 'https://trakt.example')`,
		`INSERT INTO list_filter (list_id, filter_id) VALUES (1, 1)`,
		`INSERT INTO list_filter (list_id, filter_id) VALUES (2, 1)`,
	)
}

func validateArrRemoved(db *sql.DB, t *testing.T) {
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT id FROM action ORDER BY id`))
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT id FROM list ORDER BY id`))
	assert.Equal(t, []int{2}, queryIDs(t, db, `SELECT list_id FROM list_filter ORDER BY list_id`), "only the remaining list keeps its filter link")

	var arrHistoryActionID sql.NullInt64
	require.NoError(t, db.QueryRow(`SELECT action_id FROM release_action_status WHERE id = 1`).Scan(&arrHistoryActionID))
	assert.False(t, arrHistoryActionID.Valid, "release history is kept but unlinked from the deleted action")

	var webhookHistoryActionID int
	require.NoError(t, db.QueryRow(`SELECT action_id FROM release_action_status WHERE id = 2`).Scan(&webhookHistoryActionID))
	assert.Equal(t, 2, webhookHistoryActionID)

	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'client'`).Scan(&count))
	assert.Equal(t, 0, count, "client table must be dropped")
}

func setupUpdateCheckerData(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO notification (name, type, enabled, events, webhook) VALUES
		('Only update', 'WEBHOOK', TRUE, '{APP_UPDATE_AVAILABLE}', 'https://hook.example/1'),
		('Quoted', 'WEBHOOK', TRUE, '{"PUSH_ERROR","APP_UPDATE_AVAILABLE","IRC_HEALTHY"}', 'https://hook.example/2'),
		('Untouched', 'WEBHOOK', TRUE, '{PUSH_APPROVED}', 'https://hook.example/3')`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO notification_inbox (event, title, message) VALUES
		('APP_UPDATE_AVAILABLE', 'New update available!', 'v1.90.0'),
		('PUSH_ERROR', 'Push error', 'failed')`)

	return err
}

func validateUpdateCheckerRemoved(db *sql.DB, t *testing.T) {
	want := map[string][]string{
		"Built-in":    {"PUSH_ERROR", "IRC_DISCONNECTED"},
		"Only update": {},
		"Quoted":      {"PUSH_ERROR", "IRC_HEALTHY"},
		"Untouched":   {"PUSH_APPROVED"},
	}
	for name, wantEvents := range want {
		var events []string
		err := db.QueryRow(`SELECT events FROM notification WHERE name = $1`, name).Scan(pq.Array(&events))
		require.NoError(t, err, name)
		if len(wantEvents) == 0 {
			assert.Empty(t, events, name)
			continue
		}
		assert.Equal(t, wantEvents, events, name)
	}

	var inbox []string
	rows, err := db.Query(`SELECT event FROM notification_inbox ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var event string
		require.NoError(t, rows.Scan(&event))
		inbox = append(inbox, event)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, []string{"PUSH_ERROR"}, inbox)
}

func setupExecData(db *sql.DB) error {
	stmts := []string{
		`INSERT INTO filter (id, name) VALUES (1, 'exec filter')`,
		`INSERT INTO action (id, name, type, enabled, exec_cmd, exec_args, filter_id) VALUES (1, 'script', 'EXEC', TRUE, '/bin/true', '{{ .TorrentName }}', 1)`,
		`INSERT INTO action (id, name, type, enabled, webhook_host, filter_id) VALUES (2, 'hook', 'WEBHOOK', TRUE, 'https://hook.example', 1)`,
		`INSERT INTO "release" (id, torrent_name) VALUES (1, 'Test.Release-GRP')`,
		`INSERT INTO release_action_status (id, status, action, action_id, type, filter_id, release_id) VALUES (1, 'PUSH_APPROVED', 'script', 1, 'EXEC', 1, 1)`,
		`INSERT INTO release_action_status (id, status, action, action_id, type, filter_id, release_id) VALUES (2, 'PUSH_APPROVED', 'hook', 2, 'WEBHOOK', 1, 1)`,
		`INSERT INTO filter_external (id, name, type, enabled, exec_cmd, exec_args, exec_expect_status, filter_id) VALUES (1, 'script', 'EXEC', TRUE, '/bin/true', '', 0, 1)`,
		`INSERT INTO filter_external (id, name, type, enabled, webhook_host, webhook_expect_status, filter_id) VALUES (2, 'hook', 'WEBHOOK', TRUE, 'https://hook.example', 200, 1)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func validateExecRemoved(db *sql.DB, t *testing.T) {
	var actions []int
	rows, err := db.Query(`SELECT id FROM action ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var id int
		require.NoError(t, rows.Scan(&id))
		actions = append(actions, id)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, []int{2}, actions)

	var externals int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM filter_external WHERE type = 'EXEC'`).Scan(&externals))
	assert.Equal(t, 0, externals)
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM filter_external WHERE type = 'WEBHOOK'`).Scan(&externals))
	assert.Equal(t, 1, externals)

	var execHistoryActionID sql.NullInt64
	require.NoError(t, db.QueryRow(`SELECT action_id FROM release_action_status WHERE id = 1`).Scan(&execHistoryActionID))
	assert.False(t, execHistoryActionID.Valid, "release history is kept but unlinked from the deleted action")

	var webhookHistoryActionID int
	require.NoError(t, db.QueryRow(`SELECT action_id FROM release_action_status WHERE id = 2`).Scan(&webhookHistoryActionID))
	assert.Equal(t, 2, webhookHistoryActionID)

	_, err = db.Exec(`SELECT exec_cmd FROM action`)
	assert.Error(t, err, "action.exec_cmd must be dropped")
	_, err = db.Exec(`SELECT exec_expect_status FROM filter_external`)
	assert.Error(t, err, "filter_external.exec_expect_status must be dropped")
}

func setupAPIKeyData(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO api_key (name, key) VALUES ('script', 'abc123')`)

	return err
}

func validateAPIKeysRemoved(db *sql.DB, t *testing.T) {
	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'api_key'`).Scan(&count))
	assert.Equal(t, 0, count, "api_key table must be dropped")
}

func setupLoginData(db *sql.DB) error {
	if _, err := db.Exec(`INSERT INTO users (username, password) VALUES ('admin', 'hash')`); err != nil {
		return err
	}

	_, err := db.Exec(`INSERT INTO sessions (token, data, expiry) VALUES ('token', x'00', 0)`)

	return err
}

func validateLoginRemoved(db *sql.DB, t *testing.T) {
	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name IN ('users', 'sessions', 'sessions_expiry_idx')`).Scan(&count))
	assert.Equal(t, 0, count, "users, sessions and the sessions index must be dropped")
}
