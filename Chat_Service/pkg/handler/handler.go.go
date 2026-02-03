package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/handler/interfacesHandler"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/usecase/interfacesUsecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	ChatUsecase interfacesUsecase.ChatUsecase
	KafkaProducer         interfacesHandler.KafkaProducer
}

func NewChatHandler(usecase interfacesUsecase.ChatUsecase,	kafkaProducer interfacesHandler.KafkaProducer) *ChatHandler {
	return &ChatHandler{
		ChatUsecase: usecase,
		KafkaProducer: kafkaProducer,
	}
}

type Client struct {
	Conn   *websocket.Conn
	UserID uint64
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
			fmt.Println("registering", h.clients)

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

	client := &Client{
		UserID: userID,
		Conn:   wsConn,
	}
	hub.register <- client
	// 4️⃣ Start WebSocket read loop
	go as.reader(client, hub)
}
func (h *Hub) SendToGroup(dm requestmodels.MessageRequest, userIds []uint64) {
	fmt.Println("map", h.clients)
	fmt.Println("userIds", userIds)
	for _, v := range userIds {
		if v == dm.SenderID {
			continue
		}
		client, ok := h.clients[v]
		if !ok {
			log.Println("User offline:",v)
			return
		}
		data, _ := json.Marshal(dm)
		client.Conn.WriteMessage(websocket.TextMessage, data)
	}

	// payload := map[string]interface{}{
	// 	"type":    "individual",
	// 	"from":    from,
	// 	"message": message,
	// }

}
func (h *Hub) SendToUser(dm requestmodels.MessageRequest) {
	fmt.Println("map", h.clients)
	client, ok := h.clients[dm.RecipientID]
	if !ok {
		log.Println("User offline:", dm.RecipientID)
		return
	}

	// payload := map[string]interface{}{
	// 	"type":    "individual",
	// 	"from":    from,
	// 	"message": message,
	// }

	data, _ := json.Marshal(dm)
	client.Conn.WriteMessage(websocket.TextMessage, data)
}

func (as *ChatHandler) reader(c *Client, hub *Hub) {
	defer func() {
		hub.unregister <- c.UserID
	}()

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("error while reading", err)
			return // triggers unregister via defer
		}
		
		var dm requestmodels.MessageRequest
		if err := json.Unmarshal(p, &dm); err != nil {
			log.Println("Invalid JSON")
			c.Conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			continue
		}
		

		// cancel()
		switch dm.Type {
		case "individual":
			if dm.RecipientID == 0 {
				log.Println("invalid recipient id")
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
				log.Println("Conversation handling failed", err)
				// Decide if you want to stop here or continue
			}

			// 2. Now store the message using the actualConvID
			msgStruct := domain.Message{
				MessageID:      uuid.NewString(),
				ConversationID: actualConvID, // Now linked correctly!
				SenderID:       dm.SenderID,
				RecipientID:    dm.RecipientID,
				Content:        dm.Content,
				CreatedAt:      syncTime,
				Type:           "individual",
			}
			// Insert into MongoDB
			err = as.ChatUsecase.StoreIndividualChatInMessages(msgStruct)
			if err != nil {
				log.Println("Store individual chat in messages collection failed")
			}

			hub.SendToUser(dm)
			event := map[string]interface{}{
				"type":          "DIRECT_MESSAGE",
				"actorId":       dm.SenderID,     // Person who clicked 'like'
				"recipientId":   dm.RecipientID,    // Person receiving the notification
				"conversationId": actualConvID,     // The content being liked
				"timestamp":     time.Now().Unix(),
			}
			// Convert to JSON and publish to topic "post-events"
			err = as.KafkaProducer.PublishEvent("chat-events", event)
			if err != nil {
				// Log the error but don't necessarily fail the request 
				// unless real-time notification is critical.
				log.Printf("Failed to emit Kafka event: %v", err)
			}
		case "group":
			if dm.GroupID == "" {
				log.Println("invalid group id")
				break
			}
			userIds, err := as.ChatUsecase.FetchMembersOfGroup(dm.GroupID)
			if err != nil {
				log.Println("can't send message to group failed to fetch user ids", err)
				break
			}
			if !slices.Contains(userIds, dm.SenderID) {
				log.Println("can't send message because sender is not a group member")
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
				log.Println("failed to sync conversation:", err)
			}

			// 2. SECOND: Store the Message with the correct ConversationID
			msgStruct := domain.Message{
				MessageID: uuid.NewString(),
				ConversationID: actualConvID, // Linked!
				SenderID:  dm.SenderID,
				GroupID:   dm.GroupID,
				Content:   dm.Content,
				CreatedAt: syncTime,
				Type:      "group",
			}
			err = as.ChatUsecase.StoreGroupChatInMessages(msgStruct)
			if err != nil {
				log.Println("store group chat in messages failed", err)
			}
			
			hub.SendToGroup(dm, userIds)

			groupName,err:=as.ChatUsecase.GetGroupName(dm.GroupID)
			if err!=nil{
				log.Println("failed to get group name by grouop id")
			}
			event := map[string]interface{}{
				"type":          "GROUP_MESSAGE",
				"actorId":       dm.SenderID,     // Person who clicked 'like'
				"recipientId":   userIds,    // Person receiving the notification
				"groupName":groupName,
				"conversationId": actualConvID,     // The content being liked
				"timestamp":     time.Now().Unix(),
			}
			// Convert to JSON and publish to topic "post-events"
			err = as.KafkaProducer.PublishEvent("chat-events", event)
			if err != nil {
				// Log the error but don't necessarily fail the request 
				// unless real-time notification is critical.
				log.Printf("Failed to emit Kafka event: %v", err)
			}
		default:
			log.Println("invalid message type")
		}
		// if dm.Type=="individual"{
		// 	hub.SendToUser(dm)
		// } else if dm.Type=="grou"{
		// 	hub.SendToGroup(dm)
		// }

		// err = c.Conn.WriteMessage(msgType, msg)
		// if err != nil {
		// 	break
		// }
	}
}

