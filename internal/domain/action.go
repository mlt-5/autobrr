// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"strings"

	"github.com/autobrr/autobrr/pkg/errors"
)

type Action struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Type           ActionType `json:"type"`
	Enabled        bool       `json:"enabled"`
	ExecCmd        string     `json:"exec_cmd,omitempty"`
	ExecArgs       string     `json:"exec_args,omitempty"`
	WatchFolder    string     `json:"watch_folder,omitempty"`
	WebhookHost    string     `json:"webhook_host,omitempty"`
	WebhookType    string     `json:"webhook_type,omitempty"`
	WebhookMethod  string     `json:"webhook_method,omitempty"`
	WebhookData    string     `json:"webhook_data,omitempty"`
	WebhookHeaders []string   `json:"webhook_headers,omitempty"`
	FilterID       int        `json:"filter_id,omitempty"`
}

// NeedsTorrentDownloaded check if the action type uploads the torrent itself.
// These are resolved up front so a release with several actions fetches the
// torrent from the indexer once, instead of once per action.
func (a *Action) NeedsTorrentDownloaded() bool {
	switch a.Type {
	case ActionTypeWatchFolder:
		return true
	default:
		return false
	}
}

// macroFields returns the action fields that get macro expanded and could
// reference the torrent file. Keep in sync with ParseMacros.
func (a *Action) macroFields() []string {
	return []string{a.ExecArgs, a.WatchFolder, a.WebhookData}
}

// CheckMacrosNeedTorrentTmpFile check if macros need the torrent written to disk.
// Only the path macros need a real file, everything else works off the raw bytes.
func (a *Action) CheckMacrosNeedTorrentTmpFile(release *Release) bool {
	if release.TorrentTmpFile != "" {
		return false
	}

	return containsAnyMacro(a.macroFields(), "TorrentPathName", "TorrentTmpFile")
}

// CheckMacrosNeedRawDataBytes check if macros need the torrent downloaded into memory.
// This is a superset of CheckMacrosNeedTorrentTmpFile since a tmp file can only be
// written once we hold the contents.
func (a *Action) CheckMacrosNeedRawDataBytes(release *Release) bool {
	if len(release.TorrentDataRawBytes) != 0 {
		return false
	}

	if a.Type == ActionTypeWatchFolder {
		return true
	}

	return containsAnyMacro(a.macroFields(), "TorrentPathName", "TorrentTmpFile", "TorrentDataRawBytes", "TorrentHash")
}

func containsAnyMacro(fields []string, macros ...string) bool {
	for _, field := range fields {
		if field == "" {
			continue
		}

		for _, macro := range macros {
			if strings.Contains(field, macro) {
				return true
			}
		}
	}

	return false
}

// ParseMacros parse all macros on action
func (a *Action) ParseMacros(release *Release) error {
	var err error

	m := NewMacro(*release)

	a.ExecArgs, err = m.Parse(a.ExecArgs)
	if err != nil {
		return errors.Wrap(err, "could not parse exec args")
	}

	a.WatchFolder, err = m.Parse(a.WatchFolder)
	if err != nil {
		return errors.Wrap(err, "could not parse watch folder")
	}

	a.WebhookData, err = m.Parse(a.WebhookData)
	if err != nil {
		return errors.Wrap(err, "could not parse webhook_data")
	}

	return nil
}

type ActionType string

const (
	ActionTypeTest        ActionType = "TEST"
	ActionTypeExec        ActionType = "EXEC"
	ActionTypeWatchFolder ActionType = "WATCH_FOLDER"
	ActionTypeWebhook     ActionType = "WEBHOOK"
)

type GetActionRequest struct {
	Id int
}

type DeleteActionRequest struct {
	ActionId int
}
