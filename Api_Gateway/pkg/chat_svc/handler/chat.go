package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/chat_svc/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	config *config.Config
}

func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		config: cfg,
	}
}

// func (as *ChatHandler) WebSocketConnection(c *gin.Context) {
// 	claims, exists := c.Get("claims")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Claims not found", nil))
// 		return
// 	}
// 	fmt.Printf("Claims type: %T\n", claims)
// 	fmt.Println(claims)
// 	jwtClaims, ok := claims.(responsemodels.JwtClaims)
// 	if !ok {
// 		c.JSON(http.StatusUnauthorized, response.ClientResponse(http.StatusUnauthorized, "Invalid claims", nil))
// 		return
// 	}
// 	// 2️⃣ Hijack client connection
// 	hj, ok := c.Writer.(http.Hijacker)
// 	if !ok {
// 		c.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}

// 	clientConn, _, err := hj.Hijack()
// 	if err != nil {
// 		return
// 	}

// 	// 3️⃣ Connect to chat service
// 	backendConn, err := net.Dial("tcp", "localhost:50053")
// 	if err != nil {
// 		clientConn.Close()
// 		return
// 	}

// 	// 4️⃣ Clone request & inject trusted headers
// 	req := c.Request.Clone(context.Background())

// 	req.Header.Del("Authorization") // IMPORTANT: do not forward JWT

// 	req.Header.Set("X-User-ID", strconv.FormatUint(jwtClaims.ID, 10))
// 	//req.Header.Set("X-User-Role", jwtClaims.Role)
// 	//req.Header.Set("X-User-Email", jwtClaims.Email)
// 	req.Header.Set("X-Auth-Source", "gateway")

// 	// 5️⃣ Send request to chat service
// 	if err := req.Write(backendConn); err != nil {
// 		clientConn.Close()
// 		backendConn.Close()
// 		return
// 	}

// 	// 6️⃣ Start raw TCP tunnel
// 	go io.Copy(backendConn, clientConn)
// 	go io.Copy(clientConn, backendConn)
// }

