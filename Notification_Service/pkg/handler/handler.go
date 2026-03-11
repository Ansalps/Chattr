package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Ansalps/Chattr_Notification_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/logger"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/infrastructure/websockethub"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Notification_Service/pkg/usecase/interfacesUsecase"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type NotificationHandler struct {
	NotificationUsecase interfacesUsecase.NotificationUsecase
	Hub                 *websockethub.Hub
	Log logger.Logger
}

func NewNotificationHandler(usecase interfacesUsecase.NotificationUsecase, hub *websockethub.Hub,log logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		NotificationUsecase: usecase,
		Hub:                 hub,
		Log: log,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// API Gateway already validated auth
		return true
	},
}

func (as *NotificationHandler) WebSocketConnection(c *gin.Context) {
	// 1️⃣ Read trusted headers from API Gateway
	userIdStr := c.GetHeader("X-User-ID")
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("failed to convert userid from string to uint", err)
	}
	authSource := c.GetHeader("X-Auth-Source")

	log.Println("Headers:", c.Request.Header)
	log.Println("User ID:", userID)

	if userIdStr == "" || authSource != "gateway" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized websocket request",
		})
		return
	}

	// 2️⃣ Upgrade HTTP → WebSocket
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 3️⃣ Register connection (optional but recommended)
	// client,err:=as.ChatUsecase.RegisterClient(userID, wsConn)
	// if err!=nil{
	// 	log.Println("failed to register client",err)
	// }

	client := &websockethub.Client{
		UserID: userID,
		Conn:   wsConn,
	}
	as.Hub.Register <- client
	// 4️⃣ Start WebSocket read loop
	//go as.reader(client)
	go as.ConsumeWebsocket(client)
}

func (as *NotificationHandler) ConsumeWebsocket(client *websockethub.Client) {
	// 3️⃣ DEFER Unregister: This runs when this function exits
	defer func() {
		as.Hub.Unregister <- client.UserID
	}()

	for {
		// We read from the connection. If the user disconnects,
		// ReadMessage will return an error, breaking the loop.
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("User %d disconnected: %v", client.UserID, err)
			break
		}
	}
}

func (as *NotificationHandler) GetAllNotifications(c *gin.Context) {

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

	var req requestmodels.GetNotificationsequest
	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Println("error fetching userid from header")
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("error parsing userid to uint64")
	}
	req.UserID = userID
	req.Limit = limit
	req.Offset = offset
	//var resp []domain.Notification
	resp, err := as.NotificationUsecase.GetAllNotifications(req)
	if err != nil {

		log.Println(err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	resp1 := domain.NotificationResponse{
		Notifications: resp,
		Pagination: domain.PaginationDetails{
			CurrentPage: page,
			PageSize:    limit,
		},
	}
	c.JSON(http.StatusOK, resp1)
}

// func (as *NotificationHandler) reader(c *websockethub.Client) {
// 	defer func() {
// 		as.Hub.Unregister <- c.UserID
// 	}()

// 	for {
// 		_, p, err := c.Conn.ReadMessage()
// 		if err != nil {
// 			log.Println("error while reading", err)
// 			return // triggers unregister via defer
// 		}

// 		var dm requestmodels.MessageRequest
// 		if err := json.Unmarshal(p, &dm); err != nil {
// 			log.Println("Invalid JSON")
// 			c.Conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
// 			continue
// 		}

// 		// cancel()
// 		switch dm.Type {
// 		case "individual":
// 			if dm.RecipientID == 0 {
// 				log.Println("invalid recipient id")
// 				break
// 			}
// 			syncTime := time.Now()

// 			// 1. First, Upsert the Conversation
// 			conversationStruct := domain.Conversation{
// 				ConversationID:  uuid.NewString(),
// 				Participants:    []uint64{dm.SenderID, dm.RecipientID},
// 				LastMessage:     dm.Content,
// 				LastMessageTime: syncTime,
// 				Type:            "individual",
// 			}

// 			// This now returns the ACTUAL id (either existing or the one above)
// 			actualConvID, err := as.ChatUsecase.StoreOrUpdateIndividualChatInConversation(conversationStruct)
// 			if err != nil {
// 				log.Println("Conversation handling failed", err)
// 				// Decide if you want to stop here or continue
// 			}

// 			// 2. Now store the message using the actualConvID
// 			msgStruct := domain.Message{
// 				MessageID:      uuid.NewString(),
// 				ConversationID: actualConvID, // Now linked correctly!
// 				SenderID:       dm.SenderID,
// 				RecipientID:    dm.RecipientID,
// 				Content:        dm.Content,
// 				CreatedAt:      syncTime,
// 				Type:           "individual",
// 			}
// 			// Insert into MongoDB
// 			err = as.ChatUsecase.StoreIndividualChatInMessages(msgStruct)
// 			if err != nil {
// 				log.Println("Store individual chat in messages collection failed")
// 			}

// 			hub.SendToUser(dm)
// 		case "group":
// 			if dm.GroupID == "" {
// 				log.Println("invalid group id")
// 				break
// 			}
// 			userIds, err := as.ChatUsecase.FetchMembersOfGroup(dm.GroupID)
// 			if err != nil {
// 				log.Println("can't send message to group failed to fetch user ids", err)
// 				break
// 			}
// 			if !slices.Contains(userIds, dm.SenderID) {
// 				log.Println("can't send message because sender is not a group member")
// 				break
// 			}
// 			syncTime := time.Now()

// 			// 1. FIRST: Handle the Conversation
// 			conversationStruct := domain.Conversation{
// 				ConversationID:  uuid.NewString(),
// 				Participants:    userIds,
// 				GroupID:         dm.GroupID,
// 				LastMessage:     dm.Content,
// 				LastMessageTime: syncTime,
// 				Type:            "group",
// 			}

// 			// Get the ID that MongoDB is actually using
// 			actualConvID, err := as.ChatUsecase.StoreOrUpdateGroupChatInConversation(conversationStruct)
// 			if err != nil {
// 				log.Println("failed to sync conversation:", err)
// 			}

// 			// 2. SECOND: Store the Message with the correct ConversationID
// 			msgStruct := domain.Message{
// 				MessageID: uuid.NewString(),
// 				ConversationID: actualConvID, // Linked!
// 				SenderID:  dm.SenderID,
// 				GroupID:   dm.GroupID,
// 				Content:   dm.Content,
// 				CreatedAt: syncTime,
// 				Type:      "group",
// 			}
// 			err = as.ChatUsecase.StoreGroupChatInMessages(msgStruct)
// 			if err != nil {
// 				log.Println("store group chat in messages failed", err)
// 			}

// 			hub.SendToGroup(dm, userIds)
// 		default:
// 			log.Println("invalid message type")
// 		}
// 		// if dm.Type=="individual"{
// 		// 	hub.SendToUser(dm)
// 		// } else if dm.Type=="grou"{
// 		// 	hub.SendToGroup(dm)
// 		// }

// 		// err = c.Conn.WriteMessage(msgType, msg)
// 		// if err != nil {
// 		// 	break
// 		// }
// 	}
// }
