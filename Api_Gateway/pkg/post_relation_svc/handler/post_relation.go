package handler

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/client/interfaces"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type PostRelationHandler struct {
	GPPC_Client interfaces.PostRelationClient
	config      *config.Config
}

func NewPostRelationHandler(postRelationClient interfaces.PostRelationClient, cfg *config.Config) *PostRelationHandler {
	return &PostRelationHandler{
		GPPC_Client: postRelationClient,
		config:      cfg,
	}
}

func (as *PostRelationHandler) CreatePost(c *gin.Context) {
	var createPostReq requestmodels.CreatePostRequest
	if err := c.ShouldBindJSON(&createPostReq); err != nil {
		if validationErrors := utils.FormatValidationError(err); validationErrors != nil {
			c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Validation failed", validationErrors))
			return
		}
		log.Printf("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "Invalid request body", nil))
		return
	}

	claims, exists := c.Get("Claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
		return
	}
	jwtClaims, ok := claims.(responsemodels.JwtClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
		return
	}
	createPostReq.UserID=jwtClaims.ID
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
	createPostResponse,err:=as.GPPC_Client.CreatePost(createPostReq)
	if err!=nil{

	}


	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"urls":    uploadedUrls,
		"res":createPostResponse,
	})
}
