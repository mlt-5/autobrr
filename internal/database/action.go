// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package database

import (
	"context"
	"database/sql"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog"
)

type ActionRepo struct {
	log zerolog.Logger
	db  *DB
}

func NewActionRepo(log zerolog.Logger, db *DB) *ActionRepo {
	return &ActionRepo{
		log: log.With().Str("repo", "action").Logger(),
		db:  db,
	}
}

func (r *ActionRepo) FindByFilterID(ctx context.Context, filterID int, active *bool) ([]*domain.Action, error) {
	queryBuilder := r.db.squirrel.
		Select(
			"a.id",
			"a.name",
			"a.type",
			"a.enabled",
			"a.exec_cmd",
			"a.exec_args",
			"a.watch_folder",
			"a.webhook_host",
			"a.webhook_type",
			"a.webhook_method",
			"a.webhook_data",
		).
		From("action a").
		Where(sq.Eq{"a.filter_id": filterID})

	if active != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"enabled": *active})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	rows, err := r.db.Handler.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	defer rows.Close()

	actions := make([]*domain.Action, 0)
	for rows.Next() {
		var a domain.Action

		var execCmd, execArgs, watchFolder, webhookHost, webhookType, webhookMethod, webhookData sql.NullString

		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Enabled, &execCmd, &execArgs, &watchFolder, &webhookHost, &webhookType, &webhookMethod, &webhookData); err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}

		a.ExecCmd = execCmd.String
		a.ExecArgs = execArgs.String
		a.WatchFolder = watchFolder.String

		a.WebhookHost = webhookHost.String
		a.WebhookType = webhookType.String
		a.WebhookMethod = webhookMethod.String
		a.WebhookData = webhookData.String

		actions = append(actions, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row error")
	}

	return actions, nil
}

func (r *ActionRepo) List(ctx context.Context) ([]domain.Action, error) {
	queryBuilder := r.db.squirrel.
		Select(
			"id",
			"name",
			"type",
			"enabled",
			"exec_cmd",
			"exec_args",
			"watch_folder",
			"webhook_host",
			"webhook_type",
			"webhook_method",
			"webhook_data",
		).
		From("action")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	rows, err := r.db.Handler.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	defer rows.Close()

	actions := make([]domain.Action, 0)
	for rows.Next() {
		var a domain.Action

		var execCmd, execArgs, watchFolder, webhookHost, webhookType, webhookMethod, webhookData sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Enabled, &execCmd, &execArgs, &watchFolder, &webhookHost, &webhookType, &webhookMethod, &webhookData); err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}

		a.WebhookHost = webhookHost.String
		a.WebhookType = webhookType.String
		a.WebhookMethod = webhookMethod.String
		a.WebhookData = webhookData.String

		actions = append(actions, a)

		if err := rows.Err(); err != nil {
			return nil, errors.Wrap(err, "rows error")
		}
	}

	return actions, nil
}

func (r *ActionRepo) Get(ctx context.Context, req *domain.GetActionRequest) (*domain.Action, error) {
	queryBuilder := r.db.squirrel.
		Select(
			"id",
			"name",
			"type",
			"enabled",
			"exec_cmd",
			"exec_args",
			"watch_folder",
			"webhook_host",
			"webhook_type",
			"webhook_method",
			"webhook_data",
			"filter_id",
		).
		From("action").
		Where(sq.Eq{"id": req.Id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	row := r.db.Handler.QueryRowContext(ctx, query, args...)
	if err := row.Err(); err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	var a domain.Action

	var execCmd, execArgs, watchFolder, webhookHost, webhookType, webhookMethod, webhookData sql.NullString
	var filterID sql.NullInt32

	if err := row.Scan(&a.ID, &a.Name, &a.Type, &a.Enabled, &execCmd, &execArgs, &watchFolder, &webhookHost, &webhookType, &webhookMethod, &webhookData, &filterID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}

		return nil, errors.Wrap(err, "error scanning row")
	}

	a.ExecCmd = execCmd.String
	a.ExecArgs = execArgs.String
	a.WatchFolder = watchFolder.String

	a.WebhookHost = webhookHost.String
	a.WebhookType = webhookType.String
	a.WebhookMethod = webhookMethod.String
	a.WebhookData = webhookData.String

	a.FilterID = int(filterID.Int32)

	return &a, nil
}

func (r *ActionRepo) Delete(ctx context.Context, req *domain.DeleteActionRequest) error {
	queryBuilder := r.db.squirrel.
		Delete("action").
		Where(sq.Eq{"id": req.ActionId})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	result, err := r.db.Handler.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return errors.Wrap(err, "error getting rows affected")
	} else if rowsAffected == 0 {
		return domain.ErrRecordNotFound
	}

	r.log.Debug().Int("action_id", req.ActionId).Msg("action delete")

	return nil
}

