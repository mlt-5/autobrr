// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package database

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/logger"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type testDB struct {
	db      *DB
	cleanup cleanupFunc
}

type cleanupFunc func()

var testDBs map[string]*testDB

func setupSqliteForTest() (*DB, cleanupFunc) {
	cfg := &domain.Config{
		LogLevel:    "ERROR",
		DatabaseDSN: ":memory:",
	}

	// Init a new logger
	dbLogger := logger.New(cfg, nil)

	// Initialize a new DB connection
	db, err := NewDB(cfg, dbLogger)
	if err != nil {
		log.Fatalf("Could not create database: %v", err)
	}

	// Open the database connection
	if err := db.Open(); err != nil {
		log.Fatalf("Could not open db connection: %v", err)
	}

	// migrate db
	if err = db.migrateSQLite(); err != nil {
		log.Fatalf("Could not migrate sqlite database: %q", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	testDBInstance := &testDB{
		db:      db,
		cleanup: cleanup,
	}

	testDBs["sqlite"] = testDBInstance

	return db, cleanup
}

func setupLoggerForTest() zerolog.Logger {
	return logger.New(&domain.Config{LogLevel: "ERROR"}, nil)
}

func TestPingDatabase(t *testing.T) {
	// Setup database
	for dbType, testDb := range testDBs {
		db := testDb.db
		t.Run(fmt.Sprintf("Test_db_ping [%s]", dbType), func(t *testing.T) {
			err := db.Ping()
			assert.NoError(t, err, "Database should be reachable")
		})
	}
}

func TestMain(m *testing.M) {
	if err := os.Setenv("IS_TEST_ENV", "true"); err != nil {
		log.Fatalf("Could not set env variable: %v", err)
	}

	testDBs = make(map[string]*testDB)

	fmt.Println("setup")

	setupSqliteForTest()

	fmt.Println("running tests")

	//Run tests
	code := m.Run()

	fmt.Println("teardown")
	for _, d := range testDBs {
		d.cleanup()
	}

	if err := os.Setenv("IS_TEST_ENV", "false"); err != nil {
		log.Fatalf("Could not set env variable: %v", err)
	}

	os.Exit(code)
}
