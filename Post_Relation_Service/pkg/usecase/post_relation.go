package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/client"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostRelationUsecase struct {
	PostRelationRepository interfacesRepository.PostRelationRepository
	AuthSubscriptionClient pb.AuthSubscriptionServiceClient
	RedisRepository        interfacesRepository.RedisRepository
	KafkaProducer          interfacesUsecase.KafkaProducer // <--- Add this
	Log                    logger.Logger
	Cfg                    *config.Config
}

var ()

func NewPostRelationUsecase(repository interfacesRepository.PostRelationRepository,
	authSubClient pb.AuthSubscriptionServiceClient,
	redisRepository interfacesRepository.RedisRepository,
	kafkaProducer interfacesUsecase.KafkaProducer,
	log logger.Logger, cfg *config.Config) interfacesUsecase.PostRelationUsecase {
	return &PostRelationUsecase{
		PostRelationRepository: repository,
		AuthSubscriptionClient: authSubClient,
		RedisRepository:        redisRepository,
		KafkaProducer:          kafkaProducer,
		Log:                    log,
		Cfg:                    cfg,
	}
}

func (as *PostRelationUsecase) InsertUserIntoFollowCount(userid uint64) error {
	err := as.PostRelationRepository.InsertUserIntoFollowCount(userid)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}
	return nil
}
func (as *PostRelationUsecase) CreatePost(createPostReq requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error) {
	createPostRes, err := as.PostRelationRepository.CreatePost(createPostReq)
	if err != nil {
		return responsemodels.CreatePostResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", createPostReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon create post",
			logger.Field{Key: "error", Value: err})
	}

	// 2. Start a background process for fan-out
	go func() {
		// followers, _ := as.PostRelationRepository.FetchFollowersUserIds(createPostReq.UserID)
		// for _, fID := range followers {
		//     versionKey := fmt.Sprintf("user:%d:feed_version", fID.FollowerID)
		//     as.RedisRepository.Incr(context.Background(), versionKey)
		// }
		// 1. Fetch followers from SQL
		followers, err := as.PostRelationRepository.FetchFollowersUserIds(createPostReq.UserID)
		if err != nil || len(followers) == 0 {
			if err != nil {
				as.Log.Error("failed to fetch follower user ids from auth service",
					logger.Field{Key: "error", Value: err})
			}
			return
		}
		if len(followers) > 5 {
			key := fmt.Sprintf("celeb:posts:%d", createPostReq.UserID)

			pipe := as.RedisRepository.Pipeline()
			//1. Add the new post ID
			pipe.ZAdd(context.Background(), key, redis.Z{Score: float64(createPostRes.PostID), Member: createPostRes.PostID})

			//2. Keep only the latest 50 posts (remove everything from index 0 to -51)
			pipe.ZRemRangeByRank(context.Background(), key, 0, -51)

			//3. Set a long TTL (e.g., 7 days) because this is the primary cache for their feed
			pipe.Expire(context.Background(), key, 7*24*time.Hour)

			_, err = pipe.Exec(context.TODO())
			if err != nil {
				as.Log.Warn("redis pipeline failed(adding post id to celeb cache and removing old posts and setting expiry upon create post)",
					logger.Field{Key: "error", Value: err})
			}
			return
		}

		ctx := context.Background()
		pipe := as.RedisRepository.Pipeline()

		for i, fID := range followers {
			versionKey := fmt.Sprintf("user:%d:feed_version", fID.FollowerID)

			// This just queues the command locally in memory
			pipe.Incr(ctx, versionKey)
			pipe.Expire(ctx, versionKey, 48*time.Hour) // Keep the 48h TTL alive

			// 2. Batch Execution: Every 500 followers, send the batch to Redis
			if (i+1)%500 == 0 {
				_, err := pipe.Exec(ctx)
				if err != nil {
					//log.Printf("Pipeline execution error: %v", err)
					as.Log.Error("Pipeline execution error(invalidation all followers of cache for normal user upon create post)",
						logger.Field{Key: "error", Value: err})
				}
				// Create a fresh pipeline for the next batch
				pipe = as.RedisRepository.Pipeline()
			}
		}

		// 3. Final Execution for any remaining followers in the last batch
		if pipe.Len() > 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				as.Log.Error("Final Pipeline execution error(invalidation all followers of cache for normal user upon create post)",
					logger.Field{Key: "error", Value: err})
			}
		}
	}()

	return responsemodels.CreatePostResponse{
		PostID: createPostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) EditPost(editPostReq requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error) {

	editPostRes, err := as.PostRelationRepository.EditPostById(editPostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.EditPostResponse{}, domain.ErrPostNotFound
		}
		return responsemodels.EditPostResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", editPostReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon edit post",
			logger.Field{Key: "error", Value: err})
	}

	return responsemodels.EditPostResponse{
		Caption: editPostRes.Caption,
	}, nil
}

