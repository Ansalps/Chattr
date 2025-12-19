package repository

import (
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
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
	var postId uint64
	query := `INSERT INTO posts (created_at,updated_at,user_id,caption) VALUES ($1,$2,$3,$4) RETURNING id`
	err := ad.DB.Raw(query, time.Now(), time.Now(), createPostReq.UserID, createPostReq.Caption).Scan(&postId).Error
	if err != nil {
		return responsemodels.CreatePostResponse{}, err
	}
	mediaInsertQuery := `INSERT INTO post_media (create_at,updated_at,post_id,media_url) VALUES ($1,$2,$3,$4)`
	for _, url := range createPostReq.MediaUrls {
		errIns := ad.DB.Exec(mediaInsertQuery, time.Now(), time.Now(), postId, url).Error
		if errIns != nil {
			return responsemodels.CreatePostResponse{}, errIns
		}
	}
	return responsemodels.CreatePostResponse{
		PostID: postId,
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
	if err := ad.DB.Exec(query, deletePostReq.UserID, deletePostReq.PostID).Error; err != nil {
		return responsemodels.DeletePostResponse{}, err
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
	query := `INSERT INTO comments (created_at,updated_at,user_id,post_id,comment_text,parent_comment_id) VALUES ($1,$2,$3,$4,$5,$6)`
	if err := ad.DB.Exec(query, time.Now(), time.Now(), addCommentReq.UserID, addCommentReq.PostID, addCommentReq.CommentText, addCommentReq.ParentCommentId).Error; err != nil {
		return responsemodels.AddCommentResponse{}, err
	}
	return responsemodels.AddCommentResponse{
		UserID:          addCommentReq.UserID,
		PostID:          addCommentReq.PostID,
		CommentText:     addCommentReq.CommentText,
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

func (ad *PostRelationRepository) FetchCommentsByPostId(fetchCommentsReq requestmodels.FetchCommentsReqeust) (responsemodels.FetchCommentsResponse, error) {
	var comments []responsemodels.Comment
	query := `SELECT id,comment_text FROM comments WHERE post_id=$1 WHERE parent_comment_id IS NULL`
	resutl := ad.DB.Raw(query, fetchCommentsReq.PostID).Scan(&comments)
	if resutl.Error != nil {
		return responsemodels.FetchCommentsResponse{}, resutl.Error
	}
	return responsemodels.FetchCommentsResponse{
		Comments: comments,
	}, nil
}
func (ad *PostRelationRepository) FetchCommentsOfCommentByParentCommentId(fetchCommentsOfCommentReq requestmodels.FetchCommentsOfCommentReqeust) (responsemodels.FetchCommentsOfCommentResponse, error) {
	var comments []responsemodels.Comment
	query := `SELECT id,comment_text FROM comments WHERE post_id=$1 AND parent_comment_id=$2`
	result := ad.DB.Raw(query, fetchCommentsOfCommentReq.PostID, fetchCommentsOfCommentReq.ParentCommentId).Scan(&comments)
	if result.Error != nil {
		return responsemodels.FetchCommentsOfCommentResponse{}, result.Error
	}
	// fmt.Println(result)
	// fmt.Pin
	return responsemodels.FetchCommentsOfCommentResponse{
		Comments: comments,
	}, nil
}
func (ad *PostRelationRepository)FetchPostCountByUserId(userid uint64)(uint64,error){
	var postCount uint64
	query:=`SELECT COUNT(*) as post_count FROM posts WHERE user_id=$1`
	result:=ad.DB.Raw(query,userid).Scan(&postCount)
	if result.Error!=nil{
		return 0,result.Error
	}
	return postCount,nil
}
func (ad *PostRelationRepository)FetchFollowCountByUserId(userid uint64)(responsemodels.PostFollowCountResponse,error){
	var resp responsemodels.PostFollowCountResponse
	query:=`SELECT COUNT(*) FILTER (WHERE following_id=$1) AS follower_count,COUNT(*) FILTER (WHERE follower_id=$2) AS following_count FROM relations`
	result:=ad.DB.Raw(query,userid,userid).Scan(&resp)
	if result.Error!=nil{
		return responsemodels.PostFollowCountResponse{},result.Error
	}
	return resp,nil
}
