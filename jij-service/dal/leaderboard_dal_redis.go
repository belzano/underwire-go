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

package dal

import (
	"context"
	"fmt"
	"jij-service/configuration"

	"github.com/redis/go-redis/v9"
)

type LeaderboardDalRedis struct {
	client *redis.Client
}

func NewLeaderboardDalRedis(redisConfiguration configuration.RedisConfiguration) *LeaderboardDalRedis {
	client := redis.NewClient(&redis.Options{
		Addr: redisConfiguration.Addr,
	})

	return &LeaderboardDalRedis{
		client: client,
	}
}

func leaderboardKey(leaderboardName string) string {
	return "leaderboard:" + leaderboardName
}

func (r *LeaderboardDalRedis) SubmitScore(ctx context.Context, entity *LeaderboardEntryEntity) error {
	_, err := r.client.ZAddGT(ctx, leaderboardKey(entity.LeaderboardName), redis.Z{
		Score:  float64(entity.Score),
		Member: entity.ProfileID,
	}).Result()
	if err != nil {
		return fmt.Errorf("submit score failure : %v", err)
	}
	return nil
}

func (r *LeaderboardDalRedis) Top(ctx context.Context, leaderboardName string, limit int) ([]LeaderboardEntryEntity, error) {
	results, err := r.client.ZRevRangeWithScores(ctx, leaderboardKey(leaderboardName), 0, int64(limit)-1).Result()
	if err != nil {
		return nil, fmt.Errorf("retrieval failure : %v", err)
	}

	entries := make([]LeaderboardEntryEntity, 0, len(results))
	for _, member := range results {
		entries = append(entries, LeaderboardEntryEntity{
			LeaderboardName: leaderboardName,
			ProfileID:       fmt.Sprintf("%v", member.Member),
			Score:           int(member.Score),
		})
	}
	return entries, nil
}
