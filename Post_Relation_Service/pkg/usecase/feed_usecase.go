package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/events"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
)

type FeedUsecase struct {
	repo  interfacesRepository.PostRelationRepository
	redis interfacesRepository.RedisRepository
	cfg   *config.Config
	log   logger.Logger
}

func NewFeedUsecase(
	repo interfacesRepository.PostRelationRepository,
	redis interfacesRepository.RedisRepository,
	cfg *config.Config,
	log logger.Logger,
) interfacesUsecase.FeedUsecase {

	return &FeedUsecase{
		repo:  repo,
		redis: redis,
		cfg:   cfg,
		log:   log,
	}
}

func (f *FeedUsecase) ProcessPostCreated(event events.PostCreatedEvent) error {

	ctx := context.Background()

	// 1. Parse celebrity threshold
	celebfollowcount, err := strconv.Atoi(f.cfg.CelebrityFollowCount)
	if err != nil {
		return fmt.Errorf("invalid celeb follow count: %w", err)
	}

	// 2. Fetch followers
	followers, err := f.repo.FetchFollowersUserIds(event.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		f.log.Error("failed to fetch followers",
			logger.Field{Key: "error", Value: err})
		return err
	}

	//⚠️ OPTIONAL (recommended later) — idempotency check
	// idempotencyKey := fmt.Sprintf("fanout:post:%d", event.PostID)
	// ok, _ := f.redis.SetNX(ctx, idempotencyKey, 1, time.Hour)
	// if !ok {
	// 	return nil // already processed
	// }

	// 3. Celebrity path (pull model)
	if len(followers) > celebfollowcount {

		key := fmt.Sprintf("celeb:posts:%d", event.UserID)

		pipe := f.redis.Pipeline()

		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(event.CreatedAt),
			Member: event.PostID,
		})

		pipe.ZRemRangeByRank(ctx, key, 0, -51)

		pipe.Expire(ctx, key, 7*24*time.Hour)

		_, err := pipe.Exec(ctx)
		if err != nil {
			f.log.Warn("celeb redis pipeline failed",
				logger.Field{Key: "error", Value: err})
			return err
		}

		return nil
	}

	// 4. Push model (fan-out)
	pipe := f.redis.Pipeline()

	// include self
	followers = append(followers, responsemodels.FollowerIds{
		FollowerID: event.UserID,
	})

	for i, fID := range followers {

		feedKey := fmt.Sprintf("feed:user:%d", fID.FollowerID)

		pipe.ZAdd(ctx, feedKey, redis.Z{
			Score:  float64(event.CreatedAt),
			Member: event.PostID,
		})

		pipe.ZRemRangeByRank(ctx, feedKey, 0, -101)

		pipe.Expire(ctx, feedKey, 72*time.Hour)

		// batch every 500
		if (i+1)%500 == 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				f.log.Error("redis pipeline batch failed",
					logger.Field{Key: "error", Value: err})
				return err
			}
			pipe = f.redis.Pipeline()
		}
	}

	// final flush
	if pipe.Len() > 0 {
		_, err := pipe.Exec(ctx)
		if err != nil {
			f.log.Error("final redis pipeline failed",
				logger.Field{Key: "error", Value: err})
			return err
		}
	}

	return nil
}
