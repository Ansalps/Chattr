package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/chat_svc/requestmodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
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

	target, _ := url.Parse("http://" + as.config.ChatSvcUrl)

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.Header.Del("Authorization")
		req.Header.Set("X-User-ID", strconv.FormatUint(claims.ID, 10))
		req.Header.Set("X-Auth-Source", "gateway")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (as *ChatHandler) CreateGroup(c *gin.Context) {
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

	url := "http://localhost:50053/user/group"

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	//httpReq.Header.Set("X-User-Role", claims.Role)
	//httpReq.Header.Set("X-User-Email", claims.Email)
	httpReq.Header.Set("X-Internal-Secret", "internalSecret")

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
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	groupIdStr:=c.Param("group_id")

	var req requestmodels.AddMembersRequest

	req.GroupID=groupIdStr
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}


	req.GroupMembers = slices.DeleteFunc(req.GroupMembers, func(id uint64) bool {
		return id == claims.ID
	})

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode request"})
		return
	}

	url := "http://localhost:50053/user/group/add-members"

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	//httpReq.Header.Set("X-User-Role", claims.Role)
	//httpReq.Header.Set("X-User-Email", claims.Email)
	httpReq.Header.Set("X-Internal-Secret", "internalSecret")

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

func (as *ChatHandler)RemoveMember(c *gin.Context){
	claims := c.MustGet("claims").(responsemodels.JwtClaims)

	groupIdStr:=c.Param("group_id")

	var req requestmodels.RemoveMembersRequest

	req.GroupID=groupIdStr
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if req.MemberID==claims.ID{
		c.JSON(412,gin.H{"error":"cannot remove yourself from the group"})
		return
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode request"})
		return
	}

	url := "http://localhost:50053/user/group/remove-member"

	httpReq, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Identity headers
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	//httpReq.Header.Set("X-User-Role", claims.Role)
	//httpReq.Header.Set("X-User-Email", claims.Email)
	httpReq.Header.Set("X-Internal-Secret", "internalSecret")

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