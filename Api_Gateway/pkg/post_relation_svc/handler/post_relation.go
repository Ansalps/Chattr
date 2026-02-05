package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	authClient "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/client"
	authResponseModel "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/pb/auth_subscription"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/pb/post_relation"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostRelationHandler struct {
	GPPC_Client      interfaces.PostRelationClientInterface
	config           *config.Config
	DirectAuthClient *authClient.AuthSubscriptionClient
	DirectPostClient *client.PostRelationClient
}

func NewPostRelationHandler(postRelationClient interfaces.PostRelationClientInterface, cfg *config.Config, directAuthClient *authClient.AuthSubscriptionClient, postDirectClient *client.PostRelationClient) *PostRelationHandler {
	return &PostRelationHandler{
		GPPC_Client:      postRelationClient,
		config:           cfg,
		DirectAuthClient: directAuthClient,
		DirectPostClient: postDirectClient,
	}
}

func (as *PostRelationHandler) CreatePost(c *gin.Context) {
	var createPostReq requestmodels.CreatePostRequest
	createPostReq.Caption = c.PostForm("caption")

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	createPostReq.UserID = jwtClaims.ID
	// 1. Parse form
	err := c.Request.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse form"})
		return
	}

	files := c.Request.MultipartForm.File["media"]
	if len(files) < 1 || len(files) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Files count must be between 1 and 5"})
		return
	}

	// Allowed formats
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".mp4": true,
	}
	var uploadedUrls []string

	cld, _ := cloudinary.NewFromParams(
		as.config.Cloudinary.CloundName,
		as.config.Cloudinary.ApiKey,
		as.config.Cloudinary.ApiSecret,
	)

	for _, file := range files {
		// Validate size (<1MB)
		if file.Size > 5<<20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Each file must be < 5 MB"})
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot open file"})
			return
		}
		defer src.Close()

		// Upload to Cloudinary
		uploadResp, err := cld.Upload.Upload(
			c,
			src,
			uploader.UploadParams{
				Folder:       "posts",
				PublicID:     fmt.Sprintf("%d-%s", time.Now().UnixNano(), file.Filename),
				ResourceType: "auto", // auto detects (image/video)
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cloudinary upload failed"})
			return
		}

		uploadedUrls = append(uploadedUrls, uploadResp.SecureURL)
	}
	createPostReq.MediaUrls = uploadedUrls
	createPostResponse, err := as.GPPC_Client.CreatePost(createPostReq)
	if err != nil {

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"urls":    uploadedUrls,
		"res":     createPostResponse,
	})
}

func (as *PostRelationHandler) EditPost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid post id", nil))
		return
	}
	var editPostRequest requestmodels.EditPostRequest
	editPostRequest.PostID = postId
	if err := c.ShouldBindJSON(&editPostRequest); err != nil {
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Request Body", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims Not Found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid Claims", nil))
		return
	}
	editPostRequest.UserID = jwtClaims.ID
	editPostResponse, err := as.GPPC_Client.EditPost(editPostRequest)
	if err != nil {
		fmt.Println("will it reach inside")
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "Post edited Successfully", editPostResponse))
}

func (as *PostRelationHandler) DeletePost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid post id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims Not Found", nil))
		return
	}
	fmt.Println("print claims", claims)
	fmt.Printf("claims type = %T\n", claims)
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalid claims", nil))
		return
	}
	var deletePostReq requestmodels.DeletePostRequest
	deletePostReq.UserID = jwtClaims.ID
	deletePostReq.PostID = postId
	deletePostResponse, err := as.GPPC_Client.DeletePost(deletePostReq)
	if err != nil {
		var obj response.Response
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusPreconditionFailed, st.Message(), nil)
			default:
				obj = response.ClientResponse(http.StatusInternalServerError, "Internal Server Error", nil)
			}
		}
		c.JSON(obj.StatusCode, obj)
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "post deleted successfully", deletePostResponse))
}

