// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package downloader

import (
	"context"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/arr/lidarr"
	"github.com/autobrr/autobrr/pkg/arr/radarr"
	"github.com/autobrr/autobrr/pkg/arr/readarr"
	"github.com/autobrr/autobrr/pkg/arr/sonarr"
	"github.com/autobrr/autobrr/pkg/arr/sportarr"
	"github.com/autobrr/autobrr/pkg/arr/whisparr"
	"github.com/autobrr/autobrr/pkg/errors"
)

func (s *Service) testConnection(ctx context.Context, cfg *domain.Downloader) error {
	instance, err := s.initInstance(cfg)
	if err != nil {
		return err
	}

	switch cfg.Type {

	case domain.DownloaderTypeRadarr:
		return s.testConnectionRadarr(ctx, instance)

	case domain.DownloaderTypeSonarr:
		return s.testConnectionSonarr(ctx, instance)

	case domain.DownloaderTypeLidarr:
		return s.testConnectionLidarr(ctx, instance)

	case domain.DownloaderTypeWhisparr, domain.DownloaderTypeWhisparrV3:
		return s.testConnectionWhisparr(ctx, instance)

	case domain.DownloaderTypeReadarr:
		return s.testConnectionReadarr(ctx, instance)

	case domain.DownloaderTypeSportarr:
		return s.testConnectionSportarr(ctx, instance)

	default:
		return errors.New("unsupported client: %s", cfg.Type)
	}
}

func (s *Service) testConnectionRadarr(ctx context.Context, instance *Instance) error {
	cfg := instance.Config()

	client, err := instance.ClientAs[*radarr.Client]()
	if err != nil {
		return err
	}

	if _, err := client.Test(ctx); err != nil {
		return errors.Wrap(err, "radarr: connection test failed: %v", cfg.Host)
	}

	s.log.Debug().Msg("test client connection for Radarr: success")

	return nil
}

func (s *Service) testConnectionSportarr(ctx context.Context, instance *Instance) error {
	cfg := instance.Config()

	client, err := instance.ClientAs[*sportarr.Client]()
	if err != nil {
		return err
	}

	status, err := client.Test(ctx)
	if err != nil {
		return errors.Wrap(err, "sportarr: connection test failed: %v", cfg.Host)
	}

	// The native release/push route ships with 4.0.1024; older versions
	// answer system/status fine but 404 the push, so fail the test early
	// with an actionable message instead of failing on the first release.
	if !status.SupportsNativeAPI() {
		return errors.New("sportarr: version %s is too old, autobrr needs Sportarr %s or newer", status.Version, sportarr.MinimumVersion)
	}

	s.log.Debug().Str("version", status.Version).Msg("test client connection for Sportarr: success")

	return nil
}

func (s *Service) testConnectionSonarr(ctx context.Context, instance *Instance) error {
	cfg := instance.Config()

	client, err := instance.ClientAs[*sonarr.Client]()
	if err != nil {
		return err
	}

	if _, err := client.Test(ctx); err != nil {
		return errors.Wrap(err, "sonarr: connection test failed: %v", cfg.Host)
	}

	s.log.Debug().Msg("test client connection for Sonarr: success")

	return nil
}

func (s *Service) testConnectionLidarr(ctx context.Context, instance *Instance) error {
	cfg := instance.Config()

	client, err := instance.ClientAs[*lidarr.Client]()
	if err != nil {
		return err
	}

	if _, err := client.Test(ctx); err != nil {
		return errors.Wrap(err, "lidarr: connection test failed: %v", cfg.Host)
	}

	s.log.Debug().Msg("test client connection for Lidarr: success")

	return nil
}

// whisparrVersion maps the client type to the Whisparr major version it talks
// to. v2 is Sonarr based and serves series, v3 is Radarr based and serves
// movies, so the two are separate client types.
func whisparrVersion(clientType domain.DownloaderType) int {
	if clientType == domain.DownloaderTypeWhisparrV3 {
		return whisparr.VersionV3
	}

	return whisparr.VersionV2
}

func (s *Service) testConnectionWhisparr(ctx context.Context, instance *Instance) error {
	cfg := instance.Config()

	client, err := instance.ClientAs[*whisparr.Client]()
	if err != nil {
		return err
	}

	status, err := client.Test(ctx)
	if err != nil {
		return errors.Wrap(err, "whisparr: connection test failed: %v", cfg.Host)
	}

	s.log.Debug().Str("version", status.Version).Msg("test client connection for whisparr: success")

	return nil
}

func (s *Service) testConnectionReadarr(ctx context.Context, instance *Instance) error {
	cfg := instance.Config()

	client, err := instance.ClientAs[*readarr.Client]()
	if err != nil {
		return err
	}

	if _, err := client.Test(ctx); err != nil {
		return errors.Wrap(err, "readarr: connection test failed: %v", cfg.Host)
	}

	s.log.Debug().Msg("test client connection for readarr: success")

	return nil
}
