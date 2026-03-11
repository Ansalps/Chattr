package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	Config *config.Config
}

func NewNotificationHandler(cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{
		Config: cfg,
	}
}

func (as *NotificationHandler) WebSocketConnection(c *gin.Context) {
	claims := c.MustGet("claims").(responsemodels.JwtClaims)
	requestID := c.GetString("request_id")
	target, _ := url.Parse("http://" + as.Config.NotificationSvcUrl)

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.Header.Del("Authorization")
		req.Header.Set("X-User-ID", strconv.FormatUint(claims.ID, 10))
		req.Header.Set("X-Auth-Source", "gateway")
		// propagate request tracing
		if requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (as *NotificationHandler) GetAllNotifications(c *gin.Context) {
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

	fullURL := fmt.Sprintf("http://localhost:50054/user/notifications?limit=%d&offset=%d", limit, offset)
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("X-User-Id", strconv.FormatUint(claims.ID, 10))
	httpReq.Header.Set("X-Internal-Secret", "internalSecret")
	httpReq.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(502, gin.H{"error": "notification service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(resp.StatusCode, "application/json", body)
}
