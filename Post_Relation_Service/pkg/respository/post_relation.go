package repository

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type PostRelationRepository struct {
	DB *gorm.DB
}

func NewPostRelationRepository(db *gorm.DB) interfacesRepository.PostRelationRepository {
	return &PostRelationRepository{
		DB: db,
	}
}

func (ad *PostRelationRepository) CreatePost(createPostReq requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse, error) {
	var mediaRecords []domain.PostMedia
	for _, url := range createPostReq.MediaUrls {
		mediaRecords = append(mediaRecords, domain.PostMedia{MediaUrl: url})
	}
	newPost := domain.Post{
		UserID:    uint(createPostReq.UserID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Caption:   createPostReq.Caption,
		Media:     mediaRecords,
	}
	err := ad.DB.Create(&newPost).Error
	if err != nil {
		return responsemodels.CreatePostResponse{}, err
	}

	return responsemodels.CreatePostResponse{
		PostID: uint64(newPost.ID),
	}, nil
}

func (ad *PostRelationRepository) EditPostById(editPostReq requestmodels.EditPostRequest) (responsemodels.EditPostResponse, error) {
	query := `UPDATE posts SET caption=? WHERE user_id=? and id=?`
	result := ad.DB.Exec(query, editPostReq.Caption, editPostReq.UserID, editPostReq.PostID)
	if result.Error != nil {
		return responsemodels.EditPostResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.EditPostResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.EditPostResponse{
		Caption: editPostReq.Caption,
	}, nil
}
func (ad *PostRelationRepository) DeletePostById(deletePostReq requestmodels.DeletePostRequest) (responsemodels.DeletePostResponse, error) {
	query := `DELETE FROM posts WHERE user_id=? and id=?`
	result := ad.DB.Exec(query, deletePostReq.UserID, deletePostReq.PostID)
	if result.Error != nil {
		return responsemodels.DeletePostResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.DeletePostResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.DeletePostResponse{
		PostID: deletePostReq.PostID,
	}, nil
}

func (ad *PostRelationRepository) LikePostById(likePostReq requestmodels.LikePostRequest) (responsemodels.LikePostResponse, error) {
	query := `INSERT INTO post_likes (user_id,post_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
	result := ad.DB.Exec(query, likePostReq.UserID, likePostReq.PostID)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23503" {
			return responsemodels.LikePostResponse{}, domain.ErrForeignKeyViolationCommentPost
		}
		return responsemodels.LikePostResponse{}, fmt.Errorf("%w: %v",domain.ErrDatabase,result.Error)
	}
	return responsemodels.LikePostResponse{
		PostID: likePostReq.PostID,
	}, nil
}

func (ad *PostRelationRepository) FetchPostOwnerIdByPostId(postId uint64) (uint64, error) {
	var postOwnerId uint64
	query := `select user_id from posts where id=?`
	result := ad.DB.Raw(query, postId).Scan(&postOwnerId)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return postOwnerId, nil
}

func (ad *PostRelationRepository) UnlikePostById(unlikePostReq requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error) {
	query := `DELETE FROM post_likes WHERE user_id=? AND post_id=?`
	result := ad.DB.Exec(query, unlikePostReq.UserID, unlikePostReq.PostID)
	if result.Error != nil {
		return responsemodels.UnlikePostResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.UnlikePostResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.UnlikePostResponse{
		PostID: unlikePostReq.PostID,
	}, nil
}

// func (ad *PostRelationRepository) CheckCommentHieracrchy(commentId *uint64) (bool, error) {
// 	var parentId *uint64
// 	query := `SELECT parent_comment_id FROM comments WHERE id=?`
// 	res := ad.DB.Raw(query, commentId).Scan(&parentId)
// 	if res.Error!= nil {
// 		fmt.Println("is it reaching in error")
// 		return false, res.Error
// 	}
// 	if res.RowsAffected==0{
// 		fmt.Println("no rows affected, comment id might be wrong")
// 		return false,gorm.ErrRecordNotFound
// 	}
// 	if parentId == nil {
// 		fmt.Println("or is parentId just nil")
// 		return false, nil
// 	}
// 	return true, nil
// }
func (ad *PostRelationRepository) IsSubComment(commentId *uint64) (bool, error) {
    // We use a struct or a map to capture the result safely
    var result struct {
        ParentCommentID *uint64 `gorm:"column:parent_comment_id"`
    }
    
    // .Table() specifies the table
    // .Select() picks only the column we need
    // .Where() filters by the comment the user is trying to reply to
    // .Take() fetches one record or returns gorm.ErrRecordNotFound if empty
    err := ad.DB.Table("comments").
        Select("parent_comment_id").
        Where("id = ?", commentId).
        Take(&result).Error

    if err != nil {
        // This will now properly catch cases where the comment doesn't exist
        return false, err 
    }

    // Logic: If ParentCommentID is NOT nil, it means the comment found
    // is already a sub-comment.
    return result.ParentCommentID != nil, nil
}
func (ad *PostRelationRepository) AddComment(addCommentReq requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error) {
	var commetId uint64
	query := `INSERT INTO comments (created_at,updated_at,user_id,post_id,comment_text,parent_comment_id) VALUES ($1,$2,$3,$4,$5,$6) returning id`
	result := ad.DB.Raw(query, time.Now(), time.Now(), addCommentReq.UserID, addCommentReq.PostID, addCommentReq.CommentText, addCommentReq.ParentCommentId).Scan(&commetId)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23503" {
			return responsemodels.AddCommentResponse{}, domain.ErrForeignKeyViolationCommentPost
		}
		return responsemodels.AddCommentResponse{}, fmt.Errorf("%w: %v",domain.ErrDatabase,result.Error)
	}
	return responsemodels.AddCommentResponse{
		UserID:          addCommentReq.UserID,
		PostID:          addCommentReq.PostID,
		CommentText:     addCommentReq.CommentText,
		CommentID:       commetId,
		ParentCommentId: addCommentReq.ParentCommentId,
	}, nil
}
func (ad *PostRelationRepository) EditComment(editCommentReq requestmodels.EditCommentRequest) (responsemodels.EditCommentResponse, error) {
	query := `UPDATE comments SET comment_text=?,updated_at=? WHERE user_id=? AND post_id=? AND id=?`
	result := ad.DB.Exec(query, editCommentReq.CommentText, time.Now(), editCommentReq.UserID, editCommentReq.PostID, editCommentReq.CommentID)
	if result.Error != nil {
		return responsemodels.EditCommentResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.EditCommentResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.EditCommentResponse{
		PostID:      editCommentReq.PostID,
		CommentID:   editCommentReq.CommentID,
		CommentText: editCommentReq.CommentText,
	}, nil
}
func (ad *PostRelationRepository) DeleteCommentById(deleteCommentReq requestmodels.DeleteCommentRequest) (responsemodels.DeleteCommentResponse, error) {
	query := `DELETE FROM comments WHERE user_id=? and post_id=? and id=?`
	result := ad.DB.Exec(query, deleteCommentReq.UserID, deleteCommentReq.PostID, deleteCommentReq.CommentID)
	if result.Error != nil {
		return responsemodels.DeleteCommentResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		//fmt.Println("is it really happening in database?")
		return responsemodels.DeleteCommentResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.DeleteCommentResponse{
		CommentID: deleteCommentReq.CommentID,
	}, nil
}
func (ad *PostRelationRepository) Follow(followReq requestmodels.FollowRequest) (responsemodels.FollowResponse, error) {
	query := `INSERT INTO relations (follower_id,following_id,created_at,updated_at) VALUES ($1,$2,$3,$4) 
	ON CONFLICT (follower_id,following_id) DO NOTHING`
	result := ad.DB.Exec(query, followReq.UserID, followReq.FollowingUserID, time.Now(), time.Now())
	if result.Error != nil {
		return responsemodels.FollowResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.FollowResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.FollowResponse{
		FollowingUserID: followReq.FollowingUserID,
	}, nil
}
func (ad *PostRelationRepository) UnfollowUserById(unfollowReq requestmodels.UnfollowRequest) (responsemodels.UnfollowResponse, error) {
	query := `DELETE FROM relations WHERE follower_id=$1 AND following_id=$2`
	result := ad.DB.Exec(query, unfollowReq.UserID, unfollowReq.UnfollowingUserID)
	if result.Error != nil {
		return responsemodels.UnfollowResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.UnfollowResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.UnfollowResponse{
		UnfollowingUserID: unfollowReq.UnfollowingUserID,
	}, nil
}

func (ad *PostRelationRepository) FetchCommentsByPostId(fetchCommentsReq requestmodels.FetchCommentsReqeust) ([]responsemodels.Comments, error) {
	var resp []responsemodels.Comments
	query := `SELECT * FROM comments WHERE post_id=$1 limit $2 offset $3`
	result := ad.DB.Raw(query, fetchCommentsReq.PostID, fetchCommentsReq.Limit, fetchCommentsReq.Offset).Scan(&resp)
	if result.Error != nil {
		return []responsemodels.Comments{}, result.Error
	}
	if result.RowsAffected == 0 {
		return []responsemodels.Comments{}, gorm.ErrRecordNotFound
	}
	return resp, nil
}

func (ad *PostRelationRepository) FetchPostCountByUserId(userid uint64) (uint64, error) {
	var postCount uint64
	query := `SELECT COUNT(*) as post_count FROM posts WHERE user_id=$1`
	result := ad.DB.Raw(query, userid).Scan(&postCount)
	if result.Error != nil {
		return 0, result.Error
	}
	return postCount, nil
}
func (ad *PostRelationRepository) FetchFollowCountByUserId(userid uint64) (responsemodels.PostFollowCountResponse, error) {
	var resp responsemodels.PostFollowCountResponse
	query := `SELECT COUNT(*) FILTER (WHERE following_id=$1) AS follower_count,COUNT(*) FILTER (WHERE follower_id=$2) AS following_count FROM relations`
	result := ad.DB.Raw(query, userid, userid).Scan(&resp)
	if result.Error != nil {
		return responsemodels.PostFollowCountResponse{}, result.Error
	}
	return resp, nil
}

func (as *PostRelationRepository) FetchPostByPostID(req requestmodels.FetchPostByPostIDRequest) (responsemodels.PostWithCounts, error) {
	var post responsemodels.PostWithCounts

	err := as.DB.Model(&domain.Post{}).
		Select("posts.*, "+
			"(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count, "+
			"(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count, "+
			"EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?) as is_liked",
			req.UserID).
		Where("id = ?", req.PostID).
		Preload("Media").
		First(&post).Error
	if err != nil {
		return responsemodels.PostWithCounts{}, err
	}
	return post, nil
}
func (ad *PostRelationRepository) FetchAllPosts(req requestmodels.FetchAllPostsReq) ([]responsemodels.PostWithCounts, error) {
	var posts []responsemodels.PostWithCounts

	err := ad.DB.Model(&domain.Post{}).
		// 1. Select all post fields + Subqueries for counts
		Select("posts.*, "+
			"(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count, "+
			"(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count, "+
			// "Is Liked" Subquery (Returns true if record exists)
			"EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?) as is_liked",
			req.CurrentUserID). // Pass the logged-in user's ID here).
		// 2. Filter by User
		Where("user_id = ?", req.TargetUserID).
		// 3. Still Preload your Media slice
		Preload("Media").
		Order("created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&posts).Error
	if err!=nil{
		return []responsemodels.PostWithCounts{},err
	}
	return posts, err
}

func (ad *PostRelationRepository) FetchFollowersUserIds(userid uint64) ([]responsemodels.FollowerIds, error) {
	var resp []responsemodels.FollowerIds
	query := `SELECT follower_id FROM relations WHERE following_id=$1`
	result := ad.DB.Raw(query, userid).Scan(&resp)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return resp, nil
}
func (ad *PostRelationRepository) FetchFollowersUserIds1(req requestmodels.FetchFollowersRequest) ([]responsemodels.FollowerIds, error) {
	var resp []responsemodels.FollowerIds
	query := `SELECT follower_id FROM relations WHERE following_id=$1 LIMIT $2 OFFSET $3`
	result := ad.DB.Raw(query, req.UserID, req.Limit, req.Offset).Scan(&resp)
	if result.Error != nil {
		return nil, result.Error
	}
	// if result.RowsAffected == 0 {
	// 	return nil, gorm.ErrRecordNotFound
	// }
	return resp, nil
}
func (ad *PostRelationRepository) FetchFollowingUserIds(req requestmodels.FetchFollowingRequest) ([]responsemodels.FollowingIds, error) {
	var resp []responsemodels.FollowingIds
	query := `SELECT following_id FROM relations WHERE follower_id=$1 LIMIT $2 OFFSET $3`
	result := ad.DB.Raw(query, req.UserID, req.Limit, req.Offset).Scan(&resp)
	if result.Error != nil {
		//fmt.Println("is it reaching in error")
		return nil, result.Error
	}
	// if result.RowsAffected == 0 {
	// 	return nil, gorm.ErrRecordNotFound
	// }
	return resp, nil
}
func (ad *PostRelationRepository) FetchPostData(newsfeedReq requestmodels.FetchNewsFeedRequest) ([]responsemodels.PostWithStatus, error) {
	var resp []responsemodels.PostWithStatus

	// Subquery to get Followings who are NOT celebrities
	normalFollowingSubquery := ad.DB.Table("relations").
		Select("following_id").
		Where("follower_id = ?", newsfeedReq.UserID)
		//fmt.Println("user id",newsfeedReq.UserID)
	query := ad.DB.Table("posts").
		Select(`posts.*, 
		(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count,
		(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count,
		EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?)  as is_liked
		`, newsfeedReq.UserID). // Use your existing SELECT logic
		Where("(posts.user_id = ? OR posts.user_id IN (?))", newsfeedReq.UserID, normalFollowingSubquery).
		Where("posts.post_status = ?", "normal")

	if newsfeedReq.LastID > 0 {
		query = query.Where("posts.id < ?", newsfeedReq.LastID)
	}

	query1 := query.Order("posts.id DESC").Limit(int(newsfeedReq.Limit) + 1).Preload("Media").Find(&resp)
	if query1.Error!=nil{
		return nil,fmt.Errorf("error in fetch normal post data: %w",query1.Error)
	}
	return resp, nil
}

func (ad *PostRelationRepository) FetchNormalPostData(newsfeedReq requestmodels.FetchNewsFeedRequest) ([]responsemodels.PostWithStatus, error) {
	var resp []responsemodels.PostWithStatus

	// Subquery to get Followings who are NOT celebrities
	normalFollowingSubquery := ad.DB.Table("relations").
		Select("following_id").
		Where("follower_id = ? AND following_id NOT IN (SELECT id FROM celebrities)", newsfeedReq.UserID)
		//fmt.Println("user id",newsfeedReq.UserID)
	query := ad.DB.Table("posts").
		Select(`posts.*, 
		(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count,
		(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count,
		EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?)  as is_liked
		`, newsfeedReq.UserID). // Use your existing SELECT logic
		Where("(posts.user_id = ? OR posts.user_id IN (?))", newsfeedReq.UserID, normalFollowingSubquery).
		Where("posts.post_status = ?", "normal")

	if newsfeedReq.LastID > 0 {
		query = query.Where("posts.id < ?", newsfeedReq.LastID)
	}

	query1 := query.Order("posts.id DESC").Limit(int(newsfeedReq.Limit) + 1).Preload("Media").Find(&resp)
	if query1.Error!=nil{
		return nil,fmt.Errorf("error in fetch normal post data: %w",query1.Error)
	}
	return resp, nil
}
func (ad *PostRelationRepository) GetFollowedNormalUsersIDs(userID uint64) ([]uint64, error) {
	var normalUserIDs []uint64
	query := ad.DB.Table("relations").
		Select("following_id").
		Where("follower_id = ? AND following_id NOT IN (SELECT id FROM celebrities)", userID).
		Pluck("following_id", &normalUserIDs)
	if query.Error!=nil{
		return nil,fmt.Errorf("error fething followed celebrity ids: %w",query.Error)
	}
	return normalUserIDs, nil
}
func (ad *PostRelationRepository) GetFollowedCelebrityIDs(userID uint64) ([]uint64, error) {
	var celebIDs []uint64
	query := ad.DB.Table("relations").
		Select("following_id").
		Where("follower_id = ? AND following_id IN (SELECT id FROM celebrities)", userID).
		Pluck("following_id", &celebIDs)
	if query.Error!=nil{
		return nil,fmt.Errorf("error fething followed celebrity ids: %w",query.Error)
	}
	return celebIDs, nil
}
func (ad *PostRelationRepository) FetchPostsByIDs(postIDs []uint64, viewerID uint64) ([]responsemodels.PostWithStatus, error) {
	var resp []responsemodels.PostWithStatus
	if len(postIDs) == 0 {
		return resp, nil
	}

	query1 := ad.DB.Table("posts").
		Select(`posts.*, 
		(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count,
		(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count,
		EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?)  as is_liked
		`, viewerID). // Existing SELECT with counts
		Where("id IN ?", postIDs).
		Preload("Media").
		Order("id DESC"). // Keep them sorted for the merge
		Find(&resp)
	if query1.Error!=nil{
		return nil,fmt.Errorf("error in fetching posts by id: %w",query1.Error)
	}
	return resp, nil
}
func (ad *PostRelationRepository) FetchCelebrityPostIDsFromSQL(celebIDs []uint64, lastID uint64, limit int) ([]uint64, error) {
	fmt.Println("here----------")
	var postIDs []uint64

	query := ad.DB.Table("posts").
		Select("id").
		Where("user_id IN ? AND post_status = 'normal'", celebIDs)

	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}

	// We pull 'limit' posts per request to ensure we have enough to merge
	query1 := query.Order("id DESC").Limit(limit).Pluck("id", &postIDs)
	if query1.Error!=nil{
		fmt.Println("hi-----------",query1.Error)
		return nil, fmt.Errorf("error in fetching celebrity post ids: %w",query1.Error)
	}
	fmt.Println("hello----------")
	return postIDs, nil
}
func (ad *PostRelationRepository) FetchLatestPostIDsByUserID(userID uint64, limit int) ([]uint64, error) {
	var ids []uint64
	err := ad.DB.Table("posts").
		Select("id").
		Where("user_id = ? AND post_status = 'normal'", userID).
		Order("id DESC").
		Limit(limit).
		Pluck("id", &ids).Error
	return ids, err
}
func (ad *PostRelationRepository) PromoteToCelebrity(userid uint64) error {
	query := `INSERT INTO celebrities (id,created_at) VALUES ($1,$2)`
	result := ad.DB.Exec(query, userid, time.Now())
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func (ad *PostRelationRepository) IsUserCelebrity(userid uint64)(bool,error){
	fmt.Println("user id",userid)
	var exists int
	query := `SELECT 1 FROM celebrities WHERE id=$1 LIMIT 1`

	result := ad.DB.Raw(query, userid).Scan(&exists)
	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected == 0 {
		fmt.Println("no rows found")
		return false, nil
	}
	return true,nil
}
func (ad *PostRelationRepository) DepromoteToNormalUser(userid uint64) error {
	query := `DELETE FROM celebrities WHERE id=$1`
	result := ad.DB.Exec(query, userid)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (ad *PostRelationRepository) FetchGlobalTrendingSQL(req requestmodels.GlobalNewsFeedRequest) ([]responsemodels.PostWithStatusWithTrendingScore, error) {
	var posts []responsemodels.PostWithStatusWithTrendingScore

	// err := ad.DB.Model(&domain.Post{}).
	// 	// 1. Select all post fields + Subqueries for counts
	// 	Select("posts.*, "+
	// 		"(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count, "+
	// 		"(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count, "+
	// 		"("+
	// 		"  ("+
	// 		"    (SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) + "+
	// 		"    (SELECT COUNT(*) FROM comments   WHERE comments.post_id   = posts.id) "+
	// 		"  ) / "+
	// 		"  (EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600 + 1) "+
	// 		") AS trending_score, "+
	// 		// "Is Liked" Subquery (Returns true if record exists)
	// 		"EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?) as is_liked",
	// 		req.UserID). // Pass the logged-in user's ID here).
	// 	// 2. Filter by User
	// 	//Where("user_id = ?", req.TargetUserID).
	// 	// 3. Still Preload your Media slice
	// 	Preload("Media").
	// 	Order("trending_score DESC").
	// 	Limit(req.Limit).
	// 	Offset(int(req.Offset)).
	// 	Find(&posts).Error

	err := ad.DB.Raw(`
    WITH CalculatedStats AS (
        SELECT 
            p.id as post_id,
            COUNT(DISTINCT pl.user_id) as l_count,
            COUNT(DISTINCT c.id) as c_count,
            EXISTS(SELECT 1 FROM post_likes WHERE post_id = p.id AND user_id = ?) as is_liked
        FROM posts p
        LEFT JOIN post_likes pl ON pl.post_id = p.id
        LEFT JOIN comments c ON c.post_id = p.id
        GROUP BY p.id
    )
    SELECT 
        posts.*, 
        cs.l_count as likes_count, 
        cs.c_count as comments_count, 
        cs.is_liked,
        ( 
			(CAST(cs.l_count AS FLOAT) + CAST(cs.c_count AS FLOAT)) / 
			(EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600 + 1) 
		  ) as trending_score
    FROM posts
    JOIN CalculatedStats cs ON cs.post_id = posts.id
    ORDER BY trending_score DESC
    LIMIT ? OFFSET ?
`, req.UserID, req.Limit, req.Offset).
Preload("Media"). // GORM can still Preload as long as we SELECT posts.*
Find(&posts).Error

	if err != nil {
		log.Println("database error in global newsfeed", err)
		return []responsemodels.PostWithStatusWithTrendingScore{}, err
	}

	return posts, nil
}

func (r *PostRelationRepository) UpdatFollowCountOnFollow(
	followerID uint64,
	followingID uint64,
) (uint64, error) {

	var followerCount uint64

	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// 1. increase follower count of the user being followed
	err := tx.Raw(`
		UPDATE follow_counts
		SET follow_count = follow_count + 1,
		    updated_at = NOW()
		WHERE user_id = ?
		RETURNING follow_count
	`, followingID).Scan(&followerCount).Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// 2. increase following count of follower
	err = tx.Exec(`
		UPDATE follow_counts
		SET following_count = following_count + 1,
		    updated_at = NOW()
		WHERE user_id = ?
	`, followerID).Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return followerCount, nil
}

func (r *PostRelationRepository) UpdatFollowCountOnUnFollow(
	followerID uint64,
	followingID uint64,
) (uint64, error) {

	var followerCount uint64

	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// 1. decrease follower count of the user being unfollowed
	err := tx.Raw(`
		UPDATE follow_counts
		SET follow_count = GREATEST(follow_count - 1, 0),
		    updated_at = NOW()
		WHERE user_id = ?
		RETURNING follow_count
	`, followingID).Scan(&followerCount).Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// 2. decrease following count of follower
	err = tx.Exec(`
		UPDATE follow_counts
		SET following_count = GREATEST(following_count - 1,0),
		    updated_at = NOW()
		WHERE user_id = ?
	`, followerID).Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return followerCount, nil
}

func (ad *PostRelationRepository)InsertUserIntoFollowCount(userid uint64)(error){
	query:=`INSERT INTO follow_counts (user_id,follow_count,following_count,updated_at) VALUES ($1,0,0,NOW())`
	err:=ad.DB.Exec(query,userid).Error
	if err!=nil{
		return err
	}
	return nil
}