func (as *PostRelationUsecase) DeletePost(deletePostReq requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error) {
	deletePostRes, err := as.PostRelationRepository.DeletePostById(deletePostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeletePostResponse{}, domain.ErrPostNotFound
		}
		return responsemodels.DeletePostResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", deletePostReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon delete post",
			logger.Field{Key: "error", Value: err})
	}

	// 3. Handle Celebrity Cache Removal
	// We run this in a background goroutine to keep the response time fast
	go func() {
		ctx := context.Background()
		followers, err := as.PostRelationRepository.FetchFollowersUserIds(deletePostReq.UserID)
		if err != nil || len(followers) == 0 {
			if err != nil {
				as.Log.Error("failed to fetch follower user ids from auth service",
					logger.Field{Key: "error", Value: err})
			}
			return
		}
		num, err := strconv.Atoi(as.Cfg.CelebrityFollowCount) // returns (int, error)
		if err != nil {
			return
		}
		if len(followers) > num {
			key := fmt.Sprintf("celeb:posts:%d", deletePostReq.UserID)
			// ZRem removes the specific PostID from the ZSet
			pipe := as.RedisRepository.Pipeline()
			pipe.ZRem(ctx, key, deletePostRes.PostID)
			_, err := pipe.Exec(ctx)
			if err != nil {
				//log.Printf("Failed to remove post from celeb cache: %v", err)
				as.Log.Error("Failed to remove post from celeb cache:",
					logger.Field{Key: "error", Value: err})
			}
			return
		}
		// Check if the user is a celebrity
		// Option A: Use the Celebrity Table we discussed
		//isCeleb := as.PostRelationRepository.IsUserCelebrity(deletePostReq.UserID)

		// Option B: Fallback check by follower count if table doesn't exist yet
		//if !isCeleb {
		//  followers, _ := as.PostRelationRepository.FetchFollowersUserIds(deletePostReq.UserID)
		//if len(followers) > 50000 {
		//  isCeleb = true
		//}
		//}

		//if isCeleb {
		//  key := fmt.Sprintf("celeb:posts:%d", deletePostReq.UserID)
		// ZRem removes the specific PostID from the ZSet
		//err := as.RedisRepository.ZRem(ctx, key, deletePostRes.PostID)
		//if err != nil {
		//  log.Printf("Failed to remove post %d from celeb cache: %v", deletePostRes.PostID, err)
		//}
		//} else {
		// For non-celebrities, we should ideally increment the feed_version
		// of all followers so they don't see the deleted post in their cached feed.
		// followers, _ := as.PostRelationRepository.FetchFollowersUserIds(deletePostReq.UserID)
		pipe := as.RedisRepository.Pipeline()
		for _, fID := range followers {
			fVersionKey := fmt.Sprintf("user:%d:feed_version", fID.FollowerID)
			pipe.Incr(ctx, fVersionKey)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			as.Log.Error("pipeline execution error, failed to increment version key of normal users upon delete post",
				logger.Field{Key: "error", Value: err})
		}
	}()

	return responsemodels.DeletePostResponse{
		PostID: deletePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) LikePost(likePostReq requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error) {
	postOwnerId, err := as.PostRelationRepository.FetchPostOwnerIdByPostId(likePostReq.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.LikePostResponse{}, domain.ErrPostNotFound
		}
		log.Println(err)
		return responsemodels.LikePostResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	//fmt.Println("post Owner",postOwnerId)

	likePostRes, err := as.PostRelationRepository.LikePostById(likePostReq)
	if err != nil {
		return responsemodels.LikePostResponse{}, err
	}
	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", likePostReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon like post",
			logger.Field{Key: "error", Value: err})
	}

	event := map[string]interface{}{
		"type":        "POST_LIKE",
		"actorId":     likePostReq.UserID, // Person who clicked 'like'
		"postOwnerId": postOwnerId,        // Person receiving the notification
		"postId":      likePostRes.PostID, // The content being liked
		"timestamp":   time.Now().Unix(),
	}

	// Convert to JSON and publish to topic "post-events"
	err = as.KafkaProducer.PublishEvent("post-events", event)
	if err != nil {
		// Log the error but don't necessarily fail the request
		// unless real-time notification is critical.
		log.Printf("Failed to emit Kafka event: %v", err)
	}

	return responsemodels.LikePostResponse{
		PostID: likePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) UnlikePost(unlikePostReq requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error) {
	_, err := as.PostRelationRepository.FetchPostOwnerIdByPostId(unlikePostReq.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UnlikePostResponse{}, domain.ErrPostNotFound
		}
		log.Println(err)
		return responsemodels.UnlikePostResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	unlikePostRes, err := as.PostRelationRepository.UnlikePostById(unlikePostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UnlikePostResponse{}, domain.ErrPostLikeNotFound
		}
		return responsemodels.UnlikePostResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", unlikePostReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon unlike post",
			logger.Field{Key: "error", Value: err})
	}

	return responsemodels.UnlikePostResponse{
		PostID: unlikePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) AddComment(addCommentReq requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error) {
	if addCommentReq.ParentCommentId != nil {
		isReplytoReply, err := as.PostRelationRepository.CheckCommentHieracrchy(addCommentReq.ParentCommentId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return responsemodels.AddCommentResponse{}, domain.ErrCommentIdNotFound
			}
			return responsemodels.AddCommentResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
		}

		if isReplytoReply {
			return responsemodels.AddCommentResponse{}, domain.ErrRecursiveComment
		}
	}
	addCommentRes, err := as.PostRelationRepository.AddComment(addCommentReq)
	if err != nil {
		return responsemodels.AddCommentResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", addCommentReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon add comment",
			logger.Field{Key: "error", Value: err})
	}

	postOwnerId, err := as.PostRelationRepository.FetchPostOwnerIdByPostId(addCommentReq.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.AddCommentResponse{}, domain.ErrPostNotFound
		}
		return responsemodels.AddCommentResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	event := map[string]interface{}{
		"type":        "POST_COMMENT",
		"actorId":     addCommentReq.UserID, // Person who clicked 'like'
		"postOwnerId": postOwnerId,          // Person receiving the notification
		"postId":      addCommentReq.PostID, // The content being liked
		"timestamp":   time.Now().Unix(),
	}

	// Convert to JSON and publish to topic "post-events"
	err = as.KafkaProducer.PublishEvent("post-events", event)
	if err != nil {
		// Log the error but don't necessarily fail the request
		// unless real-time notification is critical.
		log.Printf("Failed to emit Kafka event: %v", err)
	}

	return addCommentRes, nil
}
func (as *PostRelationUsecase) EditComment(editCommentReq requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error) {
	postOwnerId, err := as.PostRelationRepository.FetchPostOwnerIdByPostId(editCommentReq.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.EditCommentResponse{}, domain.ErrPostIdNotFound
		}
		return responsemodels.EditCommentResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	if postOwnerId != editCommentReq.UserID {
		return responsemodels.EditCommentResponse{}, domain.ErrCommentEditDenied
	}
	resp, err := as.PostRelationRepository.EditComment(editCommentReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.EditCommentResponse{}, domain.ErrCommentNotFound
		}
		return responsemodels.EditCommentResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", editCommentReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon delete comment",
			logger.Field{Key: "error", Value: err})
	}

	return resp, nil
}
func (as *PostRelationUsecase) DeleteComment(deleteCommentReq requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error) {
	postOwnerId, err := as.PostRelationRepository.FetchPostOwnerIdByPostId(deleteCommentReq.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeleteCommentResponse{}, domain.ErrPostIdNotFound
		}
		log.Println("some database error occured while fetching post owner id by post id")
		return responsemodels.DeleteCommentResponse{}, err
	}
	if postOwnerId != deleteCommentReq.UserID {
		return responsemodels.DeleteCommentResponse{}, domain.ErrCommentDeleteDenied
	}
	deleteCommentRes, err := as.PostRelationRepository.DeleteCommentById(deleteCommentReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeleteCommentResponse{}, domain.ErrCommentNotFound
		}
		return responsemodels.DeleteCommentResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", deleteCommentReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon delete comment",
			logger.Field{Key: "error", Value: err})
	}

	return responsemodels.DeleteCommentResponse{
		CommentID: deleteCommentRes.CommentID,
	}, nil
}
func (as *PostRelationUsecase) Follow(followReq requestmodels.FollowRequest) (responsemodels.FollowResponse, error) {
	if followReq.UserID == followReq.FollowingUserID {
		return responsemodels.FollowResponse{}, domain.ErrFollowOwn
	}
	err := client.CheckUserExists(
		as.AuthSubscriptionClient,
		followReq.FollowingUserID,
	)
	if err != nil {
		return responsemodels.FollowResponse{}, err
	}
	// exists, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
	// 	UserId: followReq.FollowingUserID,
	// })
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.FollowResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.FollowResponse{}, domain.ErrUserNotFound

	// 	case codes.Internal:
	// 		return responsemodels.FollowResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.FollowResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	// if !exists.Exists {
	// 	return responsemodels.FollowResponse{}, domain.ErrUserNotFound
	// }
	followRes, err := as.PostRelationRepository.Follow(followReq)
	if err != nil {

		if err == gorm.ErrRecordNotFound {
			//fmt.Println("is it actually")
			return responsemodels.FollowResponse{}, domain.ErrAlreadyFollowing
		}
		return responsemodels.FollowResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", followReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon user follow",
			logger.Field{Key: "error", Value: err})
	}

	go func() {
		// ans, err := as.PostRelationRepository.FetchFollowCountByUserId(followReq.FollowingUserID)
		// if err != nil {
		// 	log.Println("error in executing goroutine for fetching follow count")
		// 	return
		// }
		fmt.Println("hi hereeeeeeeeee000000000")
		follower_count, err := as.PostRelationRepository.UpdatFollowCountOnFollow(followReq.UserID, followReq.FollowingUserID)
		if err != nil {
			fmt.Println("errrrrrr", err)
			log.Println("error in executing goroutine for fetching follow count")
			return
		}
		num, err := strconv.Atoi(as.Cfg.CelebrityFollowCount) // returns (int, error)
		if err != nil {
			return
		}
		if follower_count > uint64(num) {
			err := as.PostRelationRepository.PromoteToCelebrity(followReq.FollowingUserID)
			if err != nil {
				log.Println("database error in promoting to celebrrity", err)
				return
			}
		}
		// if ans.FollowerCount > 5 {
		// 	err := as.PostRelationRepository.PromoteToCelebrity(followReq.FollowingUserID)
		// 	if err != nil {
		// 		log.Println("database error in promoting to celebrrity", err)
		// 	}
		// }
	}()

	event := map[string]interface{}{
		"type":        "USER_FOLLOW",
		"actorId":     followReq.UserID,          // Person who clicked 'like'
		"followingId": followReq.FollowingUserID, // Person receiving the notification
		//"postId":        addCommentReq.PostID,     // The content being liked
		"timestamp": time.Now().Unix(),
	}

	// Convert to JSON and publish to topic "post-events"
	err = as.KafkaProducer.PublishEvent("user-events", event)
	if err != nil {
		// Log the error but don't necessarily fail the request
		// unless real-time notification is critical.
		log.Printf("Failed to emit Kafka event: %v", err)
	}

	return responsemodels.FollowResponse{
		FollowingUserID: followRes.FollowingUserID,
	}, nil
}

