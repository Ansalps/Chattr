package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Ansalps/Chattr_Chat_Service/pkg/domain"
	interfacesrepository "github.com/Ansalps/Chattr_Chat_Service/pkg/repository/interfacesRepository"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/requestmodels"
	"github.com/Ansalps/Chattr_Chat_Service/pkg/responsemodels"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatRepository struct {
	MongoClient *mongo.Client
}

func NewChatRepository(mongoClient *mongo.Client) interfacesrepository.ChatRepository {
	return &ChatRepository{
		MongoClient: mongoClient,
	}
}

// Helper to get the group collection
func (ad *ChatRepository) groupColl() *mongo.Collection {
	return ad.MongoClient.Database("chat").Collection("group")
}

// Helper to get the conversation collection
func (ad *ChatRepository) convColl() *mongo.Collection {
	return ad.MongoClient.Database("chat").Collection("conversations")
}

// func (ad *ChatRepository) CreateGroup(req requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error) {
// 	groupCollection := ad.MongoClient.Database("chat").Collection("group")

// 	// Insert the request object directly
// 	// Note: Ensure your struct has `bson` tags (see below)
// 	_, err := groupCollection.InsertOne(context.TODO(), req)
// 	if err != nil {
// 		log.Println("Error inserting group into MongoDB:", err)
// 		return responsemodels.CreateGroupResponse{}, err
// 	}

//		return responsemodels.CreateGroupResponse{
//			GroupID: req.GroupID,
//		}, nil
//	}
//
// GroupExists checks if a group exists by its string groupid
func (r *ChatRepository) GroupExists(ctx context.Context, groupID string) (bool, error) {
	// We only need to know if at least one document matches
	filter := bson.M{"groupid": groupID}

	// Using CountDocuments is idiomatic and efficient
	count, err := r.groupColl().CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (ad *ChatRepository) CreateGroup(req requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error) {
	groupCollection := ad.MongoClient.Database("chat").Collection("group")

	// Use a background context with a specific timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a, err := groupCollection.InsertOne(ctx, req)
	if err != nil {
		return responsemodels.CreateGroupResponse{}, err
	}
	fmt.Println("a", a)
	return responsemodels.CreateGroupResponse{
		GroupID: req.GroupID,
	}, nil
}

func (ad *ChatRepository) CreatorID(groupId string) (uint64, error) {
	//fmt.Println("goupId", groupId)
	var result struct {
		CreatorID uint64 `bson:"creatorid"`
	}

	filter := bson.M{"groupid": groupId}
	opts := options.FindOne().SetProjection(bson.M{"creatorid": 1, "_id": 0})

	err := ad.groupColl().FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("group not found")
		}
		return 0, err
	}
	return result.CreatorID, nil
}

func (ad *ChatRepository) ExistingMembers(groupId string) ([]uint64, error) {
	var result struct {
		GroupMembers []uint64 `bson:"groupmembers"`
	}

	filter := bson.M{"groupid": groupId}
	opts := options.FindOne().SetProjection(bson.M{"groupmembers": 1, "_id": 0})

	err := ad.groupColl().FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %v", domain.ErrGroupNotFound, err)
		}
		return nil, err
	}
	return result.GroupMembers, nil
}