func (as *ChatHandler) WebSocketConnection(c *gin.Context) {
	claims := c.MustGet("claims").(responsemodels.JwtClaims)
	requestID := c.GetString("request_id")
	target, _ := url.Parse("http://" + as.config.ChatSvcUrl)

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.Header.Del("Authorization")
		req.Header.Set("X-User-ID", strconv.FormatUint(claims.ID, 10))
		req.Header.Set("X-Auth-Source", as.config.AuthSource)
		// propagate request tracing
		if requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (as *ChatHandler) CreateGroup(c *gin.Context) {
	requestID := c.GetString("request_id")
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	var req requestmodels.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	// Inject creator ID from JWT
	req.CreatorID = claims.ID

	if !slices.Contains(req.GroupMembers, req.CreatorID) {
		req.GroupMembers = append(req.GroupMembers, req.CreatorID)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode request"})
		return
	}

	url := "http://chat_service:50053/user/group"

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Del("Authorization")
	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler)SetGroupProfileImage(c *gin.Context){
	requestID := c.GetString("request_id")
	claims := c.MustGet("claims").(responsemodels.JwtClaims)
	groupIdStr := c.Param("group_id")
	if groupIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	var req requestmodels.GroupProfileImageRequest
	file, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "Image is required"})
		return
	}
	str := as.config.ProfileImgSize
    num, err := strconv.Atoi(str) // returns (int, error)
    if err != nil {
       // fmt.Println("Error:", err)
	   c.JSON(http.StatusInternalServerError,gin.H{"error":"error in converting profile image size from sting to int"})
        return
    }
	// Check file size < 2 MB
	if file.Size > int64(num)*1024*1024 {
		c.JSON(400, gin.H{"error": "Image must be less than 2MB"})
		return
	}
	src, err := file.Open()
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "Cannot open image"})
		return
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buf := make([]byte, 512)
	_, err = src.Read(buf)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "Invalid image"})
		return
	}

	// Detect MIME type
	contentType := http.DetectContentType(buf)

	// Allowed types
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowed[contentType] {
		c.JSON(400, gin.H{"error": "Only JPG, PNG, or WebP images are allowed"})
		return
	}

	// Reset file pointer (since we read 512 bytes)
	src.Seek(0, 0)

	// Read full bytes
	data, err := io.ReadAll(src)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "Cannot read image"})
		return

	}

	req.Image=data
	req.ContentType=contentType

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode request"})
		return
	}
	url := fmt.Sprintf("http://chat_service:50053/user/group/%s/set-group-profile-image",groupIdStr)
	
	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	httpReq.Header.Del("Authorization")
	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler) AddMembers(c *gin.Context) {
	requestID := c.GetString("request_id")
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	groupIdStr := c.Param("group_id")
	if groupIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	var req requestmodels.AddMembersRequest

	req.GroupID = groupIdStr

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	req.GroupMembers = slices.DeleteFunc(req.GroupMembers, func(id uint64) bool {
		return id == claims.ID
	})
	//fmt.Println("in api gateway", req)
	jsonData, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode request"})
		return
	}

	url := "http://chat_service:50053/user/group/add-members"

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Del("Authorization")
	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler) RemoveMember(c *gin.Context) {
	requestID := c.GetString("request_id")
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	groupIdStr := c.Param("group_id")
	if groupIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	memberIdStr := c.Param("member_id")
	if memberIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member id missing"})
		return
	}
	memberId, err := strconv.ParseUint(memberIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "member is parsing error"})
	}
	var req requestmodels.RemoveMembersRequest

	req.GroupID = groupIdStr
	req.MemberID = memberId
	req.UserID = claims.ID

	// if memberId == claims.ID {
	// 	c.JSON(412, gin.H{"error": "cannot remove yourself from the group"})
	// 	return
	// }

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode request"})
		return
	}

	url := "http://chat_service:50053/user/group/remove-member"

	httpReq, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Del("Authorization")
	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler)GetGroupMembers(c *gin.Context){
	requestID := c.GetString("request_id")
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	groupIdStr := c.Param("group_id")
	if groupIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}

	// 1. Parse Pagination
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 || limit < 1 || limit > 100 {
        c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page or limit", nil))
        return
    }

	offset := (page - 1) * limit

	url := fmt.Sprintf("http://chat_service:50053/user/group/%s/members?limit=%d&offset=%d",groupIdStr,limit,offset)

	httpReq, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Del("Authorization")
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler) RecentChatProfiles(c *gin.Context) {
	requestID := c.GetString("request_id")
	// 1. Parse Pagination
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 || limit < 1 || limit > 100 {
        c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page or limit", nil))
        return
    }

	offset := (page - 1) * limit
	


	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	fullURL := fmt.Sprintf("http://chat_service:50053/user/get-recent-chat-profiles?limit=%d&offset=%d", limit, offset)
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Del("Authorization")
	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler)GetChat(c *gin.Context){
	requestID := c.GetString("request_id")
	convIdStr := c.Param("conv_id")
	if convIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	// 1. Parse Pagination
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 || limit < 1 || limit > 100 {
        c.JSON(http.StatusBadRequest, response.ClientResponse(http.StatusBadRequest, "invalid page or limit", nil))
        return
    }

	offset := (page - 1) * limit
	
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	fullURL := fmt.Sprintf("http://chat_service:50053/user/chat/%s?limit=%d&offset=%d",convIdStr, limit, offset)
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	httpReq.Header.Del("Authorization")
	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	httpReq.Header.Set("X-Auth-Source", as.config.AuthSource)
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(resp.StatusCode, "application/json", body)
}

func (as *ChatHandler) PanicInChat(c *gin.Context){
	//panic("intentional test panic in api gateway")
	fullURL := fmt.Sprintf("http://chat_service:50053/chat/panic")
	httpReq, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "chat service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(resp.StatusCode, "application/json", body)

}