func (as *PostRelationUsecase) Unfollow(unfollowReq requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error) {
	if unfollowReq.UserID == unfollowReq.UnfollowingUserID {
		return responsemodels.UnfollowResponse{}, domain.ErrUnfollowOwn
	}
	err := client.CheckUserExists(
		as.AuthSubscriptionClient,
		unfollowReq.UnfollowingUserID,
	)
	if err != nil {
		return responsemodels.UnfollowResponse{}, err
	}
	// exists, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
	// 	UserId: unfollowReq.UnfollowingUserID,
	// })
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.UnfollowResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.UnfollowResponse{}, domain.ErrUserNotFound

	// 	case codes.Internal:
	// 		return responsemodels.UnfollowResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.UnfollowResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	// if !exists.Exists {
	// 	return responsemodels.UnfollowResponse{}, domain.ErrUserNotFound
	// }
	unfollowRes, err := as.PostRelationRepository.UnfollowUserById(unfollowReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//fmt.Println("is it actually")
			return responsemodels.UnfollowResponse{}, domain.ErrNoFollower
		}
		return responsemodels.UnfollowResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", unfollowReq.UserID)
	_, err = as.RedisRepository.Incr(context.Background(), versionKey)
	if err != nil {
		as.Log.Error("Version key increment failed in redis upon unfollow user",
			logger.Field{Key: "error", Value: err})
	}

	go func() {
		// ans, err := as.PostRelationRepository.FetchFollowCountByUserId(unfollowReq.UnfollowingUserID)
		// if err != nil {
		// 	log.Println("error in executing goroutine for fetching follow count")
		// 	return
		// }
		follower_count, err := as.PostRelationRepository.UpdatFollowCountOnUnFollow(unfollowReq.UserID, unfollowReq.UnfollowingUserID)
		if err != nil {
			log.Println("error in executing goroutine for fetching follow count")
			return
		}
		if follower_count <= 5 {
			err := as.PostRelationRepository.DepromoteToNormalUser(unfollowReq.UnfollowingUserID)
			if err != nil {
				log.Println("database error in depromoting to normal user", err)
				return
			}
		}
		// if ans.FollowerCount <= 5 {
		// 	err := as.PostRelationRepository.DepromoteToNormalUser(unfollowReq.UnfollowingUserID)
		// 	if err != nil {
		// 		log.Println("database error in depromoting to normal user", err)
		// 	}
		// }
	}()

	return responsemodels.UnfollowResponse{
		UnfollowingUserID: unfollowRes.UnfollowingUserID,
	}, nil
}

