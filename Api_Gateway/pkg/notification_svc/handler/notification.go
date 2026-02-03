package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/models/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
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

	target, _ := url.Parse("http://" + as.Config.NotificationSvcUrl)

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
