package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/pb"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/utils"
	"gorm.io/gorm"
)

type PostRelationUsecase struct {
	PostRelationRepository interfacesRepository.PostRelationRepository
	AuthSubscriptionClient pb.AuthSubscriptionServiceClient
	RedisRepository        interfacesRepository.RedisRepository
}

var (
	ErrPostNotFound     = errors.New("Post Not found or user does not have permission")
	ErrPostLikeNotFound = errors.New("Post Like Not found or like does not belong to the user")
	ErrRecursiveComment = errors.New("can't reply to a comment reply")
	ErrCommentNotFound  = errors.New("comment doesn't exist or post doesn't exist or user does not have permission")
	ErrFollowOwn        = errors.New("can't follow yourself")
	ErrUsertNotFound    = errors.New("User not found")
	ErrUnfollowOwn      = errors.New("can't unfollow yourself")
	ErrNoComments       = errors.New("No comments to Fetch for the Post or Post doesn't exist")
	ErrNoPosts          = errors.New("No Posts to Fetch")
)

func NewPostRelationUsecase(repository interfacesRepository.PostRelationRepository, authSubClient pb.AuthSubscriptionServiceClient, redisRepository interfacesRepository.RedisRepository) interfacesUsecase.PostRelationUsecase {
	return &PostRelationUsecase{
		PostRelationRepository: repository,
		AuthSubscriptionClient: authSubClient,
		RedisRepository:        redisRepository,
	}
}

func (as *PostRelationUsecase) CreatePost(createPostReq requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error) {
	createPostRes, err := as.PostRelationRepository.CreatePost(createPostReq)
	if err != nil {
		return responsemodels.CreatePostResponse{}, nil
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", createPostReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

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
			return
		}
		if len(followers) > 50000 {
			//key := fmt.Sprintf("celeb:posts:%d", createPostReq.UserID)

			//pipe := as.RedisRepository.Pipeline()
			// 1. Add the new post ID
			//pipe.ZAdd(context.Background(), key, redis.Z{Score: float64(createPostRes.PostID), Member: createPostRes.PostID})

			// 2. Keep only the latest 50 posts (remove everything from index 0 to -51)
			//pipe.ZRemRangeByRank(context.Background(), key, 0, -51)

			// 3. Set a long TTL (e.g., 7 days) because this is the primary cache for their feed
			//pipe.Expire(context.Background(), key, 7*24*time.Hour)

			//_, _ = pipe.Exec(context.TODO())
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
					log.Printf("Pipeline execution error: %v", err)
				}
				// Create a fresh pipeline for the next batch
				pipe = as.RedisRepository.Pipeline()
			}
		}

		// 3. Final Execution for any remaining followers in the last batch
		if pipe.Len() > 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				log.Printf("Final pipeline execution error: %v", err)
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
			return responsemodels.EditPostResponse{}, ErrPostNotFound
		}
		return responsemodels.EditPostResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", editPostReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)


	return responsemodels.EditPostResponse{
		Caption: editPostRes.Caption,
	}, nil
}