func (as *PostRelationUsecase) FetchComments(fetchCommentsReq requestmodels.FetchCommentsReqeust) (responsemodels.FetchCommentsResponse, error) {
	commentsRes, err := as.PostRelationRepository.FetchCommentsByPostId(fetchCommentsReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FetchCommentsResponse{}, domain.ErrPostIdNotFound
		}
		return responsemodels.FetchCommentsResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	userIDs := make(map[uint64]bool)
	for _, v := range commentsRes {
		userIDs[v.UserID] = true
	}
	userids := make([]uint64, len(userIDs))
	i := 0
	for k := range userIDs {
		userids[i] = k
		i++
	}
	// userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
	// 	UserId: userids,
	// })
	userResp, err := client.FetchUserMetaData(
		as.AuthSubscriptionClient,
		userids,
	)
	if err != nil {
		return responsemodels.FetchCommentsResponse{}, err
	}
	//v:=userResp[userIDs]
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.FetchCommentsResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.FetchCommentsResponse{}, domain.ErrUsersNotFound

	// 	case codes.Internal:
	// 		return responsemodels.FetchCommentsResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.FetchCommentsResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	var comments []responsemodels.Comment
	for i, v := range commentsRes {
		if commentsRes[i].ParentCommentID == nil {
			comments = append(comments, responsemodels.Comment{
				CommentID:   v.ID,
				CommentText: v.CommentText,
				CreatedAt:   v.CreatedAt,
				CommentAge:  utils.CalcuateCommentAge(v.CreatedAt),
				UserDetails: responsemodels.UserMetaData{
					UserID:        userResp.Users[v.UserID].UserId,
					UserName:      userResp.Users[v.UserID].UserName,
					Name:          userResp.Users[v.UserID].Name,
					ProfileImgUrl: userResp.Users[v.UserID].ProfileImgUrl,
					BlueTick:      userResp.Users[v.UserID].BlueTick,
				},
				ParentCommentID: v.ParentCommentID,
			})
		}
	}
	// create index lookup for parents
	parentIndex := make(map[uint64]int)
	for i, c := range comments {
		parentIndex[c.CommentID] = i
	}
	for i, v := range commentsRes {
		if commentsRes[i].ParentCommentID != nil {
			parentIdx, ok := parentIndex[*v.ParentCommentID]
			if !ok {
				return responsemodels.FetchCommentsResponse{}, domain.ErrInvalidParentCommentId
			}

			comments[parentIdx].ChildComment = append(comments[parentIdx].ChildComment, responsemodels.Comment{
				CommentID:   v.ID,
				CommentText: v.CommentText,
				CreatedAt:   v.CreatedAt,
				CommentAge:  utils.CalcuateCommentAge(v.CreatedAt),
				UserDetails: responsemodels.UserMetaData{
					UserID:        userResp.Users[v.UserID].UserId,
					UserName:      userResp.Users[v.UserID].UserName,
					Name:          userResp.Users[v.UserID].Name,
					ProfileImgUrl: userResp.Users[v.UserID].ProfileImgUrl,
					BlueTick:      userResp.Users[v.UserID].BlueTick,
				},
				ParentCommentID: v.ParentCommentID,
			})
		}
	}
	return responsemodels.FetchCommentsResponse{
		Comments: comments,
	}, nil
}