func (as *PostRelationHandler) LikePost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid post id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	var likePostReq requestmodels.LikePostRequest
	likePostReq.UserID = jwtClaims.ID
	likePostReq.PostID = postId
	likePostResponse, err := as.GPPC_Client.LikePost(likePostReq)
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "post like successfully", likePostResponse))
}

func (as *PostRelationHandler) UnlikePost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalide post id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalid claims", nil))
		return
	}
	var unlikePostReq requestmodels.UnlikePostRequest
	unlikePostReq.UserID = jwtClaims.ID
	unlikePostReq.PostID = postId
	unlikePostResponse, err := as.GPPC_Client.UnlikePost(unlikePostReq)
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "post unliked successfully", unlikePostResponse))
}

func (as *PostRelationHandler) AddComment(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid Post id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	var addCommentRequest requestmodels.AddCommentRequest
	addCommentRequest.UserID = jwtClaims.ID
	addCommentRequest.PostID = postId
	if err := c.ShouldBindJSON(&addCommentRequest); err != nil {
		if validateioErrors := utils.FormatValidationError(err); validateioErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", validateioErrors))
			return
		}
		log.Println("Bind Error: ", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", err))
		return
	}
	addCommentResponse, err := as.GPPC_Client.AddComment(addCommentRequest)
	if err != nil {
		var obj response.Response
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusNotFound, st.Message(), nil)
			case codes.FailedPrecondition:
				obj = response.ClientResponse(http.StatusPreconditionFailed, st.Message(), nil)
			default:
				obj = response.ClientResponse(http.StatusInternalServerError, "Internal Server Error", nil)
			}
		}
		c.JSON(obj.StatusCode, obj)
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "comment added succesfully", addCommentResponse))
}

func (as *PostRelationHandler) EditComment(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid post di", nil))
		return
	}
	commentIdStr := c.Param("comment_id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid comment id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	var editCommentReq requestmodels.EditCommentRequest
	if err := c.ShouldBindJSON(&editCommentReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "validation failed", validationErrors))
			return
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "bind error", err))
		return
	}
	editCommentReq.UserID = jwtClaims.ID
	editCommentReq.PostID = postId
	editCommentReq.CommentID = commentId
	editCommentResponse, err := as.GPPC_Client.EditComment(editCommentReq)
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return // Stop execution
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "comment edited successfully", editCommentResponse))
}
func (as *PostRelationHandler) DeleteComment(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid post di", nil))
		return
	}
	commentIdStr := c.Param("comment_id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid comment id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	var deletCommentReq requestmodels.DeleteCommentRequest
	deletCommentReq.UserID = jwtClaims.ID
	deletCommentReq.PostID = postId
	deletCommentReq.CommentID = commentId
	deleteCommentRes, err := as.GPPC_Client.DeleteComment(deletCommentReq)
	if err != nil {
		var obj response.Response
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				obj = response.ClientResponse(http.StatusNotFound, st.Message(), nil)
			default:
				obj = response.ClientResponse(http.StatusInternalServerError, "Internal Server Error", nil)
			}
			c.JSON(obj.StatusCode, obj)
			return
		}
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "comment deleted succesfully", deleteCommentRes))
}

func (as *PostRelationHandler) Follow(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	followingUserIdStr := c.Param("following_user_id")
	followingUserId, err := strconv.ParseUint(followingUserIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid following user id", nil))
		return
	}
	var followRequest requestmodels.FollowRequest
	followRequest.UserID = jwtClaims.ID
	followRequest.FollowingUserID = followingUserId
	followResponse, err := as.GPPC_Client.Follow(followRequest)
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "followed user successfully", followResponse))
}

