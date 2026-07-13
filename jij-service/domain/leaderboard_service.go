// Copyright (c) 2026 Guillaume Evrard
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package domain

import (
	"context"
	"jij-service/dal"
)

type LeaderboardEntry struct {
	ProfileID string
	Score     int
	Rank      int
}

type LeaderboardService struct {
	LeaderboardDal dal.LeaderboardDal
}

func (s LeaderboardService) SubmitScore(ctx context.Context, leaderboardName string, profileID string, score int) error {
	entity := dal.LeaderboardEntryEntity{
		LeaderboardName: leaderboardName,
		ProfileID:       profileID,
		Score:           score,
	}
	return s.LeaderboardDal.SubmitScore(ctx, &entity)
}

func (s LeaderboardService) GetLeaderboard(ctx context.Context, leaderboardName string, limit int) ([]LeaderboardEntry, error) {
	entities, err := s.LeaderboardDal.Top(ctx, leaderboardName, limit)
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, 0, len(entities))
	for i, entity := range entities {
		entries = append(entries, LeaderboardEntry{
			ProfileID: entity.ProfileID,
			Score:     entity.Score,
			Rank:      i + 1,
		})
	}
	return entries, nil
}
