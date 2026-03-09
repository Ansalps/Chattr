package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/logger"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/config"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler/interfacesHandler"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase/interfacesUsecase"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	ChatUsecase   interfacesUsecase.ChatUsecase
	KafkaProducer interfacesHandler.KafkaProducer
	Config        *config.Config
	Log           logger.Logger
}

func NewChatHandler(usecase interfacesUsecase.ChatUsecase, kafkaProducer interfacesHandler.KafkaProducer, config *config.Config, log logger.Logger) *ChatHandler {
	return &ChatHandler{
		ChatUsecase:   usecase,
		KafkaProducer: kafkaProducer,
		Config:        config,
		Log:           log,
	}
}

type Client struct {
	Conn      *websocket.Conn
	UserID    uint64
	RequestID string
}
type Hub struct {
	clients    map[uint64]*Client
	register   chan *Client
	unregister chan uint64
	broadcast  chan []byte
}

var hub = &Hub{
	clients:    make(map[uint64]*Client),
	register:   make(chan *Client),
	unregister: make(chan uint64),
	broadcast:  make(chan []byte),
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c.UserID] = c
			//fmt.Println("registering", h.clients)

		case id := <-h.unregister:
			if c, ok := h.clients[id]; ok {
				c.Conn.Close()
				delete(h.clients, id)
			}

		case msg := <-h.broadcast:
			for _, c := range h.clients {
				err := c.Conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					c.Conn.Close()
					delete(h.clients, c.UserID)
				}
			}
		}
	}
}
func StartHub() {
	//fmt.Println("please do tell me if hub is starting")
	go hub.run()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// API Gateway already validated auth
		return true
	},
}

