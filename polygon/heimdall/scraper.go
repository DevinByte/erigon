// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package heimdall

import (
	"context"
	"time"

	"github.com/ledgerwatch/erigon-lib/log/v3"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/polygon/polygoncommon"
)

type scraper[TEntity Entity] struct {
	store EntityStore[TEntity]

	fetcher   entityFetcher[TEntity]
	pollDelay time.Duration

	observers *polygoncommon.Observers[[]TEntity]
	syncEvent *polygoncommon.EventNotifier

	logger log.Logger
}

func newScrapper[TEntity Entity](
	store EntityStore[TEntity],
	fetcher entityFetcher[TEntity],
	pollDelay time.Duration,
	logger log.Logger,
) *scraper[TEntity] {
	return &scraper[TEntity]{
		store: store,

		fetcher:   fetcher,
		pollDelay: pollDelay,

		observers: polygoncommon.NewObservers[[]TEntity](),
		syncEvent: polygoncommon.NewEventNotifier(),

		logger: logger,
	}
}

func (s *scraper[TEntity]) Run(ctx context.Context) error {
	defer s.store.Close()
	if err := s.store.Prepare(ctx); err != nil {
		return err
	}

	for ctx.Err() == nil {
		lastKnownId, hasLastKnownId, err := s.store.GetLastEntityId(ctx)
		if err != nil {
			return err
		}

		idRange, err := s.fetcher.FetchEntityIdRange(ctx)
		if err != nil {
			return err
		}

		if hasLastKnownId {
			idRange.Start = max(idRange.Start, lastKnownId+1)
		}

		if idRange.Start > idRange.End {
			s.syncEvent.SetAndBroadcast()
			if err := libcommon.Sleep(ctx, s.pollDelay); err != nil {
				s.syncEvent.Reset()
				return err
			}
		} else {
			entities, err := s.fetcher.FetchEntitiesRange(ctx, idRange)
			if err != nil {
				return err
			}

			for i, entity := range entities {
				if err = s.store.PutEntity(ctx, idRange.Start+uint64(i), entity); err != nil {
					return err
				}
			}

			go s.observers.Notify(entities)
		}
	}
	return ctx.Err()
}

func (s *scraper[TEntity]) RegisterObserver(observer func([]TEntity)) polygoncommon.UnregisterFunc {
	return s.observers.Register(observer)
}

func (s *scraper[TEntity]) Synchronize(ctx context.Context) {
	s.syncEvent.Wait(ctx)
}