func (ad *ChatRepository) AddMembers(req requestmodels.AddMembersRequest) (responsemodels.AddMembersResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Update the Group Collection (Source of Truth)
	groupFilter := bson.M{"groupid": req.GroupID}
	groupUpdate := bson.M{
		"$addToSet": bson.M{"groupmembers": bson.M{"$each": req.GroupMembers}},
		"$set":      bson.M{"updatedat": time.Now()},
	}

	var updatedDoc struct {
		GroupID      string   `bson:"groupid"`
		GroupMembers []uint64 `bson:"groupmembers"`
		CreatorID    uint64   `bson:"creatorid"`
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := ad.groupColl().FindOneAndUpdate(ctx, groupFilter, groupUpdate, opts).Decode(&updatedDoc)
	if err != nil {
		return responsemodels.AddMembersResponse{}, err
	}
	// 2. IMMEDIATE VISIBILITY: Sync to Conversations Collection
	// We use Upsert so the group appears in the inbox immediately.
	convFilter := bson.M{"group_id": req.GroupID, "type": "group"}
	convUpdate := bson.M{
		"$addToSet": bson.M{"participants": bson.M{"$each": req.GroupMembers}},
		"$setOnInsert": bson.M{
			"conversation_id":   uuid.NewString(),
			"type":              "group",
			"last_message":      "New members added to the group", // Placeholder text
			"last_message_time": time.Now(),
		},
	}

	// SetUpsert(true) creates the doc if it's the first activity in this group
	convOpts := options.Update().SetUpsert(true)
	_, err = ad.convColl().UpdateOne(ctx, convFilter, convUpdate, convOpts)
	if err != nil {
		return responsemodels.AddMembersResponse{}, err
	}

	return responsemodels.AddMembersResponse{
		GroupID:      updatedDoc.GroupID,
		GroupMembers: updatedDoc.GroupMembers,
		UserID:       updatedDoc.CreatorID,
	}, nil
}

func (ad *ChatRepository) RemoveMember(req requestmodels.RemoveMemberRequest) (responsemodels.RemoveMemberResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Update the Group Collection
	groupFilter := bson.M{"groupid": req.GroupID}
	groupUpdate := bson.M{
		"$pull": bson.M{"groupmembers": req.MemberID},
		"$set":  bson.M{"updatedat": time.Now()},
	}

	var updatedDoc struct {
		GroupID      string   `bson:"groupid"`
		CreatorID    uint64   `bson:"creatorid"`
		GroupMembers []uint64 `bson:"groupmembers"`
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := ad.groupColl().FindOneAndUpdate(ctx, groupFilter, groupUpdate, opts).Decode(&updatedDoc)
	if err != nil {
		return responsemodels.RemoveMemberResponse{}, err
	}

	// 2. SYNC: Remove from the Conversation Participants
	// No Upsert here: if the chat doesn't exist, we don't create it just to remove someone.
	convFilter := bson.M{"group_id": req.GroupID, "type": "group"}
	convUpdate := bson.M{
		"$pull": bson.M{"participants": req.MemberID},
	}

	_, err = ad.convColl().UpdateOne(ctx, convFilter, convUpdate)
	if err != nil {
		return responsemodels.RemoveMemberResponse{}, err
	}

	return responsemodels.RemoveMemberResponse{
		GroupID:   updatedDoc.GroupID,
		MemberID:  req.MemberID,
		CreatorID: updatedDoc.CreatorID,
	}, nil
}

func (ad *ChatRepository) FetchMembersOfGroup(groupId string) ([]uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Now we just use the string directly - no more parsing!
	filter := bson.M{"groupid": groupId}

	// Projection: only get the members array to keep it fast
	opts := options.FindOne().SetProjection(bson.M{"groupmembers": 1, "_id": 0})

	var result struct {
		GroupMembers []int64 `bson:"groupmembers"`
	}

	err := ad.MongoClient.Database("chat").Collection("group").
		FindOne(ctx, filter, opts).
		Decode(&result)

	if err != nil {
		return nil, err
	}

	// Convert int64 from Mongo to uint64 for your Hub
	members := make([]uint64, len(result.GroupMembers))
	for i, v := range result.GroupMembers {
		members[i] = uint64(v)
	}

	return members, nil
}

func (ad *ChatRepository) StoreIndividualChatInMessages(dm domain.Message) error {
	collection := ad.MongoClient.Database("chat").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messageDoc := bson.M{
		"message_id":      dm.MessageID,
		"conversation_id": dm.ConversationID, // New Field
		"sender_id":       dm.SenderID,
		"recipient_id":    dm.RecipientID,
		"content":         dm.Content,
		"created_at":      dm.CreatedAt,
		"type":            dm.Type,
		"status":          "sent",
	}

	_, err := collection.InsertOne(ctx, messageDoc)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

func (ad *ChatRepository) StoreOrUpdateIndividualChatInConversation(conv domain.Conversation) (string, error) {
	collection := ad.MongoClient.Database("chat").Collection("conversations")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Sort participants (Crucial for deterministic matching)
	sortedParticipants := make([]uint64, len(conv.Participants))
	copy(sortedParticipants, conv.Participants)
	sort.Slice(sortedParticipants, func(i, j int) bool {
		return sortedParticipants[i] < sortedParticipants[j]
	})

	// 2. Filter
	filter := bson.M{
		"participants": sortedParticipants,
		"type":         "individual",
	}

	// 3. Update logic
	update := bson.M{
		"$set": bson.M{
			"last_message":      conv.LastMessage,
			"last_message_time": conv.LastMessageTime,
		},
		"$setOnInsert": bson.M{
			"conversation_id": conv.ConversationID,
			"participants":    sortedParticipants, // Include here for the new document
			"type":            "individual",
		},
	}

	// 4. Options: ReturnDocument(After) gives us the doc AFTER the upsert
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result struct {
		ConversationID string `bson:"conversation_id"`
	}

	// FindOneAndUpdate executes the update and decodes the result into our struct
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.ConversationID, nil
}

func (ad *ChatRepository) StoreGroupChatInMessages(gm domain.Message) error {
	collection := ad.MongoClient.Database("chat").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messageDoc := bson.M{
		"message_id":      gm.MessageID,
		"conversation_id": gm.ConversationID, // Added this
		"sender_id":       gm.SenderID,
		"group_id":        gm.GroupID,
		"content":         gm.Content,
		"created_at":      gm.CreatedAt,
		"type":            gm.Type,
		"status":          "sent",
	}

	_, err := collection.InsertOne(ctx, messageDoc)
	return err
}

func (ad *ChatRepository) StoreOrUpdateGroupChatInConversation(conv domain.Conversation) (string, error) {
	collection := ad.MongoClient.Database("chat").Collection("conversations")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"group_id": conv.GroupID,
		"type":     "group",
	}

	update := bson.M{
		"$set": bson.M{
			"last_message":      conv.LastMessage,
			"last_message_time": conv.LastMessageTime,
			"participants":      conv.Participants, // Important: update members in case someone joined/left
		},
		"$setOnInsert": bson.M{
			"conversation_id": conv.ConversationID,
			"group_id":        conv.GroupID,
			"type":            "group",
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		ConversationID string `bson:"conversation_id"`
	}

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.ConversationID, nil
}

func (ad *ChatRepository) GetUserConversation(req requestmodels.RecentChatProfilesRequest) ([]domain.Conversation, error) {
	collection := ad.MongoClient.Database("chat").Collection("conversations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Filter: Find all conversations where the UserID is in the participants array
	filter := bson.M{"participants": req.UserID}

	// 2. Options: Handle Pagination and Sorting
	findOptions := options.Find()
	findOptions.SetLimit(int64(req.Limit))
	findOptions.SetSkip(int64(req.Offset))

	// Always sort by latest message first for a "Recent Chats" view
	findOptions.SetSort(bson.D{{Key: "last_message_time", Value: -1}})

	// 3. Execution
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %v",err,"error in fetching conversations of user_id")
	}
	defer cursor.Close(ctx)

	var conversations []domain.Conversation
	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, fmt.Errorf("%w: %v",err,"error decoding each mongo document into struct")
	}

	return conversations, nil
}

// func (ad *ChatRepository) GetGroupNamesBatch(groupIDs []string) (map[string]string, error) {
// 	collection := ad.MongoClient.Database("chat").Collection("group")

// 	// Query: { groupid: { $in: ["id1", "id2"...] } }
// 	filter := bson.M{"groupid": bson.M{"$in": groupIDs}}
// 	projection := bson.M{"groupid": 1, "groupname": 1, "_id": 0}

// 	cursor, err := collection.Find(context.Background(), filter, options.Find().SetProjection(projection))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(context.Background())

//		results := make(map[string]string)
//		for cursor.Next(context.Background()) {
//			var temp struct {
//				ID   string `bson:"groupid"`
//				Name string `bson:"groupname"`
//			}
//			if err := cursor.Decode(&temp); err == nil {
//				results[temp.ID] = temp.Name
//			}
//		}
//		return results, nil
//	}
func (ad *ChatRepository) GetGroupMetaBatch(groupIDs []string) (map[string]responsemodels.GroupMeta, error) {
	collection := ad.MongoClient.Database("chat").Collection("group")

	filter := bson.M{"groupid": bson.M{"$in": groupIDs}}
	projection := bson.M{
		"groupid":       1,
		"groupname":     1,
		"groupimageurl": 1,
		"_id":           0,
	}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("%w: %v",err,"error in finding group documents")
	}
	defer cursor.Close(context.Background())

	results := make(map[string]responsemodels.GroupMeta)

	for cursor.Next(context.Background()) {
		var temp struct {
			ID       string `bson:"groupid"`
			Name     string `bson:"groupname"`
			ImageURL string `bson:"groupimageurl"`
		}
		// results[temp.ID] = responsemodels.GroupMeta{
		// 	Name:     temp.Name,
		// 	ImageURL: temp.ImageURL, // may be empty
		// }
		if err := cursor.Decode(&temp); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (ad *ChatRepository) GetUserMessagesByConversationId(req requestmodels.GetChatRequest) ([]domain.Message, error) {
	collection := ad.MongoClient.Database("chat").Collection("messages")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query messages belonging to this specific conversation
	filter := bson.M{"conversation_id": req.ConvID}

	findOptions := options.Find()
	findOptions.SetLimit(int64(req.Limit))
	findOptions.SetSkip(int64(req.Offset))
	// Sort by newest first so pagination works correctly (scrolling up for older messages)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []domain.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("decode messages: %w", err)
	}

	return messages, nil
}

func (ad *ChatRepository) IsUserInConversation(convID string, userID uint64) (bool, error) {
	collection := ad.MongoClient.Database("chat").Collection("conversations")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	filter := bson.M{
		"conversation_id": convID,
		// Convert uint64 to int64 to ensure it matches the Long type in MongoDB
		"participants": int64(userID),
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (ad *ChatRepository) GetGroupNameByGroupID(groupID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"groupid": groupID}

	var result struct {
		GroupName string `bson:"groupname"`
	}

	err := ad.groupColl().FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("group not found")
		}
		return "", err
	}

	return result.GroupName, nil
}

func (r *ChatRepository) SetGroupProfileImage(groupID string, imageURL string) error {
	filter := bson.M{
		"groupid": groupID,
	}

	update := bson.M{
		"$set": bson.M{
			"groupimageurl": imageURL,
			"updatedat":     time.Now(),
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := r.groupColl().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// if no document updated → group not found
	if result.MatchedCount == 0 {
		return domain.ErrGroupNotFound
	}

	return nil
}