func (as *ChatHandler) WebSocketConnection(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	log := as.Log.With(
		logger.Field{Key: "request_id", Value: requestID},
	)
	// 1️⃣ Read trusted headers from API Gateway
	userIdStr := c.GetHeader("X-User-ID")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Error("failed to parse user id",
			logger.Field{Key: "user_id", Value: userIdStr},
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ailed to convert userid from string to uint64",
		})
		return
	}
	authSource := c.GetHeader("X-Auth-Source")

	//log.Println("Headers:", c.Request.Header)
	//log.Println("User ID:", userID)

	if userIdStr == "" || authSource != as.Config.AuthSource {
		log.Warn("unauthorized websocket request",
			logger.Field{Key: "user_id", Value: userIdStr},
			logger.Field{Key: "auth_source", Value: authSource},
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	log.Info("websocket upgrade request received",
		logger.Field{Key: "user_id", Value: userID},
	)
	// 2️⃣ Upgrade HTTP → WebSocket
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("websocket upgrade failed",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err},
		)
		return
	}

	// 3️⃣ Register connection (optional but recommended)
	// client,err:=as.ChatUsecase.RegisterClient(userID, wsConn)
	// if err!=nil{
	// 	log.Println("failed to register client",err)
	// }

	client := &Client{
		UserID:    userID,
		Conn:      wsConn,
		RequestID: requestID,
	}
	hub.register <- client
	log.Info("websocket client connected",
		logger.Field{Key: "user_id", Value: userID},
	)
	// 4️⃣ Start WebSocket read loop
	go as.reader(client, hub)
}
func (h *Hub) SendToGroup(dm requestmodels.MessageRequest, userIds []uint64, log logger.Logger) {
	for _, v := range userIds {
		if v == dm.SenderID {
			continue
		}
		client, ok := h.clients[v]
		if !ok {
			log.Warn("User offline:",
				logger.Field{Key: "user_id", Value: v})
			continue
		}
		data, _ := json.Marshal(dm)
		client.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (h *Hub) SendToUser(dm requestmodels.MessageRequest, log logger.Logger) {
	client, ok := h.clients[dm.RecipientID]
	if !ok {
		log.Warn("User offline:",
			logger.Field{Key: "user_id", Value: dm.RecipientID})
		return
	}
	data, _ := json.Marshal(dm)
	client.Conn.WriteMessage(websocket.TextMessage, data)
}

func (as *ChatHandler) reader(c *Client, hub *Hub) {
	log := as.Log.With(
		logger.Field{Key: "user_id", Value: c.UserID},
		logger.Field{Key: "request_id", Value: c.RequestID},
	)

	log.Info("websocket reader started")
	defer func() {
		hub.unregister <- c.UserID
		log.Info("websocket client disconnected")
	}()

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Error("failed to read websocket message",
				logger.Field{Key: "error", Value: err},
			)
			return // triggers unregister via defer
		}

		log.Debug("websocket message received")

		var dm requestmodels.MessageRequest
		if err := json.Unmarshal(p, &dm); err != nil {
			log.Warn("invalid websocket json payload: " + err.Error())
			c.Conn.WriteMessage(websocket.TextMessage, []byte("invalid json paylad"))
			continue
		}
		dm.SenderID = c.UserID

		// cancel()
		switch dm.Type {
		case "individual":
			if dm.RecipientID == 0 {
				log.Warn("invalid recipient id")
				data, _ := json.Marshal("invalid receipient id")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			exists, err := as.ChatUsecase.DoesUserExist(dm.RecipientID)
			if err != nil {
				log.Error("failed to verify recipient existence",
					logger.Field{Key: "recipient_id", Value: dm.RecipientID},
					logger.Field{Key: "error", Value: err.Error()},
				)
				data, _ := json.Marshal(err.Error())
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			if !exists {
				log.Warn("recipient does not exist",
					logger.Field{Key: "recipient_id", Value: dm.RecipientID},
				)
				data, _ := json.Marshal("invalid recipient id")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			syncTime := time.Now()

			// 1. First, Upsert the Conversation
			conversationStruct := domain.Conversation{
				ConversationID:  uuid.NewString(),
				Participants:    []uint64{dm.SenderID, dm.RecipientID},
				LastMessage:     dm.Content,
				LastMessageTime: syncTime,
				Type:            "individual",
			}

			// This now returns the ACTUAL id (either existing or the one above)
			actualConvID, err := as.ChatUsecase.StoreOrUpdateIndividualChatInConversation(conversationStruct)
			if err != nil {
				log.Error("conversation sync failed",
					logger.Field{Key: "recipient_id", Value: dm.RecipientID},
					logger.Field{Key: "error", Value: err},
				)
				// Decide if you want to stop here or continue
				data, _ := json.Marshal("failed to store individual chat in conversations")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			//fmt.Println("c.UserID", c.UserID)
			// 2. Now store the message using the actualConvID
			msgStruct := domain.Message{
				MessageID:      uuid.NewString(),
				ConversationID: actualConvID, // Now linked correctly!
				SenderID:       c.UserID,
				RecipientID:    dm.RecipientID,
				Content:        dm.Content,
				CreatedAt:      syncTime,
				Type:           "individual",
			}
			// Insert into MongoDB
			err = as.ChatUsecase.StoreIndividualChatInMessages(msgStruct)
			if err != nil {
				log.Error("failed to store individual message",
					logger.Field{Key: "conversation_id", Value: actualConvID},
					logger.Field{Key: "error", Value: err},
				)
				data, _ := json.Marshal("failed to store individual chat in messages")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			log.Info("direct message stored",
				logger.Field{Key: "conversation_id", Value: actualConvID},
				logger.Field{Key: "recipient_id", Value: dm.RecipientID},
			)
			hub.SendToUser(dm, log)
			event := map[string]interface{}{
				"type":           "DIRECT_MESSAGE",
				"actorId":        c.UserID,       // Person who clicked 'like'
				"recipientId":    dm.RecipientID, // Person receiving the notification
				"conversationId": actualConvID,   // The content being liked
				"timestamp":      time.Now().Unix(),
			}
			// Convert to JSON and publish to topic "post-events"
			err = as.KafkaProducer.PublishEvent("chat-events", event)
			if err != nil {
				// Log the error but don't necessarily fail the request
				// unless real-time notification is critical.
				log.Error("failed to publish kafka event",
					logger.Field{Key: "topic", Value: "chat-events"},
					logger.Field{Key: "error", Value: err},
				)
			}
		case "group":
			if dm.GroupID == "" {
				log.Warn("invalid group id")
				data, _ := json.Marshal("invalid group id")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			userIds, err := as.ChatUsecase.FetchMembersOfGroup(dm.GroupID)
			if err != nil {
				log.Error("failed to fetch group members",
					logger.Field{Key: "group_id", Value: dm.GroupID},
					logger.Field{Key: "error", Value: err},
				)

				data, _ := json.Marshal("fetching group members failed or group id not found")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			if !slices.Contains(userIds, c.UserID) {
				log.Warn("sender not part of group",
					logger.Field{Key: "group_id", Value: dm.GroupID},
				)
				data, _ := json.Marshal("can't send message because sender is not a group member")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			syncTime := time.Now()

			// 1. FIRST: Handle the Conversation
			conversationStruct := domain.Conversation{
				ConversationID:  uuid.NewString(),
				Participants:    userIds,
				GroupID:         dm.GroupID,
				LastMessage:     dm.Content,
				LastMessageTime: syncTime,
				Type:            "group",
			}

			// Get the ID that MongoDB is actually using
			actualConvID, err := as.ChatUsecase.StoreOrUpdateGroupChatInConversation(conversationStruct)
			if err != nil {
				log.Error("failed to sync group conversation",
					logger.Field{Key: "group_id", Value: dm.GroupID},
					logger.Field{Key: "error", Value: err},
				)
				data, _ := json.Marshal("failed to store group chat in conversation")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}

			// 2. SECOND: Store the Message with the correct ConversationID
			msgStruct := domain.Message{
				MessageID:      uuid.NewString(),
				ConversationID: actualConvID, // Linked!
				SenderID:       dm.SenderID,
				GroupID:        dm.GroupID,
				Content:        dm.Content,
				CreatedAt:      syncTime,
				Type:           "group",
			}
			err = as.ChatUsecase.StoreGroupChatInMessages(msgStruct)
			if err != nil {
				log.Error("failed to store group message",
					logger.Field{Key: "conversation_id", Value: actualConvID},
					logger.Field{Key: "error", Value: err},
				)
				data, _ := json.Marshal("failed to store group chat in messages")
				c.Conn.WriteMessage(websocket.TextMessage, data)
			}
			log.Info("group message stored",
				logger.Field{Key: "conversation_id", Value: actualConvID},
				logger.Field{Key: "group_id", Value: dm.GroupID},
			)
			hub.SendToGroup(dm, userIds, log)
			log.Debug("group message broadcasted",
				logger.Field{Key: "group_id", Value: dm.GroupID},
				logger.Field{Key: "recipients_count", Value: len(userIds)},
			)
			groupName, err := as.ChatUsecase.GetGroupName(dm.GroupID)
			if err != nil {
				log.Error("failed to fetch group name",
					logger.Field{Key: "group_id", Value: dm.GroupID},
					logger.Field{Key: "error", Value: err},
				)
				data, _ := json.Marshal("failed to fetch group name")
				c.Conn.WriteMessage(websocket.TextMessage, data)
				break
			}
			event := map[string]interface{}{
				"type":           "GROUP_MESSAGE",
				"actorId":        c.UserID, // Person who clicked 'like'
				"recipientId":    userIds,  // Person receiving the notification
				"groupName":      groupName,
				"conversationId": actualConvID, // The content being liked
				"timestamp":      time.Now().Unix(),
			}
			// Convert to JSON and publish to topic "post-events"
			err = as.KafkaProducer.PublishEvent("chat-events", event)
			if err != nil {
				// Log the error but don't necessarily fail the request
				// unless real-time notification is critical.
				log.Error("failed to publish kafka event",
					logger.Field{Key: "topic", Value: "chat-events"},
					logger.Field{Key: "conversation_id", Value: actualConvID},
					logger.Field{Key: "group_id", Value: dm.GroupID},
					logger.Field{Key: "error", Value: err},
				)
			} else {

				log.Info("group message event published",
					logger.Field{Key: "topic", Value: "chat-events"},
					logger.Field{Key: "conversation_id", Value: actualConvID},
				)
			}
		default:
			log.Warn("invalid message type received",
				logger.Field{Key: "message_type", Value: dm.Type},
			)
			resp := map[string]string{
				"type":    "ERROR",
				"message": "invalid message type",
			}
			data, err := json.Marshal(resp)
			if err != nil {
				log.Error("failed to marshal websocket error response",
					logger.Field{Key: "error", Value: err},
				)
				break
			}
			err = c.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Error("failed to write websocket error message",
					logger.Field{Key: "error", Value: err},
				)
			}
		}
	}
}

func (as *ChatHandler) CreateGroup(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	log := as.Log.With(
		logger.Field{Key: "request_id", Value: requestID},
	)
	var req requestmodels.CreateGroupRequest

	creatorIdStr := c.GetHeader("X-User-Id")
	if creatorIdStr == "" {
		log.Warn("Empty string fetched as userid from header in chat service")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no access",
		})
		return
	}
	CreatorID, err := strconv.ParseUint(creatorIdStr, 10, 64)
	if err != nil {
		log.Error("failed to parse user id",
			logger.Field{Key: "user_id", Value: creatorIdStr},
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}
	authSource := c.GetHeader("X-Auth-Source")
	if authSource != as.Config.AuthSource {
		log.Warn("unauthorized websocket request",
			logger.Field{Key: "user_id", Value: creatorIdStr},
			logger.Field{Key: "auth_source", Value: authSource},
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	// 2. Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("error binding request:",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// 3. Set Metadata
	req.CreatorID = CreatorID
	groupId := uuid.New().String()
	req.GroupID = groupId
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	resp, err := as.ChatUsecase.CreateGroup(req)

	if err != nil {
		// switch{
		// case errors.As(err, &userErr):
		// 	log.Warn("Invalid userids persent",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error":            "some users do not exist",
		// 		"missing_user_ids": userErr.UserIDs,
		// 	})
		// case errors.Is(err, domain.ErrGroupNotFound):
		// 	log.Warn("Invalid group id",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(http.StatusNotFound, utils.ClientResponse(404, domain.ErrGroupNotFound.Error(), nil))
		// case errors.Is(err, domain.ErrNotGroupMember):
		// 	log.Warn("Not a Group Member",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(http.StatusForbidden, utils.ClientResponse(403, domain.ErrNotGroupMember.Error(), nil))
		// case errors.Is(err, domain.ErrDatabaseTimeout):
		// 	log.Error("Database Connection timed out",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(500, utils.ClientResponse(500, domain.ErrDatabaseTimeout.Error(), nil))
		// case errors.Is(err, domain.ErrInternal):
		// 	log.Error("Internal server error",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(500, utils.ClientResponse(500, domain.ErrInternal.Error(), nil))
		// default:
		// 	log.Error("internal server errot",
		// 		logger.Field{Key: "error", Value: err})
		// 	// ✅ fallback (VERY IMPORTANT)
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": "internal server error",
		// 	})
		// }
		utils.MapDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, resp)

}
func (as *ChatHandler) AddMembers(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	log := as.Log.With(
		logger.Field{Key: "request_id", Value: requestID},
	)
	var req requestmodels.AddMembersRequest

	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Warn("Empty string fetched as userid from header in chat service")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no access",
		})
		return
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Error("failed to parse user id",
			logger.Field{Key: "user_id", Value: userIdStr},
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error in parsing",
		})
	}
	authSource := c.GetHeader("X-Auth-Source")
	if authSource != as.Config.AuthSource {
		log.Warn("unauthorized websocket request",
			logger.Field{Key: "user_id", Value: userIdStr},
			logger.Field{Key: "auth_source", Value: authSource},
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	req.UserID = userID
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("error binding request:",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	//var userErr *domain.NonExistingUsersError
	resp, err := as.ChatUsecase.AddMembers(req)
	if err != nil {

		// if errors.As(err, &userErr) {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error":            "some users do not exist",
		// 		"missing_user_ids": userErr.UserIDs,
		// 	})
		// 	return
		// }

		// switch {
		// case errors.As(err, &userErr):
		// 	log.Warn("Invalid userids persent",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error":            "some users do not exist",
		// 		"missing_user_ids": userErr.UserIDs,
		// 	})
		// case errors.Is(err, domain.ErrGroupNotFound):
		// 	log.Warn("Invalid group id",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(http.StatusNotFound, utils.ClientResponse(404, domain.ErrGroupNotFound.Error(), nil))
		// case errors.Is(err, domain.ErrNotGroupMember):
		// 	log.Warn("Not a Group Member",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(http.StatusForbidden, utils.ClientResponse(403, domain.ErrNotGroupMember.Error(), nil))
		// case errors.Is(err, domain.ErrDatabaseTimeout):
		// 	log.Error("Database Connection timed out",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(500, utils.ClientResponse(500, domain.ErrDatabaseTimeout.Error(), nil))
		// case errors.Is(err, domain.ErrInternal):
		// 	log.Error("Internal server error",
		// 		logger.Field{Key: "error", Value: err})
		// 	c.JSON(500, utils.ClientResponse(500, domain.ErrInternal.Error(), nil))
		// default:
		// 	log.Error("internal server errot",
		// 		logger.Field{Key: "error", Value: err})
		// 	// ✅ fallback (VERY IMPORTANT)
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": "internal server error",
		// 	})
		// }
		utils.MapDomainError(c, log, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (as *ChatHandler) RemoveMember(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	log := as.Log.With(
		logger.Field{Key: "request_id", Value: requestID},
	)
	var req requestmodels.RemoveMemberRequest

	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Warn("Empty string fetched as userid from header in chat service")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no access",
		})
		return
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Error("failed to parse user id",
			logger.Field{Key: "user_id", Value: userIdStr},
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error in parsing string userid to uint",
		})
		return
	}
	authSource := c.GetHeader("X-Auth-Source")
	if authSource != as.Config.AuthSource {
		log.Warn("unauthorized websocket request",
			logger.Field{Key: "user_id", Value: userIdStr},
			logger.Field{Key: "auth_source", Value: authSource},
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	req.UserID = userID
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("error binding request:",
			logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := as.ChatUsecase.RemoveMember(req)
	if err != nil {
		//log.Println(err)
		// switch err {
		// case domain.ErrGroupNotFound:
		// 	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		// case domain.ErrNotGroupMember:
		// 	c.JSON(403, gin.H{"error": err.Error()})
		// case domain.ErrUserNotPresent:
		// 	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		// default:
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// }

		//c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		utils.MapDomainError(c,log,err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (as *ChatHandler) GetGroupMembers(c *gin.Context) {
	groupIdStr := c.Param("group_id")
	if groupIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	// 1. Parse Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page or limit",
		})
		//c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	offset := (page - 1) * limit
	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Println("error fetching userid from header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch userid from header"})
		return
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("error parsing userid to uint64")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error converting userid from string to uint"})
		return
	}
	authSource := c.GetHeader("X-Auth-Source")

	//log.Println("Headers:", c.Request.Header)
	//log.Println("User ID:", userID)

	if userIdStr == "" || authSource != as.Config.AuthSource {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	var req requestmodels.GetGroupMembersRequest
	req.GroupID = groupIdStr
	req.UserID = userID
	req.Limit = limit
	req.Offset = offset
	//var resp []responsemodels.GetGroupMembersResponse
	resp, err := as.ChatUsecase.GetGroupMembers(req)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	resp1 := responsemodels.GetGroupMembers{
		GetGroupMembers: resp,
		Pagination: responsemodels.PaginationDetails{
			CurrentPage: page,
			PageSize:    limit,
		},
	}
	c.JSON(http.StatusOK, resp1)
}

func (as *ChatHandler) GetRecentChatProfiles(c *gin.Context) {
	// 1. Parse Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page or limit",
		})
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	offset := (page - 1) * limit
	var req requestmodels.RecentChatProfilesRequest
	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Println("error fetching userid from header")
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("error parsing userid to uint64")
	}
	authSource := c.GetHeader("X-Auth-Source")

	//log.Println("Headers:", c.Request.Header)
	//log.Println("User ID:", userID)

	if userIdStr == "" || authSource != as.Config.AuthSource {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized  request",
		})
		return
	}
	req.UserID = userID
	req.Limit = limit
	req.Offset = offset
	//var resp []responsemodels.ChatProfileResponse
	resp, err := as.ChatUsecase.GetRecentChatProfiles(req)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	resp1 := responsemodels.ChatProfileFinalResponse{
		ChatProfiles: resp,
		Pagination: responsemodels.PaginationDetails{
			CurrentPage: page,
			PageSize:    limit,
		},
	}
	c.JSON(http.StatusOK, resp1)
}

func (as *ChatHandler) GetChat(c *gin.Context) {
	convIdStr := c.Param("conv_id")
	if convIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	// 1. Parse Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page or limit",
		})
		//c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	offset := (page - 1) * limit
	var req requestmodels.GetChatRequest
	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Println("error fetching userid from header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch userid from header"})
		return
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("error parsing userid to uint64")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error converting userid from string to uint"})
		return
	}
	authSource := c.GetHeader("X-Auth-Source")

	//log.Println("Headers:", c.Request.Header)
	//log.Println("User ID:", userID)

	if userIdStr == "" || authSource != as.Config.AuthSource {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	req.ConvID = convIdStr
	req.UserID = userID
	req.Limit = limit
	req.Offset = offset
	//var resp responsemodels.GetChatResponse
	resp, err := as.ChatUsecase.GetChat(req)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	resp.Pagination = responsemodels.PaginationDetails{
		CurrentPage: page,
		PageSize:    limit,
	}
	c.JSON(http.StatusOK, resp)
}
func (as *ChatHandler) SetGroupProfileImage(c *gin.Context) {
	groupIdStr := c.Param("group_id")
	if groupIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group id missing"})
		return
	}
	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Println("error fetching userid from header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch userid from header"})
		return
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("error parsing userid to uint64")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error converting userid from string to uint"})
		return
	}
	authSource := c.GetHeader("X-Auth-Source")

	if authSource != as.Config.AuthSource {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}
	var req requestmodels.GroupProfileImageRequest
	req.UserID = userID
	req.GroupID = groupIdStr
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("error binding request in chat service", err)
		return
	}
	//fmt.Println("req",req)
	resp, err := as.ChatUsecase.SetGroupProfileImage(req)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"group_profile_image_url": resp})
}