func (as *PostRelationUsecase) DeletePost(deletePostReq requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error) {
	deletePostRes, err := as.PostRelationRepository.DeletePostById(deletePostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeletePostResponse{}, ErrPostNotFound
		}
		return responsemodels.DeletePostResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", deletePostReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)


	return responsemodels.DeletePostResponse{
		PostID: deletePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) LikePost(likePostReq requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error) {
	likePostRes, err := as.PostRelationRepository.LikePostById(likePostReq)
	if err != nil {
		return responsemodels.LikePostResponse{}, err
	}
	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", likePostReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

	return responsemodels.LikePostResponse{
		PostID: likePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) UnlikePost(unlikePostReq requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error) {
	unlikePostRes, err := as.PostRelationRepository.UnlikePostById(unlikePostReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UnlikePostResponse{}, ErrPostLikeNotFound
		}
		return responsemodels.UnlikePostResponse{}, err
	}
	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", unlikePostReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

	return responsemodels.UnlikePostResponse{
		PostID: unlikePostRes.PostID,
	}, nil
}

func (as *PostRelationUsecase) AddComment(addCommentReq requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error) {
	if addCommentReq.ParentCommentId != nil {
		fmt.Println("is reaching in here in add comment where parent comment Id not nil")
		isReplytoReply, err := as.PostRelationRepository.CheckCommentHieracrchy(addCommentReq.ParentCommentId)
		if err != nil {
			return responsemodels.AddCommentResponse{}, err
		}
		fmt.Println("print the truth :", isReplytoReply)
		if isReplytoReply {
			fmt.Println("it is true")
			return responsemodels.AddCommentResponse{}, ErrRecursiveComment
			//fmt.Println("here 1")
		}
		fmt.Println("here 2")
	}
	fmt.Println("here 3")
	addCommentRes, err := as.PostRelationRepository.AddComment(addCommentReq)
	if err != nil {
		return responsemodels.AddCommentResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", addCommentReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)


	return addCommentRes, nil
}
func (as *PostRelationUsecase) EditComment(editCommentReq requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error) {
	resp, err := as.PostRelationRepository.EditComment(editCommentReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.EditCommentResponse{}, ErrCommentNotFound
		}
		return responsemodels.EditCommentResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", editCommentReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

	return resp, nil
}
func (as *PostRelationUsecase) DeleteComment(deleteCommentReq requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error) {
	deleteCommentRes, err := as.PostRelationRepository.DeleteCommentById(deleteCommentReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.DeleteCommentResponse{}, ErrCommentNotFound
		}
		return responsemodels.DeleteCommentResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", deleteCommentReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

	return responsemodels.DeleteCommentResponse{
		CommentID: deleteCommentRes.CommentID,
	}, nil
}
func (as *PostRelationUsecase) Follow(followReq requestmodels.FollowRequest) (responsemodels.FollowResponse, error) {
	if followReq.UserID == followReq.FollowingUserID {
		return responsemodels.FollowResponse{}, ErrFollowOwn
	}
	_, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
		UserId: followReq.FollowingUserID,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FollowResponse{}, ErrUsertNotFound
		}
		log.Println("inter service call for check user exist failed, error: ", err)
		return responsemodels.FollowResponse{}, err
	}
	followRes, err := as.PostRelationRepository.Follow(followReq)
	if err != nil {

		if err == gorm.ErrRecordNotFound {
			fmt.Println("is it actually")
			return responsemodels.FollowResponse{}, domain.ErrAlreadyFollowing
		}
		return responsemodels.FollowResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", followReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

	return responsemodels.FollowResponse{
		FollowingUserID: followRes.FollowingUserID,
	}, nil
}

func (as *PostRelationUsecase) Unfollow(unfollowReq requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error) {
	if unfollowReq.UserID == unfollowReq.UnfollowingUserID {
		return responsemodels.UnfollowResponse{}, ErrUnfollowOwn
	}
	_, err := as.AuthSubscriptionClient.CheckUserExists(context.Background(), &pb.CheckUserExistsRequest{
		UserId: unfollowReq.UnfollowingUserID,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.UnfollowResponse{}, ErrUsertNotFound
		}
		log.Println("inter service call for check user exist failed, error: ", err)
		return responsemodels.UnfollowResponse{}, err
	}
	unfollowRes, err := as.PostRelationRepository.UnfollowUserById(unfollowReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("is it actually")
			return responsemodels.UnfollowResponse{}, domain.ErrNoFollower
		}
		return responsemodels.UnfollowResponse{}, err
	}

	//invalidate cach
	versionKey := fmt.Sprintf("user:%d:feed_version", unfollowReq.UserID)
	_, _ = as.RedisRepository.Incr(context.Background(), versionKey)

	return responsemodels.UnfollowResponse{
		UnfollowingUserID: unfollowRes.UnfollowingUserID,
	}, nil
}