func (as *PostRelationUsecase) PostFollowCount(userid uint64) (responsemodels.PostFollowCountResponse, error) {
	postCount, err := as.PostRelationRepository.FetchPostCountByUserId(userid)
	if err != nil {
		return responsemodels.PostFollowCountResponse{}, err
	}
	//fmt.Println("print post Count in usecase", postCount)
	resp, err := as.PostRelationRepository.FetchFollowCountByUserId(userid)
	if err != nil {
		return responsemodels.PostFollowCountResponse{}, err
	}
	//fmt.Println("resp print first in usecase", resp)
	resp.PostCount = postCount
	//fmt.Println("resp print second in usecase", resp, resp.PostCount)
	return resp, nil
}
func (as *PostRelationUsecase) FetchAllPosts(req requestmodels.FetchAllPostsReq) ([]responsemodels.PostWithCounts, error) {
	// err := client.CheckUserExists(
	// 	as.AuthSubscriptionClient,
	// 	req.TargetUserID,
	// )
	// if err != nil {
	// 	return nil, err
	// }
	// exists, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
	// 	UserId: req.TargetUserID,
	// })
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return nil,
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return nil, domain.ErrUserNotFound

	// 	case codes.Internal:
	// 		return nil,
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return nil,
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	// if !exists.Exists {
	// 	return nil, domain.ErrUserNotFound
	// }
	resp, err := as.PostRelationRepository.FetchAllPosts(req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNoPosts
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	for i := range resp {
		resp[i].Age = utils.CalcuateCommentAge(resp[i].CreatedAt)
	}
	return resp, nil
}
func (as *PostRelationUsecase) FetchFollowers(req requestmodels.FetchFollowersRequest) (responsemodels.FetchFollowersResponse, error) {
	// exists, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
	// 	UserId: req.UserID,
	// })
	err := client.CheckUserExists(
		as.AuthSubscriptionClient,
		req.UserID,
	)
	if err != nil {
		return responsemodels.FetchFollowersResponse{}, err
	}
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.FetchFollowersResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.FetchFollowersResponse{}, domain.ErrUserNotFound

	// 	case codes.Internal:
	// 		return responsemodels.FetchFollowersResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.FetchFollowersResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	// if !exists.Exists {
	// 	return responsemodels.FetchFollowersResponse{}, domain.ErrUserNotFound
	// }
	resp, err := as.PostRelationRepository.FetchFollowersUserIds1(req)
	if err != nil {
		// if err == gorm.ErrRecordNotFound {
		// 	return responsemodels.FetchFollowersResponse{}, domain.ErrNoFollowers
		// }
		log.Println(err)
		return responsemodels.FetchFollowersResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	var userids []uint64
	for _, v := range resp {
		userids = append(userids, v.FollowerID)
	}

	if len(userids) == 0 {
		return responsemodels.FetchFollowersResponse{
			Followers: []responsemodels.UserMetaData{}, // Returns an empty slice, not nil
		}, nil
	}
	// userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
	// 	UserId: userids,
	// })
	userResp, err := client.FetchUserMetaData(
		as.AuthSubscriptionClient,
		userids,
	)
	if err != nil {
		return responsemodels.FetchFollowersResponse{}, err
	}
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.FetchFollowersResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.FetchFollowersResponse{}, domain.ErrUsersNotFound

	// 	case codes.Internal:
	// 		return responsemodels.FetchFollowersResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.FetchFollowersResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	var usermetada []responsemodels.UserMetaData
	for _, v := range userResp.Users {
		usermetada = append(usermetada, responsemodels.UserMetaData{
			UserID:        v.UserId,
			UserName:      v.UserName,
			Name:          v.Name,
			ProfileImgUrl: v.ProfileImgUrl,
			BlueTick:      v.BlueTick,
		})
	}
	//v:=userResp[userIDs]

	return responsemodels.FetchFollowersResponse{
		Followers: usermetada,
	}, nil
}
func (as *PostRelationUsecase) FetchFollowing(req requestmodels.FetchFollowingRequest) (responsemodels.FetchFollowingResponse, error) {
	err := client.CheckUserExists(
		as.AuthSubscriptionClient,
		req.UserID,
	)
	if err != nil {
		return responsemodels.FetchFollowingResponse{}, err
	}
	// exists, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
	// 	UserId: req.UserID,
	// })
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.FetchFollowingResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.FetchFollowingResponse{}, domain.ErrUserNotFound

	// 	case codes.Internal:
	// 		return responsemodels.FetchFollowingResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.FetchFollowingResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	// if !exists.Exists {
	// 	return responsemodels.FetchFollowingResponse{}, domain.ErrUserNotFound
	// }
	resp, err := as.PostRelationRepository.FetchFollowingUserIds(req)
	if err != nil {
		// if err == gorm.ErrRecordNotFound {
		// 	return responsemodels.FetchFollowingResponse{}, domain.ErrNoFollowing
		// }
		log.Println(err)
		return responsemodels.FetchFollowingResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
	}
	var userids []uint64
	for _, v := range resp {
		userids = append(userids, v.FollowingID)
	}
	if len(userids) == 0 {
		return responsemodels.FetchFollowingResponse{
			Following: []responsemodels.UserMetaData{}, // Returns an empty slice, not nil
		}, nil
	}
	userResp, err := client.FetchUserMetaData(
		as.AuthSubscriptionClient,
		userids,
	)
	if err != nil {
		return responsemodels.FetchFollowingResponse{}, err
	}
	// userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
	// 	UserId: userids,
	// })
	// if err != nil {
	// 	st, ok := status.FromError(err)
	// 	if !ok {
	// 		// Not a gRPC error
	// 		return responsemodels.FetchFollowingResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}

	// 	switch st.Code() {

	// 	case codes.NotFound:
	// 		return responsemodels.FetchFollowingResponse{}, domain.ErrUsersNotFound

	// 	case codes.Internal:
	// 		return responsemodels.FetchFollowingResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)

	// 	default:
	// 		return responsemodels.FetchFollowingResponse{},
	// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
	// 	}
	// }
	var usermetada []responsemodels.UserMetaData
	for _, v := range userResp.Users {
		usermetada = append(usermetada, responsemodels.UserMetaData{
			UserID:        v.UserId,
			UserName:      v.UserName,
			Name:          v.Name,
			ProfileImgUrl: v.ProfileImgUrl,
			BlueTick:      v.BlueTick,
		})
	}
	return responsemodels.FetchFollowingResponse{
		Following: usermetada,
	}, nil
}

// func (as *PostRelationUsecase) FetchPostUserDataForNewsFeed(newsfeedReq requestmodels.FetchNewsFeedRequest) (responsemodels.FetchNewsFeedResponse, error) {
// 	ctx := context.Background()
// 	version := as.getFeedVersion(ctx, newsfeedReq.UserID)

// 	// Key now uses LastID instead of Offset
// 	cacheKey := fmt.Sprintf("newsfeed:%d:v:%s:lim:%d:last:%d",
// 		newsfeedReq.UserID, version, newsfeedReq.Limit, newsfeedReq.LastID)

// 	if newsfeedReq.PullToRefresh {
// 		// Increment version to effectively "clear" all pages at once
// 		versionKey := fmt.Sprintf("user:%d:feed_version", newsfeedReq.UserID)
// 		version, _ = as.RedisRepository.Incr(context.Background(), versionKey) // Need to add Incr to your interface
// 		//if err != nil {
// 		//  version = "1" // Fallback
// 		newsfeedReq.LastID = 0
// 		cacheKey = fmt.Sprintf("newsfeed:%d:v:%s:lim:%d:last:0", newsfeedReq.UserID, version, newsfeedReq.Limit)
// 	} else {
// 		cachedData, err := as.RedisRepository.CacheGet(ctx, cacheKey)
// 		if err == nil {
// 			var cachedResp responsemodels.FetchNewsFeedResponse
// 			if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
// 				fmt.Println("getting cached response")
// 				return cachedResp, nil
// 			}
// 		}
// 	}

// 	// 2. CACHE MISS: Execute your existing logic
// 	postResp, err := as.PostRelationRepository.FetchPostDataForNewsFeed(newsfeedReq)
// 	if err != nil {
// 		return responsemodels.FetchNewsFeedResponse{}, err
// 	}
// 	if len(postResp) == 0 {
// 		if newsfeedReq.LastID == 0 {
// 			return responsemodels.FetchNewsFeedResponse{}, domain.ErrNoFollowingNoPost
// 		}
// 		// Return empty response with HasMore false
// 		return responsemodels.FetchNewsFeedResponse{HasMore: false}, nil
// 	}
// 	userIDs := make(map[uint64]bool)

// 	for _, v := range postResp {

// 		userIDs[uint64(v.UserID)] = true
// 	}

// 	userids := make([]uint64, len(userIDs))
// 	i := 0
// 	for k := range userIDs {
// 		userids[i] = k
// 		i++
// 	}

// 	userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
// 		UserId: userids,
// 	})
// 	if err != nil {
// 		log.Println("error calling service auth_subcription", err)
// 		return responsemodels.FetchNewsFeedResponse{}, err
// 	}
// 	for i, v := range postResp {
// 		uid := uint64(v.UserID)

// 		// SAFE MAPPING: check if user exists in the map
// 		if userData, ok := userResp.Users[uid]; ok {
// 			postResp[i].UserDetails = responsemodels.UserMetaData{
// 				UserID:        userData.UserId,
// 				UserName:      userData.UserName,
// 				Name:          userData.Name,
// 				ProfileImgUrl: userData.ProfileImgUrl,
// 				BlueTick:      userData.BlueTick,
// 			}
// 		} else {
// 			log.Printf("Warning: Metadata for user %d not found in auth service", uid)
// 		}

// 		postResp[i].Age = utils.CalcuateCommentAge(v.CreatedAt)
// 	}

// 	var nextCursor uint64
// 	hasMore := false
// 	if len(postResp) > int(newsfeedReq.Limit) {
// 		hasMore = true
// 		// Remove the extra item so the user only gets the 10 they asked for
// 		postResp = postResp[:newsfeedReq.Limit]
// 	}
// 	if len(postResp) > 0 {
// 		// The ID of the last item in our result is the cursor for the next request
// 		nextCursor = uint64(postResp[len(postResp)-1].ID)
// 	}

//		finalResponse := responsemodels.FetchNewsFeedResponse{
//			PostUserData: postResp,
//			NextCursor:   nextCursor,
//			HasMore:      hasMore,
//		}
//		// 3. Store in Redis for future requests (e.g., 5 minutes TTL)
//		dataToCache, err := json.Marshal(finalResponse)
//		if err != nil {
//			return responsemodels.FetchNewsFeedResponse{}, err
//		}
//		err = as.RedisRepository.CacheSet(context.Background(), cacheKey, dataToCache, 5*time.Minute)
//		if err != nil {
//			log.Printf("Failed to cache newsfeed for key %s: %v", cacheKey, err)
//			// Note: Usually we don't return an error here because we still have the data
//			// to return to the user; the cache failing shouldn't break the whole app.
//		}
//		log.Println("returning sql response")
//		return finalResponse, nil
//	}
func (as *PostRelationUsecase) FetchPostUserDataForNewsFeed(newsfeedReq requestmodels.FetchNewsFeedRequest) (responsemodels.FetchNewsFeedResponse, error) {
	as.Log.Debug("starting")
	ctx := context.Background()

	// 1. TRY CACHE (ONLY for Normal Posts)
	version, err := as.getFeedVersion(ctx, newsfeedReq.UserID)
	if err != nil {
		as.Log.Error("redis error in getting feed version",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: newsfeedReq.UserID})
	}
	normalCacheKey := fmt.Sprintf("newsfeed:normal:%d:v:%s:lim:%d:last:%d", newsfeedReq.UserID, version, newsfeedReq.Limit, newsfeedReq.LastID)

	var normalPosts []responsemodels.PostWithStatus
	cacheHit := false

	if !newsfeedReq.PullToRefresh {
		cachedData, err := as.RedisRepository.CacheGet(ctx, normalCacheKey)
		if err == nil {
			json.Unmarshal([]byte(cachedData), &normalPosts)
			cacheHit = true
			as.Log.Info("returnnig cached response of normal users")
		} else {
			as.Log.Error("Cache miss or redis error upon fetch in noremal usesrs newsfeed",
				logger.Field{Key: "error", Value: err})
		}
	}
	//3. LIVE INJECTION: Get Celebrity Posts
	celebIDs, err := as.PostRelationRepository.GetFollowedCelebrityIDs(newsfeedReq.UserID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return responsemodels.FetchNewsFeedResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
		}
	}
	//fmt.Println("print celeb ids ", celebIDs)
	var celebPosts []responsemodels.PostWithStatus

	//var userResp *pb.BatchUserMetadataResponse

	// 2. IF CACHE MISS: Get Normal Posts from DB
	if !cacheHit {
		normalPosts, err = as.PostRelationRepository.FetchNormalPostData(newsfeedReq)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return responsemodels.FetchNewsFeedResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
			}

		}
		if len(normalPosts) > 0 {
			//making a map to avoid duplicate userids
			userIDs := make(map[uint64]bool)

			for _, v := range normalPosts {

				userIDs[uint64(v.UserID)] = true
			}

			userids := make([]uint64, len(userIDs))
			i := 0
			for k := range userIDs {
				userids[i] = k
				i++
			}
			//userids = append(userids, celebIDs...)
			// userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
			// 	UserId: userids,
			// })
			userResp, err := client.FetchUserMetaData(
				as.AuthSubscriptionClient,
				userids,
			)
			if err != nil {
				return responsemodels.FetchNewsFeedResponse{}, err
			}
			// if err != nil {
			// 	st, ok := status.FromError(err)
			// 	if !ok {
			// 		// Not a gRPC error
			// 		return responsemodels.FetchNewsFeedResponse{},
			// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
			// 	}
			// 	switch st.Code() {
			// 	case codes.NotFound:
			// 		return responsemodels.FetchNewsFeedResponse{}, domain.ErrUsersNotFound
			// 	case codes.Internal:
			// 		return responsemodels.FetchNewsFeedResponse{},
			// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)
			// 	default:
			// 		return responsemodels.FetchNewsFeedResponse{},
			// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
			// 	}
			// }

			for i, v := range normalPosts {
				uid := uint64(v.UserID)
				// SAFE MAPPING: check if user exists in the map
				if userData, ok := userResp.Users[uid]; ok {
					//fmt.Println("what about here",ok,userData,userResp.Users[uid])
					normalPosts[i].UserDetails = responsemodels.UserMetaData{
						UserID:        userData.UserId,
						UserName:      userData.UserName,
						Name:          userData.Name,
						ProfileImgUrl: userData.ProfileImgUrl,
						BlueTick:      userData.BlueTick,
					}
				} else {
					return responsemodels.FetchNewsFeedResponse{}, fmt.Errorf("%v:", "Metadata for normal users not found in auth service")
				}

				normalPosts[i].Age = utils.CalcuateCommentAge(v.CreatedAt)
			}
			// Store in cache for 5 minutes
			data, err := json.Marshal(normalPosts)
			if err != nil {
				as.Log.Error("error in marshalling json of normal posts")
			}
			err = as.RedisRepository.CacheSet(ctx, normalCacheKey, data, 5*time.Minute)
			if err != nil {
				as.Log.Error("failed to set cache of normal posts",
					logger.Field{Key: "error", Value: err})
			}
		}
	}
	if len(celebIDs) > 0 {
		// A. Try pulling IDs from Redis ZSets
		celebPostIDs, err := as.RedisRepository.PullCelebPostIDsFromRedis(ctx, celebIDs, newsfeedReq.LastID, int(newsfeedReq.Limit))
		// B. CACHE MISS FALLBACK: If Redis is empty, fetch IDs from SQL
		if err != nil || len(celebPostIDs) == 0 {
			as.Log.Info("Celeb cache miss for user falling back to SQL",
				logger.Field{Key: "userid", Value: newsfeedReq.UserID})
			celebPostIDs, err = as.PostRelationRepository.FetchCelebrityPostIDsFromSQL(celebIDs, newsfeedReq.LastID, int(newsfeedReq.Limit))
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					return responsemodels.FetchNewsFeedResponse{}, fmt.Errorf("%w: %v", domain.ErrDatabase, err)
				}
			}
			// OPTIONAL: You could trigger a background job here to re-populate
			// the Redis ZSET so the next request is faster.
			// Repopulate in the background
			if len(celebPostIDs) > 0 {
				as.Log.Info("repopulating celeb cache")
				go as.RepopulateCelebrityCache(context.Background(), celebIDs)
			}
		}
		// C. Hydrate those IDs from SQL (Get full post details, counts, etc.)
		if len(celebPostIDs) > 0 {
			as.Log.Info("showning response from db, but fecthed post ids though cache")
			celebPosts, err = as.PostRelationRepository.FetchPostsByIDs(celebPostIDs, newsfeedReq.UserID)
			if err != nil {
				//log.Println("failed to fetch celeb posts by id")
				if err == gorm.ErrRecordNotFound {
					return responsemodels.FetchNewsFeedResponse{}, domain.CelebPostsNotFound
				} else {
					return responsemodels.FetchNewsFeedResponse{}, fmt.Errorf("%w:; %v", domain.ErrDatabase, err)
				}
			}
			userIDs := make(map[uint64]bool)

			for _, v := range celebPosts {

				userIDs[uint64(v.UserID)] = true
			}
			userids := make([]uint64, len(userIDs))
			i := 0
			for k := range userIDs {
				userids[i] = k
				i++
			}
			userids = append(userids, celebIDs...)

			// userResp2, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
			// 	UserId: userids,
			// })
			userResp2, err := client.FetchUserMetaData(
				as.AuthSubscriptionClient,
				userids,
			)
			if err != nil {
				return responsemodels.FetchNewsFeedResponse{}, err
			}

			// if err != nil {
			// 	st, ok := status.FromError(err)
			// 	if !ok {
			// 		// Not a gRPC error
			// 		return responsemodels.FetchNewsFeedResponse{},
			// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
			// 	}
			// 	switch st.Code() {
			// 	case codes.NotFound:
			// 		return responsemodels.FetchNewsFeedResponse{}, domain.ErrUsersNotFound
			// 	case codes.Internal:
			// 		return responsemodels.FetchNewsFeedResponse{},
			// 			fmt.Errorf("%w: %v", domain.ErrDatabase, err)
			// 	default:
			// 		return responsemodels.FetchNewsFeedResponse{},
			// 			fmt.Errorf("%w: %v", domain.ErrInternal, err)
			// 	}
			// }
			for i, v := range celebPosts {
				uid := uint64(v.UserID)
				// SAFE MAPPING: check if user exists in the map
				if userData, ok := userResp2.Users[uid]; ok {
					celebPosts[i].UserDetails = responsemodels.UserMetaData{
						UserID:        userData.UserId,
						UserName:      userData.UserName,
						Name:          userData.Name,
						ProfileImgUrl: userData.ProfileImgUrl,
						BlueTick:      userData.BlueTick,
					}
				} else if !ok {
					as.Log.Error("Metadata for celeb users not found in auth service")
					return responsemodels.FetchNewsFeedResponse{}, fmt.Errorf("Metadata for celeb users not found in auth service")
				}
				celebPosts[i].Age = utils.CalcuateCommentAge(v.CreatedAt)
			}
		}
	}
	// 4. MERGE & SORT
	allPosts := append(normalPosts, celebPosts...)
	sort.Slice(allPosts, func(i, j int) bool {
		return allPosts[i].ID > allPosts[j].ID
	})

	var nextCursor uint64
	hasMore := false
	if len(allPosts) > int(newsfeedReq.Limit) {
		hasMore = true
		// Remove the extra item so the user only gets the 10 they asked for
		allPosts = allPosts[:newsfeedReq.Limit]
	}
	if len(allPosts) > 0 {
		// The ID of the last item in our result is the cursor for the next request
		nextCursor = uint64(allPosts[len(allPosts)-1].ID)
	} else {
		//return responsemodels.FetchNewsFeedResponse{}, domain.ErrNoFollowingNoPost
	}

	finalResponse := responsemodels.FetchNewsFeedResponse{
		PostUserData: allPosts,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
	}
	return finalResponse, nil
}
func (as *PostRelationUsecase) RepopulateCelebrityCache(ctx context.Context, celebIDs []uint64) {
	for _, cID := range celebIDs {
		// 1. Fetch the latest 50 posts for this specific celebrity from SQL
		// We do this per celebrity to keep the ZSet clean
		posts, err := as.PostRelationRepository.FetchLatestPostIDsByUserID(cID, 50)
		if err != nil || len(posts) == 0 {
			if err != gorm.ErrRecordNotFound {
				as.Log.Error("Error in fetching latest post ids by user id",
					logger.Field{Key: "error", Value: err})
			}
			continue
		}

		key := fmt.Sprintf("celeb:posts:%d", cID)
		pipe := as.RedisRepository.Pipeline()

		for _, pID := range posts {
			// Score = PostID for chronological order
			// Correct way to use ZAdd in a Pipeline
			pipe.ZAdd(ctx, key, redis.Z{
				Score:  float64(pID), // The sorting criteria
				Member: pID,          // The actual value stored
			})
		}

		// 2. Keep only the last 50 and set TTL
		pipe.ZRemRangeByRank(ctx, key, 0, -51)
		pipe.Expire(ctx, key, 7*24*time.Hour)

		_, err = pipe.Exec(ctx)
		if err != nil {
			//log.Printf("Failed to repopulate cache for celeb %d: %v", cID, err)
			as.Log.Error("Failed to repopulate cache for celeb",
				logger.Field{Key: "error", Value: err})
		}
	}
}
func (as *PostRelationUsecase) getFeedVersion(ctx context.Context, userID uint64) (string, error) {
	versionKey := fmt.Sprintf("user:%d:feed_version", userID)
	version, err := as.RedisRepository.CacheGet(ctx, versionKey)
	if err != nil || len(version) == 0 {
		// If no version exists, start at 1
		if err == redis.Nil {
			as.Log.Info("setting feed version to 1",
				logger.Field{Key: "error", Value: err})
			as.RedisRepository.CacheSet(ctx, versionKey, []byte("1"), 48*time.Hour)
			return "1", nil
		}
		return "", fmt.Errorf("redis error while getting feed_version %v:", err)
	}
	// OPTIONAL: Refresh the 48h timer so active users never lose their version
	err = as.RedisRepository.ExtendTTL(ctx, versionKey, 48*time.Hour)
	if err != nil {
		return "", err
	}
	return version, nil
}

