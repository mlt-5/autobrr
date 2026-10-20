// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package cli

import (
	"fmt"

	"github.com/autobrr/autobrr/internal/database/tools"
	"github.com/autobrr/autobrr/pkg/errors"

	"github.com/spf13/cobra"
)

func CommandDb() *cobra.Command {
	var command = &cobra.Command{
		Use:   "db",
		Short: "Database maintenance commands",
	}

	command.AddCommand(CommandDbSeed())
	command.AddCommand(CommandDbReset())

	return command
}

func CommandDbSeed() *cobra.Command {
	var command = &cobra.Command{
		Use:     "seed",
		Short:   "Seed the SQLite database",
		Example: "  autobrr db seed --db-path /path/to/autobrr.db --seed-db /path/to/seed",
	}

	var dbPath, seedPath string

	command.Flags().StringVar(&dbPath, "db-path", "", "path to the database file")
	command.Flags().StringVar(&seedPath, "seed-db", "", "path to SQL seed file")
	_ = command.MarkFlagRequired("db-path")
	_ = command.MarkFlagRequired("seed-db")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if err := tools.NewSQLiteSeeder(dbPath, seedPath).Seed(); err != nil {
			return errors.Wrap(err, "could not seed database")
		}

		fmt.Println("Database seeding completed successfully!")

		return nil
	}

	return command
}

func CommandDbReset() *cobra.Command {
	var command = &cobra.Command{
		Use:     "reset",
		Short:   "Reset and reseed the SQLite database",
		Example: "  autobrr db reset --db-path /path/to/autobrr.db --seed-db /path/to/seed",
	}

	var dbPath, seedPath string

	command.Flags().StringVar(&dbPath, "db-path", "", "path to the database file")
	command.Flags().StringVar(&seedPath, "seed-db", "", "path to SQL seed file")
	_ = command.MarkFlagRequired("db-path")
	_ = command.MarkFlagRequired("seed-db")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		if err := tools.NewSQLiteSeeder(dbPath, seedPath).ResetAndSeed(); err != nil {
			return errors.Wrap(err, "could not reset and seed database")
		}

		fmt.Println("Database reset and reseed completed successfully!")

		return nil
	}

	return command
}