func (as *PostRelationUsecase) FetchComments(fetchCommentsReq requestmodels.FetchCommentsReqeust) (responsemodels.FetchCommentsResponse, error) {
	commentsRes, err := as.PostRelationRepository.FetchCommentsByPostId(fetchCommentsReq)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FetchCommentsResponse{}, ErrNoComments
		}
		return responsemodels.FetchCommentsResponse{}, err
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
	userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		UserId: userids,
	})
	//v:=userResp[userIDs]
	if err != nil {
		log.Println("error calling service auth_subcription", err)
		return responsemodels.FetchCommentsResponse{}, err
	}
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
				return responsemodels.FetchCommentsResponse{}, errors.New("invalid parent comment id")
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
	fmt.Println("print post Count in usecase", postCount)
	resp, err := as.PostRelationRepository.FetchFollowCountByUserId(userid)
	if err != nil {
		return responsemodels.PostFollowCountResponse{}, err
	}
	fmt.Println("resp print first in usecase", resp)
	resp.PostCount = postCount
	fmt.Println("resp print second in usecase", resp, resp.PostCount)
	return resp, nil
}
func (as *PostRelationUsecase) FetchAllPosts(currentuserid uint64, targetuserid uint64) ([]responsemodels.PostWithCounts, error) {
	resp, err := as.PostRelationRepository.FetchAllPosts(currentuserid, targetuserid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoPosts
		}
		return nil, err
	}
	for i := range resp {
		resp[i].Age = utils.CalcuateCommentAge(resp[i].CreatedAt)
	}
	return resp, nil
}
func (as *PostRelationUsecase) FetchFollowers(userid uint64) (responsemodels.FetchFollowersResponse, error) {
	resp, err := as.PostRelationRepository.FetchFollowersUserIds(userid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FetchFollowersResponse{}, domain.ErrNoFollowers
		}
		return responsemodels.FetchFollowersResponse{}, err
	}
	var userids []uint64
	for _, v := range resp {
		userids = append(userids, v.FollowerID)
	}
	userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		UserId: userids,
	})
	if err != nil {
		log.Println("error calling service auth_subcription", err)
		return responsemodels.FetchFollowersResponse{}, err
	}
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
func (as *PostRelationUsecase) FetchFollowing(userid uint64) (responsemodels.FetchFollowingResponse, error) {
	resp, err := as.PostRelationRepository.FetchFollowingUserIds(userid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responsemodels.FetchFollowingResponse{}, domain.ErrNoFollowing
		}
		return responsemodels.FetchFollowingResponse{}, err
	}
	var userids []uint64
	for _, v := range resp {
		userids = append(userids, v.FollowingID)
	}
	userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		UserId: userids,
	})
	if err != nil {
		log.Println("error calling service auth_subcription", err)
		return responsemodels.FetchFollowingResponse{}, err
	}
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
func (as *PostRelationUsecase) FetchPostUserDataForNewsFeed(newsfeedReq requestmodels.FetchNewsFeedRequest) (responsemodels.FetchNewsFeedResponse, error) {
	ctx := context.Background()
	version := as.getFeedVersion(ctx, newsfeedReq.UserID)

	// Key now uses LastID instead of Offset
	cacheKey := fmt.Sprintf("newsfeed:%d:v:%s:lim:%d:last:%d",
		newsfeedReq.UserID, version, newsfeedReq.Limit, newsfeedReq.LastID)

	if newsfeedReq.PullToRefresh {
		// Increment version to effectively "clear" all pages at once
		versionKey := fmt.Sprintf("user:%d:feed_version", newsfeedReq.UserID)
		version, _ = as.RedisRepository.Incr(context.Background(), versionKey) // Need to add Incr to your interface
		//if err != nil {
		//  version = "1" // Fallback
		newsfeedReq.LastID = 0
		cacheKey = fmt.Sprintf("newsfeed:%d:v:%s:lim:%d:last:0", newsfeedReq.UserID, version, newsfeedReq.Limit)
	} else {
		cachedData, err := as.RedisRepository.CacheGet(ctx, cacheKey)
		if err == nil {
			var cachedResp responsemodels.FetchNewsFeedResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
				fmt.Println("getting cached response")
				return cachedResp, nil
			}
		}
	}

	// 2. CACHE MISS: Execute your existing logic
	postResp, err := as.PostRelationRepository.FetchPostDataForNewsFeed(newsfeedReq)
	if err != nil {
		return responsemodels.FetchNewsFeedResponse{}, err
	}
	if len(postResp) == 0 {
		if newsfeedReq.LastID == 0 {
			return responsemodels.FetchNewsFeedResponse{}, domain.ErrNoFollowingNoPost
		}
		// Return empty response with HasMore false
		return responsemodels.FetchNewsFeedResponse{HasMore: false}, nil
	}
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

	userResp, err := as.AuthSubscriptionClient.FetchUserMetaData(context.Background(), &pb.UserDataReq{
		UserId: userids,
	})
	if err != nil {
		log.Println("error calling service auth_subcription", err)
		return responsemodels.FetchNewsFeedResponse{}, err
	}
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

	var nextCursor uint64
	hasMore := false
	if len(postResp) > int(newsfeedReq.Limit) {
		hasMore = true
		// Remove the extra item so the user only gets the 10 they asked for
		postResp = postResp[:newsfeedReq.Limit]
	}
	if len(postResp) > 0 {
		// The ID of the last item in our result is the cursor for the next request
		nextCursor = uint64(postResp[len(postResp)-1].ID)
	}

	finalResponse := responsemodels.FetchNewsFeedResponse{
		PostUserData: postResp,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
	}
	// 3. Store in Redis for future requests (e.g., 5 minutes TTL)
	dataToCache, err := json.Marshal(finalResponse)
	if err != nil {
		return responsemodels.FetchNewsFeedResponse{}, err
	}
	err = as.RedisRepository.CacheSet(context.Background(), cacheKey, dataToCache, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to cache newsfeed for key %s: %v", cacheKey, err)
		// Note: Usually we don't return an error here because we still have the data
		// to return to the user; the cache failing shouldn't break the whole app.
	}
	log.Println("returning sql response")
	return finalResponse, nil
}
func (as *PostRelationUsecase) getFeedVersion(ctx context.Context, userID uint64) string {
	versionKey := fmt.Sprintf("user:%d:feed_version", userID)
	version, err := as.RedisRepository.CacheGet(ctx, versionKey)
	if err != nil || len(version) == 0 {
		// If no version exists, start at 1
		as.RedisRepository.CacheSet(ctx, versionKey, []byte("1"), 48*time.Hour)
		fmt.Println(err, len(version))
		return "1"
	}
	// OPTIONAL: Refresh the 48h timer so active users never lose their version
	err = as.RedisRepository.ExtendTTL(ctx, versionKey, 48*time.Hour)
	if err != nil {
		log.Println("failed to extend ttl")
	}
	return string(version)
}