func (as *PostRelationHandler) Unfollow(c *gin.Context) {
	unfollowningUserIdStr := c.Param("unfollowing_user_id")
	unfollowningUserId, err := strconv.ParseUint(unfollowningUserIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalide unfollowing user id", nil))
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	var unfollowReq requestmodels.UnfollowRequest
	unfollowReq.UserID = jwtClaims.ID
	unfollowReq.UnfollowingUserID = unfollowningUserId
	unfollowResponse, err := as.GPPC_Client.Unfollow(unfollowReq)
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "unfollowed user successfully", unfollowResponse))
}

func (as *PostRelationHandler) FetchComments(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid post id", postId))
		return
	}
	var fetchCommentsReq requestmodels.FetchCommentsReqeust
	fetchCommentsReq.PostID = postId
	fetchCommentsResponse, err := as.DirectPostClient.Client.FetchComments(context.Background(), &post_relation.FetchCommentsRequest{
		PostId: fetchCommentsReq.PostID,
	})
	if err != nil {
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	finalComments := make([]responsemodels.Comment, 0)
	for _, v := range fetchCommentsResponse.Comments {
		var childComments []responsemodels.Comment
		if len(v.ChildComment) > 0 {
			childComments = make([]responsemodels.Comment, len(v.ChildComment))
			for i, v := range v.ChildComment {
				childComments[i] = responsemodels.Comment{
					CommentID:   v.Id,
					CommentText: v.CommentText,
					CreatedAt:   v.CreatedAt.AsTime().Local(),
					UserDetails: responsemodels.UserMetaData{
						UserID:        v.UserDetails.UserId,
						UserName:      v.UserDetails.UserName,
						Name:          v.UserDetails.Name,
						ProfileImgUrl: v.UserDetails.ProfileImgUrl,
						BlueTick:      v.UserDetails.BlueTick,
					},
					ParentCommentID: v.ParentCommentId,
				}
			}
		}
		finalComments = append(finalComments, responsemodels.Comment{
			CommentID:   v.Id,
			CommentText: v.CommentText,
			CreatedAt:   v.CreatedAt.AsTime().Local(),
			CommentAge:  v.CommentAge,
			UserDetails: responsemodels.UserMetaData{
				UserID:        v.UserDetails.UserId,
				UserName:      v.UserDetails.UserName,
				Name:          v.UserDetails.Name,
				ProfileImgUrl: v.UserDetails.ProfileImgUrl,
				BlueTick:      v.UserDetails.BlueTick,
			},
			ParentCommentID:   v.ParentCommentId,
			ChildCommentCount: v.ChildCommentCount,
			ChildComment:      childComments,
		})
	}
	//fetchCommentsResponse.Comments.CreatedAt=
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK, "comments fetched successfully", finalComments))
}
func (as *PostRelationHandler) FetchAllPosts(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		if err != nil {
			log.Printf("Error while string to int conversion(page), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page value", nil))
		return
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	offset := (page - 1) * limit
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	var req requestmodels.FetchAllPostsReq
	req.CurrentUserID = jwtClaims.ID
	targetUserIdStr:=c.Param("user_id")
	targetUserId,err:=strconv.ParseUint(targetUserIdStr,10,64)
	if err!=nil{
		c.JSON(http.StatusBadRequest,response.ClientResponse(http.StatusBadRequest,"invalid user id",nil))
	}
	req.TargetUserID=targetUserId
	req.Limit=limit
	req.Offset=offset
	authResp, err := as.DirectAuthClient.Client.GetProfileInformation(context.Background(), &auth_subscription.ProfileInfoReq{
		UserId: req.TargetUserID,
	})
	if err != nil {
		log.Println("error from grpc", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	userMetaData := responsemodels.UserMetaData{
		UserID:        authResp.UserId,
		UserName:      authResp.Username,
		Name:          authResp.Name,
		ProfileImgUrl: authResp.ProfileImageUrl,
		BlueTick:      authResp.BlueTick,
	}
	postResp, err := as.DirectPostClient.Client.FetchAllPosts(context.Background(), &post_relation.FetchAllPostsRequest{
		CurrentUserId: req.CurrentUserID,
		TargetUserId: req.TargetUserID,
		Limit: int64(req.Limit),
		Offset: int64(req.Offset),
	})
	if err != nil {
		log.Println("error from grpc", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	var finalResp []responsemodels.PostData
	for _, v := range postResp.Posts {
		var s1 []string
		for _, v1 := range v.MediaUrls {
			s1 = append(s1, v1)
		}
		finalResp = append(finalResp, responsemodels.PostData{
			PostID:        v.PostId,
			CreatedAt:     v.CreatedAt.AsTime().Local(),
			UpdatedAt:     v.UpdatedAt.AsTime().Local(),
			UserID:        v.UserId,
			Caption:       v.Caption,
			MediaUrls:     s1,
			LikeCount:     v.LikesCount,
			CommentsCount: v.CommentsCount,
			PostAge:       v.PostAge,
			IsLiked: v.IsLiked,
			UserData: userMetaData,
		})
	}
	if postResp == nil {
		c.JSON(http.StatusInternalServerError, "failed to fetch from post service")
		return
	}
	
	c.JSON(http.StatusOK, response.ClientResponse(http.StatusOK,"all posts of user fectched successfully",finalResp))
}

func (as *PostRelationHandler)FetchFollowers(c *gin.Context){
	userIdStr:=c.Param("user_id")
	userId,err:=strconv.ParseUint(userIdStr,10,64)
	if err!=nil{
		c.JSON(http.StatusBadRequest,response.ClientResponse(http.StatusBadRequest,"invalid user id",nil))
		return
	}
	resp,err:=as.DirectPostClient.Client.FetchFollowers(context.Background(),&post_relation.FetchFollowersRequest{
		UserId: userId,
	})
	if err!=nil{
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	c.JSON(http.StatusOK,resp)
}
func (as *PostRelationHandler)FetchFollowing(c *gin.Context){
	userIdStr:=c.Param("user_id")
	userId,err:=strconv.ParseUint(userIdStr,10,64)
	if err!=nil{
		c.JSON(http.StatusBadRequest,response.ClientResponse(http.StatusBadRequest,"invalid user id",nil))
		return
	}
	resp,err:=as.DirectPostClient.Client.FetchFollowing(context.Background(),&post_relation.FetchFollowingRequest{
		UserId: userId,
	})
	if err!=nil{
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	
	fmt.Println("resp in handler",resp)
	c.JSON(http.StatusOK,resp)
}

func (as *PostRelationHandler)FetchNewsFeed(c *gin.Context){
	refreshStr:=c.Query("refresh")
	lastIdStr:=c.Query("last_id")

	var req requestmodels.FetchNewsFeedRequest
	// 1. Parse LastID (The Cursor)
    lastID, _ := strconv.ParseUint(lastIdStr, 10, 64)
    req.LastID = lastID
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	req.UserID=jwtClaims.ID

	limitStr := c.Query("limit")

	

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	
	req.Limit = limit
	

	if req.LastID==0&&refreshStr=="true"{
		req.PullToRefresh=true
	} else{
		req.PullToRefresh=false
	}
	fmt.Println("req.LastID",req.LastID)
	fmt.Println("refreshStr",refreshStr)
	fmt.Println("req.PullToRefresh",req.PullToRefresh)
	resp,err:=as.DirectPostClient.Client.FetchNewsFeed(context.Background(),&post_relation.FetchNewsFeedRequest{
		UserId: req.UserID,
		Limit: int64(req.Limit),
		LastId: req.LastID,
		PullToRefresh: req.PullToRefresh,
	})
	if err!=nil{
		code, msg := utils.GRPCtoHTTP(err)
		c.JSON(code, response.ClientResponse(code, msg, nil))
		return
	}
	// c:=make([]responsemodels)
	// for _,v:=range resp{

	// }
	var finalResp []responsemodels.PostData
	for _, v := range resp.PostUserData {
		var s1 []string
		for _, v1 := range v.MediaUrls {
			s1 = append(s1, v1)
		}
		finalResp = append(finalResp, responsemodels.PostData{
			PostID:        v.PostId,
			CreatedAt:     v.CreatedAt.AsTime().Local(),
			UpdatedAt:     v.UpdatedAt.AsTime().Local(),
			UserID:        v.UserId,
			Caption:       v.Caption,
			MediaUrls:     s1,
			LikeCount:     v.LikesCount,
			CommentsCount: v.CommentsCount,
			PostAge:       v.PostAge,
			IsLiked: v.IsLiked,
			UserData: responsemodels.UserMetaData{
				UserID: v.UserMetaData.UserId,
				UserName: v.UserMetaData.UserName,
				Name: v.UserMetaData.Name,
				ProfileImgUrl: v.UserMetaData.ProfileImgUrl,
				BlueTick: v.UserMetaData.BlueTick,
			},
		})
	}

	// var lastID1 uint64
    // if len(finalResp) > 0 {
    //     // Use the ID of the last post in our mapped response
    //     lastID1 = finalResp[len(finalResp)-1].PostID
    // }
	// // Determine HasMore based on the gRPC response or slice length
    // hasMore := len(finalResp) == limit

	finallyResp:=responsemodels.FetchNewsFeedResponse{
		PostUserData: finalResp,
		NextCursor: resp.NextCursor,
        HasMore:    resp.HasMore,
	}
	c.JSON(http.StatusOK,finallyResp)
}

func (as *PostRelationHandler)FetchGlobalNewseed(c *gin.Context){
	var req requestmodels.GlobalNewsFeedRequest
	//lastIdStr:=c.Query("last_id")
	lastScoreStr:=c.Query("last_score")
	
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(authResponseModel.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "invalide claims", nil))
		return
	}
	req.UserID=jwtClaims.ID
	// 1. Parse LastID (The Cursor)
    //lastID, _ := strconv.ParseUint(lastIdStr, 10, 64)
	lastScore,_:=strconv.ParseFloat(lastScoreStr,64)
	
    req.LastScore = lastScore
	limitStr := c.Query("limit")

	

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit < 1 || limit > 100 {
		if err != nil {
			log.Printf("Error while string to int conversion(limit), error: %v", err)
		}
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid limit value, must be between 1 and 100", nil))
		return
	}

	
	req.Limit = limit
	resp,err:=as.DirectPostClient.Client.FetchGlobalNewsFeed(context.Background(),&post_relation.FetchGlobalNewsFeedRequest{
		Limit: int64(req.Limit),
		LastScore: float32(req.LastScore),
		UserId: req.UserID,
	})
	if err!=nil{

	}
	var finalResp []responsemodels.PostDataWithTrendingScore
	for _, v := range resp.PostUserData {
		var s1 []string
		for _, v1 := range v.MediaUrls {
			s1 = append(s1, v1)
		}
		finalResp = append(finalResp, responsemodels.PostDataWithTrendingScore{
			PostID:        v.PostId,
			CreatedAt:     v.CreatedAt.AsTime().Local(),
			UpdatedAt:     v.UpdatedAt.AsTime().Local(),
			UserID:        v.UserId,
			Caption:       v.Caption,
			MediaUrls:     s1,
			LikeCount:     v.LikesCount,
			CommentsCount: v.CommentsCount,
			PostAge:       v.PostAge,	
			IsLiked: v.IsLiked,
			TrendingScore: float64(v.TrendingScore),
			UserData: responsemodels.UserMetaData{
				UserID: v.UserMetaData.UserId,
				UserName: v.UserMetaData.UserName,
				Name: v.UserMetaData.Name,
				ProfileImgUrl: v.UserMetaData.ProfileImgUrl,
				BlueTick: v.UserMetaData.BlueTick,
			},
		})
	}
	c.JSON(http.StatusOK,finalResp)
}