package repository

import (
	"errors"
	"fmt"
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
	fmt.Println("if it comes here print list of urls",createPostReq.MediaUrls)
	var mediaRecords []domain.PostMedia
	for _,url:=range createPostReq.MediaUrls{
		mediaRecords=append(mediaRecords, domain.PostMedia{MediaUrl: url})
	}
	newPost:=domain.Post{
		UserID: uint(createPostReq.UserID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Caption: createPostReq.Caption,
		Media: mediaRecords,// GORM will see this and handle the insertion
	}
	// 2. Single call to Create
	// GORM opens a transaction, inserts Post, gets ID, and batch inserts Media
	// GORM will start a transaction, save the Post, 
    // grab the new Post.ID, assign it to all mediaRecords.PostID,
    // and set timestamps for EVERYTHING.
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
	if err := ad.DB.Exec(query, editPostReq.Caption, editPostReq.UserID, editPostReq.PostID).Error; err != nil {
		return responsemodels.EditPostResponse{}, err
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
	if err := ad.DB.Exec(query, likePostReq.UserID, likePostReq.PostID).Error; err != nil {
		return responsemodels.LikePostResponse{}, err
	}
	return responsemodels.LikePostResponse{
		PostID: likePostReq.PostID,
	}, nil
}

func (ad *PostRelationRepository) UnlikePostById(unlikePostReq requestmodels.UnlikePostRequest) (responsemodels.UnlikePostResponse, error) {
	query := `DELETE FROM post_likes WHERE user_id=? AND post_id=?`
	if err := ad.DB.Exec(query, unlikePostReq.UserID, unlikePostReq.PostID).Error; err != nil {
		return responsemodels.UnlikePostResponse{}, err
	}
	return responsemodels.UnlikePostResponse{
		PostID: unlikePostReq.PostID,
	}, nil
}

func (ad *PostRelationRepository) CheckCommentHieracrchy(commentId *uint64) (bool, error) {
	var parentId *uint64
	query := `SELECT parent_comment_id FROM comments WHERE id=?`
	if err := ad.DB.Raw(query, commentId).Scan(&parentId).Error; err != nil {
		return false, err
	}
	if parentId == nil {
		return false, nil
	}
	return true, nil
}
func (ad *PostRelationRepository) AddComment(addCommentReq requestmodels.AddCommentRequest) (responsemodels.AddCommentResponse, error) {
	var commetId uint64
	query := `INSERT INTO comments (created_at,updated_at,user_id,post_id,comment_text,parent_comment_id) VALUES ($1,$2,$3,$4,$5,$6) returning id`
	result := ad.DB.Raw(query, time.Now(), time.Now(), addCommentReq.UserID, addCommentReq.PostID, addCommentReq.CommentText, addCommentReq.ParentCommentId).Scan(&commetId)
	if result.Error != nil {
		fmt.Println("heelllo at least here")
		fmt.Printf("Error Type: %T\n", result.Error)
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23503" {
			fmt.Println("is it reaching in postgres err", result.Error)
			return responsemodels.AddCommentResponse{}, domain.ErrForeignKeyViolationCommentPost
		}

		return responsemodels.AddCommentResponse{}, result.Error
	}
	fmt.Println("is comment id printed",commetId)
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
		PostID: editCommentReq.PostID,
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
		fmt.Println("is it really happening in database?")
		return responsemodels.DeleteCommentResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.DeleteCommentResponse{
		CommentID: deleteCommentReq.CommentID,
	}, nil
}
func (ad *PostRelationRepository) Follow(followReq requestmodels.FollowRequest) (responsemodels.FollowResponse, error) {
	query := `INSERT INTO relations (follower_id,following_id,created_at,updated_at) VALUES ($1,$2,$3,$4) 
	ON CONFLICT (follower_id,following_id) DO NOTHING`
	if err := ad.DB.Exec(query, followReq.UserID, followReq.FollowingUserID, time.Now(), time.Now()).Error; err != nil {
		return responsemodels.FollowResponse{}, err
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
	query := `SELECT * FROM comments WHERE post_id=$1`
	result := ad.DB.Raw(query, fetchCommentsReq.PostID).Scan(&resp)
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
// func (ad *PostRelationRepository) FetchAllPosts(userid uint64) (responsemodels.FetchAllPostsResponse, error) {
// 	var resp domain.Post
// 	// query := `SELECT id as post_id,created_at,updated_at,user_id,caption,media_url FROM posts LEFT JOIN post_media 
// 	// ON posts.id=post_media.post_id WHERE user_id=$1`
// 	// result := ad.DB.Raw(query, userid).Scan(&resp)
// 	// if result.Error != nil {
// 	// 	return responsemodels.FetchAllPostsResponse{}, result.Error
// 	// }
// 	// var resp1 []responsemodels.Post
// 	// p := make(map[uint64][]string)
// 	// dup := make(map[uint64]bool)
// 	// for _, v := range resp {
// 	// 	if !dup[v.PostID] {
// 	// 		dup[v.PostID] = true
// 	// 		resp1 = append(resp1, responsemodels.Post{
// 	// 			PostID:    v.PostID,
// 	// 			CreatedAt: v.CreatedAt,
// 	// 			UpdatedAt: v.UpdatedAt,
// 	// 			UserID:    v.UserID,
// 	// 			Caption:   v.Caption,
// 	// 		})
// 	// 	}
// 	// 	p[v.PostID] = append(p[v.PostID], v.MediaUrl)
// 	// }
// 	// for i := range resp1 {
// 	// 	resp1[i].MediaUrls = append(resp1[i].MediaUrls, p[resp1[i].PostID]...)
// 	// }
// 	// return responsemodels.FetchAllPostsResponse{
// 	// 	Posts: resp1,
// 	// }, nil
// 	// Preload("Media") runs the second query automatically
//     err := ad.DB.Preload("Media").First(&resp, postId).Error
//     return resp, err
// }
// func (ad *PostRelationRepository) FetchAllPosts(userid uint64) ([]domain.Post, error) {
//     var posts []domain.Post // Use a slice to hold multiple posts

//     // Preload("Media") matches the field name in your Post struct
//     err := ad.DB.Preload("Media").Where("user_id = ?", userid).Find(&posts).Error
    
//     if err != nil {
//         return nil, err
//     }

//     return posts, nil
// }
func (ad *PostRelationRepository) FetchAllPosts(currentUserID uint64,targetUserID uint64) ([]responsemodels.PostWithCounts, error) {
    var posts []responsemodels.PostWithCounts

    err := ad.DB.Model(&domain.Post{}).
        // 1. Select all post fields + Subqueries for counts
        Select("posts.*, "+
            "(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) as likes_count, "+
            "(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comments_count, "+
			// "Is Liked" Subquery (Returns true if record exists)
            "EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.user_id = ?) as is_liked", 
            currentUserID). // Pass the logged-in user's ID here).
        // 2. Filter by User
        Where("user_id = ?", targetUserID).
        // 3. Still Preload your Media slice
        Preload("Media").
        Order("created_at DESC").
        Find(&posts).Error

    return posts, err
}

func (ad *PostRelationRepository)FetchFollowersUserIds(userid uint64)([]responsemodels.FollowerIds,error){
	var resp []responsemodels.FollowerIds
	query:=`SELECT follower_id FROM relations WHERE following_id=$1`
	result:=ad.DB.Raw(query,userid).Scan(&resp)
	if result.Error!=nil{
		return nil,result.Error
	}
	if result.RowsAffected==0{
		return nil,gorm.ErrRecordNotFound
	}
	return resp,nil
}
func (ad *PostRelationRepository)FetchFollowingUserIds(userid uint64)([]responsemodels.FollowingIds,error){
	var resp []responsemodels.FollowingIds
	query:=`SELECT following_id FROM relations WHERE follower_id=$1`
	result:=ad.DB.Raw(query,userid).Scan(&resp)
	if result.Error!=nil{
		return nil,result.Error
	}
	if result.RowsAffected==0{
		return nil,gorm.ErrRecordNotFound
	}
	return resp,nil
}