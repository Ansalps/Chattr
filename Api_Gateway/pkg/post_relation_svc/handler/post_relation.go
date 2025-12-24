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
		c.JSON(http.StatusInternalServerError, response.ClientResponse(http.StatusInternalServerError, "error from grpc", err))
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
		c.JSON(http.StatusInternalServerError, response.ClientResponse(http.StatusInternalServerError, "error from grpc", err))
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
		c.JSON(http.StatusInternalServerError, response.ClientResponse(http.StatusInternalServerError, "error from grpc", err))
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
		c.JSON(http.StatusInternalServerError, err)
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
	req.UserID = jwtClaims.ID
	authResp, err := as.DirectAuthClient.Client.GetProfileInformation(context.Background(), &auth_subscription.ProfileInfoReq{
		UserId: req.UserID,
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
		UserId: req.UserID,
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
			UserData: userMetaData,
		})
	}
	//var mediaurls []string
	// for _,v:=range postResp.MediaUrls{
	// 	mediaurls = append(mediaurls, v)
	// }
	// postsData:=&responsemodels.FetchAllPostsResponse{
	// 	PostID: postResp.PostId,
	// 	CreatedAt: postResp.CreatedAt.AsTime().Local(),
	// 	UpdatedAt: postResp.UpdatedAt.AsTime().Local(),
	// 	UserID: postResp.UserId,
	// 	Caption: postResp.Caption,
	// 	MediaUrls: mediaurls,
	// }
	// if authResp==nil{
	// 	c.JSON(http.StatusOK,postsData)
	// 	return
	// }
	if postResp == nil {
		c.JSON(http.StatusInternalServerError, "failed to fetch from post service")
		return
	}
	//for
	c.JSON(http.StatusOK, finalResp)
}
