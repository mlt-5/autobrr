// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

// Package vocab keeps the release-name parser's vocabulary: the custom entries and the removed new_parser list
// entries, stored in the database and applied to the parser whenever they change.
package vocab

import (
	"context"
	"slices"
	"sync"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"

	"github.com/rs/zerolog"
)

type vocabRepo interface {
	List(ctx context.Context) ([]domain.VocabEntry, error)
	Store(ctx context.Context, entry *domain.VocabEntry) error
	Update(ctx context.Context, entry *domain.VocabEntry) error
	Delete(ctx context.Context, id int64) error
	ListRemoved(ctx context.Context) ([]domain.VocabKey, error)
	Remove(ctx context.Context, key domain.VocabKey) error
	Restore(ctx context.Context, key domain.VocabKey) error
}

type Service struct {
	log  zerolog.Logger
	repo vocabRepo

	// m serializes changes, so the parser always matches what is stored
	m sync.Mutex
}

func NewService(log zerolog.Logger, repo vocabRepo) *Service {
	return &Service{
		log:  log.With().Str("module", "vocab").Logger(),
		repo: repo,
	}
}

// Start builds the parser with the stored vocabulary.
func (s *Service) Start(ctx context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()

	custom, removed, err := s.load(ctx)
	if err != nil {
		return err
	}

	return s.use(custom, removed)
}

// List returns every vocabulary entry.
func (s *Service) List(ctx context.Context) ([]domain.VocabEntry, error) {
	custom, removed, err := s.load(ctx)
	if err != nil {
		return nil, err
	}

	return domain.Vocab(custom, removed), nil
}

// Store adds a custom entry and applies it to the parser.
func (s *Service) Store(ctx context.Context, entry *domain.VocabEntry) error {
	entry.ID = 0

	return s.save(ctx, entry, s.repo.Store)
}

// Update changes a custom entry and applies it to the parser.
func (s *Service) Update(ctx context.Context, entry *domain.VocabEntry) error {
	if entry.ID == 0 {
		return domain.ErrRecordNotFound
	}

	return s.save(ctx, entry, s.repo.Update)
}

func (s *Service) save(ctx context.Context, entry *domain.VocabEntry, write func(context.Context, *domain.VocabEntry) error) error {
	entry.Normalize()
	if err := entry.Validate(); err != nil {
		return err
	}

	s.m.Lock()
	defer s.m.Unlock()

	custom, removed, err := s.load(ctx)
	if err != nil {
		return err
	}

	custom, err = withEntry(custom, *entry)
	if err != nil {
		return err
	}

	p, err := domain.NewVocabParser(custom, removed)
	if err != nil {
		return err
	}

	if err := write(ctx, entry); err != nil {
		return err
	}

	entry.Source = domain.VocabSourceCustom
	domain.UseVocabParser(p)

	s.log.Info().Int64("id", entry.ID).Str("kind", entry.Kind).Str("tag", entry.Tag).Msg("saved vocabulary entry")

	return nil
}

// Preview parses name with the stored vocabulary, plus entry when it is given: added, or in place of the stored
// entry with its ID. Nothing is saved.
func (s *Service) Preview(ctx context.Context, name string, entry *domain.VocabEntry) (*domain.VocabPreview, error) {
	custom, removed, err := s.load(ctx)
	if err != nil {
		return nil, err
	}

	if entry != nil {
		entry.Normalize()
		if err := entry.Validate(); err != nil {
			return nil, err
		}

		if custom, err = withEntry(custom, *entry); err != nil {
			return nil, err
		}
	}

	p, err := domain.NewVocabParser(custom, removed)
	if err != nil {
		return nil, err
	}

	preview := domain.PreviewVocab(p, name)

	return &preview, nil
}

// withEntry returns custom with entry added, or replacing the entry with the same ID.
func withEntry(custom []domain.VocabEntry, entry domain.VocabEntry) ([]domain.VocabEntry, error) {
	if slices.ContainsFunc(custom, func(e domain.VocabEntry) bool {
		return e.ID != entry.ID && e.Kind == entry.Kind && e.Tag == entry.Tag
	}) {
		return nil, domain.VocabErrorf("%s %q is already added", entry.Kind, entry.Tag)
	}

	if entry.ID == 0 {
		return append(slices.Clone(custom), entry), nil
	}

	i := slices.IndexFunc(custom, func(e domain.VocabEntry) bool { return e.ID == entry.ID })
	if i < 0 {
		return nil, domain.ErrRecordNotFound
	}

	custom = slices.Clone(custom)
	custom[i] = entry

	return custom, nil
}

// Delete removes a custom entry from the parser.
func (s *Service) Delete(ctx context.Context, id int64) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return s.reload(ctx)
}

// Remove takes a new_parser list entry out of the parser.
func (s *Service) Remove(ctx context.Context, key domain.VocabKey) error {
	if !domain.IsListedVocab(key) {
		return domain.ErrRecordNotFound
	}

	s.m.Lock()
	defer s.m.Unlock()

	if err := s.repo.Remove(ctx, key); err != nil {
		return err
	}

	return s.reload(ctx)
}

// Restore puts a removed new_parser list entry back in the parser.
func (s *Service) Restore(ctx context.Context, key domain.VocabKey) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err := s.repo.Restore(ctx, key); err != nil {
		return err
	}

	return s.reload(ctx)
}

func (s *Service) reload(ctx context.Context) error {
	custom, removed, err := s.load(ctx)
	if err != nil {
		return err
	}

	return s.use(custom, removed)
}

func (s *Service) load(ctx context.Context) ([]domain.VocabEntry, []domain.VocabKey, error) {
	custom, err := s.repo.List(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not list vocabulary")
	}

	removed, err := s.repo.ListRemoved(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not list removed vocabulary")
	}

	return custom, removed, nil
}

func (s *Service) use(custom []domain.VocabEntry, removed []domain.VocabKey) error {
	p, err := domain.NewVocabParser(custom, removed)
	if err != nil {
		return errors.Wrap(err, "could not build parser")
	}

	domain.UseVocabParser(p)

	return nil
}