func (r *ActionRepo) DeleteByFilterID(ctx context.Context, filterID int) error {
	queryBuilder := r.db.squirrel.
		Delete("action").
		Where(sq.Eq{"filter_id": filterID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	_, err = r.db.Handler.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}

	r.log.Debug().Int("filter_id", filterID).Msg("action delete by filter")

	return nil
}

func (r *ActionRepo) Store(ctx context.Context, action *domain.Action) error {
	queryBuilder := r.db.squirrel.
		Insert("action").
		Columns(
			"name",
			"type",
			"enabled",
			"exec_cmd",
			"exec_args",
			"watch_folder",
			"webhook_host",
			"webhook_type",
			"webhook_method",
			"webhook_data",
			"filter_id",
		).
		Values(
			action.Name,
			action.Type,
			action.Enabled,
			toNullString(action.ExecCmd),
			toNullString(action.ExecArgs),
			toNullString(action.WatchFolder),
			toNullString(action.WebhookHost),
			toNullString(action.WebhookType),
			toNullString(action.WebhookMethod),
			toNullString(action.WebhookData),
			toNullInt32(int32(action.FilterID)),
		).
		Suffix("RETURNING id").RunWith(r.db.Handler)

	// return values
	var retID int64

	if err := queryBuilder.QueryRowContext(ctx).Scan(&retID); err != nil {
		return errors.Wrap(err, "error executing query")
	}

	action.ID = int(retID)

	r.log.Debug().Int64("action_id", retID).Msg("action store: added new")

	return nil
}

func (r *ActionRepo) Update(ctx context.Context, action domain.Action) (*domain.Action, error) {
	queryBuilder := r.db.squirrel.
		Update("action").
		Set("name", action.Name).
		Set("type", action.Type).
		Set("enabled", action.Enabled).
		Set("exec_cmd", toNullString(action.ExecCmd)).
		Set("exec_args", toNullString(action.ExecArgs)).
		Set("watch_folder", toNullString(action.WatchFolder)).
		Set("webhook_host", toNullString(action.WebhookHost)).
		Set("webhook_type", toNullString(action.WebhookType)).
		Set("webhook_method", toNullString(action.WebhookMethod)).
		Set("webhook_data", toNullString(action.WebhookData)).
		Set("filter_id", toNullInt32(int32(action.FilterID))).
		Where(sq.Eq{"id": action.ID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	result, err := r.db.Handler.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return nil, errors.Wrap(err, "error getting rows affected")
	} else if rowsAffected == 0 {
		return nil, domain.ErrRecordNotFound
	}

	r.log.Debug().Int("action_id", action.ID).Msg("action update")

	return &action, nil
}

func (r *ActionRepo) StoreFilterActions(ctx context.Context, filterID int64, actions []*domain.Action) ([]*domain.Action, error) {
	tx, err := r.db.Handler.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error begin transaction")
	}

	defer tx.Rollback()

	for _, action := range actions {
		if action.ID > 0 {
			queryBuilder := r.db.squirrel.
				Update("action").
				Set("name", action.Name).
				Set("type", action.Type).
				Set("enabled", action.Enabled).
				Set("exec_cmd", toNullString(action.ExecCmd)).
				Set("exec_args", toNullString(action.ExecArgs)).
				Set("watch_folder", toNullString(action.WatchFolder)).
				Set("webhook_host", toNullString(action.WebhookHost)).
				Set("webhook_type", toNullString(action.WebhookType)).
				Set("webhook_method", toNullString(action.WebhookMethod)).
				Set("webhook_data", toNullString(action.WebhookData)).
				Set("filter_id", toNullInt64(filterID)).
				Where(sq.Eq{"id": action.ID})

			query, args, err := queryBuilder.ToSql()
			if err != nil {
				return nil, errors.Wrap(err, "error building query")
			}

			if _, err := tx.ExecContext(ctx, query, args...); err != nil {
				return nil, errors.Wrap(err, "error executing query")
			}

			r.log.Trace().Int("action_id", action.ID).Msg("action store filter actions: update")

		} else {
			queryBuilder := r.db.squirrel.
				Insert("action").
				Columns(
					"name",
					"type",
					"enabled",
					"exec_cmd",
					"exec_args",
					"watch_folder",
					"webhook_host",
					"webhook_type",
					"webhook_method",
					"webhook_data",
					"filter_id",
				).
				Values(
					action.Name,
					action.Type,
					action.Enabled,
					toNullString(action.ExecCmd),
					toNullString(action.ExecArgs),
					toNullString(action.WatchFolder),
					toNullString(action.WebhookHost),
					toNullString(action.WebhookType),
					toNullString(action.WebhookMethod),
					toNullString(action.WebhookData),
					toNullInt64(filterID),
				).
				Suffix("RETURNING id").RunWith(tx)

			// return values
			var retID int

			if err := queryBuilder.QueryRowContext(ctx).Scan(&retID); err != nil {
				return nil, errors.Wrap(err, "error executing query")
			}

			action.ID = retID

			r.log.Trace().Int("action_id", action.ID).Msg("action store filter actions: store")
		}

		r.log.Debug().Str("action", action.Name).Str("action_type", string(action.Type)).Int64("filter_id", filterID).Msg("store filter action")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "error updating filter actions")
	}

	return actions, nil
}

func (r *ActionRepo) ToggleEnabled(actionID int) error {
	queryBuilder := r.db.squirrel.
		Update("action").
		Set("enabled", sq.Expr("NOT enabled")).
		Where(sq.Eq{"id": actionID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	result, err := r.db.Handler.Exec(query, args...)
	if err != nil {
		return errors.Wrap(err, "error executing query")
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return errors.Wrap(err, "error getting rows affected")
	} else if rowsAffected == 0 {
		return domain.ErrRecordNotFound
	}

	r.log.Debug().Int("action_id", actionID).Msg("action toggle enabled")

	return nil
}