func (as *ChatHandler) CreateGroup(c *gin.Context) {
	var req requestmodels.CreateGroupRequest

	creatorIdStr := c.GetHeader("X-User-Id")
	CreatorID,err := strconv.ParseUint(creatorIdStr, 10, 64)
	if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
        return
    }
	// 2. Bind JSON body
    if err := c.ShouldBindJSON(&req); err != nil {
        log.Println("error binding request:", err)
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
		log.Println(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
func (as *ChatHandler) AddMembers(c *gin.Context) {
	fmt.Println("why is it not reaching in AddMembers Handler?")
	var req requestmodels.AddMembersRequest

	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no access",
		})
		return
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error in parsing",
		})
	}

	req.UserID = userID
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("error binding request in chat service", err)
		return
	}
	fmt.Println("req in handler", req)
	resp, err := as.ChatUsecase.AddMembers(req)
	if err != nil {
		log.Println(err)
		if err == domain.ErrNotGroupMember {

			c.JSON(403, gin.H{"error": err.Error()})
			return
		}
		return
	}
	fmt.Println("resp", resp)
	c.JSON(http.StatusOK, resp)
}

func (as *ChatHandler) RemoveMember(c *gin.Context) {
	var req requestmodels.RemoveMemberRequest

	userIdStr := c.GetHeader("X-User_Id")
	userID, _ := strconv.ParseUint(userIdStr, 10, 64)
	req.UserID = userID
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("error binding request in chat service", err)
		return
	}

	resp, err := as.ChatUsecase.RemoveMember(req)
	if err != nil {
		log.Println(err)
		if err == domain.ErrUserNotPresent {

			c.JSON(403, gin.H{"error": err.Error()})
			return
		}
		return
	}
	c.JSON(http.StatusOK, resp)
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
	req.UserID = userID
	req.Limit = limit
	req.Offset = offset
	resp, err := as.ChatUsecase.GetRecentChatProfiles(req)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (as *ChatHandler)GetChat(c *gin.Context){
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
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	offset := (page - 1) * limit
	var req requestmodels.GetChatRequest
	userIdStr := c.GetHeader("X-User-Id")
	if userIdStr == "" {
		log.Println("error fetching userid from header")
	}
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		log.Println("error parsing userid to uint64")
	}
	req.ConvID=convIdStr
	req.UserID = userID
	req.Limit = limit
	req.Offset = offset
	resp, err := as.ChatUsecase.GetChat(req)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}