func (as *PostRelationUsecase) FetchGlobalNewsFeed(req requestmodels.GlobalNewsFeedRequest) (responsemodels.FetchGlobalNewsFeedResponse, error) {
	postResp, err := as.PostRelationRepository.FetchGlobalTrendingSQL(req)
	if err != nil {
		return responsemodels.FetchGlobalNewsFeedResponse{}, err
	}
	//fmt.Println("post resp",postResp)
	// if len(postResp) == 0 {
	// 	return responsemodels.FetchGlobalNewsFeedResponse{}, domain.ErrNoPostGlobally
	// }
	userIDs := make(map[uint64]bool)

	for _, v := range postResp {

		userIDs[uint64(v.UserID)] = true
	}

	userids := make([]uint64, len(userIDs))
	i := 0
	for k := range userIDs {
		userids[i] = k
		i++
	}

	// userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
	// 	UserId: userids,
	// })
	userResp, err := client.FetchUserMetaData(
		as.AuthSubscriptionClient,
		userids,
	)
	if err != nil {
		if err!=domain.ErrUsersNotFound{
			return responsemodels.FetchGlobalNewsFeedResponse{}, err
		}
		
	}
	// if err != nil {
	// 	log.Println("error calling service auth_subcription", err)
	// 	return responsemodels.FetchGlobalNewsFeedResponse{}, err
	// }
	for i, v := range postResp {
		uid := uint64(v.UserID)

		// SAFE MAPPING: check if user exists in the map
		if userData, ok := userResp.Users[uid]; ok {
			postResp[i].UserDetails = responsemodels.UserMetaData{
				UserID:        userData.UserId,
				UserName:      userData.UserName,
				Name:          userData.Name,
				ProfileImgUrl: userData.ProfileImgUrl,
				BlueTick:      userData.BlueTick,
			}
		} else {
			log.Printf("Warning: Metadata for user %d not found in auth service", uid)
		}

		postResp[i].Age = utils.CalcuateCommentAge(v.CreatedAt)
	}

	var nextCursor float64
	hasMore := false
	if len(postResp) > int(req.Limit) {
		hasMore = true
		// Remove the extra item so the user only gets the 10 they asked for
		postResp = postResp[:req.Limit]
	}
	if len(postResp) > 0 {
		// The ID of the last item in our result is the cursor for the next request
		nextCursor = postResp[len(postResp)-1].TrendingScore
	}

	finalResponse := responsemodels.FetchGlobalNewsFeedResponse{
		PostUserData: postResp,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
	}

	return finalResponse, nil
}
