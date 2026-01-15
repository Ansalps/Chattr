package repository

import (
	"context"
	"fmt"
	"log"
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
	return ad.MongoClient.Database("message").Collection("conversations")
}
func (ad *ChatRepository) CreateGroup(req requestmodels.CreateGroupRequest) (responsemodels.CreateGroupResponse, error) {
	groupCollection := ad.MongoClient.Database("chat").Collection("group")

	// Insert the request object directly
	// Note: Ensure your struct has `bson` tags (see below)
	_, err := groupCollection.InsertOne(context.TODO(), req)
	if err != nil {
		log.Println("Error inserting group into MongoDB:", err)
		return responsemodels.CreateGroupResponse{}, err
	}

	return responsemodels.CreateGroupResponse{
		GroupID: req.GroupID,
	}, nil
}
func (ad *ChatRepository) CreatorID(groupId string) (uint64, error) {
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
			return nil, fmt.Errorf("group not found")
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
		log.Printf("Warning: Failed to sync conversation for group %s: %v", req.GroupID, err)
	}

	return responsemodels.AddMembersResponse{
		GroupID:      updatedDoc.GroupID,
		GroupMembers: updatedDoc.GroupMembers,
		CreatorID:    updatedDoc.CreatorID,
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
        log.Printf("Warning: Failed to remove participant from conversation %s: %v", req.GroupID, err)
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
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("group not found with ID: %s", groupId)
		}
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

	// Create a context with timeout for the database operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mapping to BSON for MongoDB
	// We use lowercase field names to match standard MongoDB conventions
	messageDoc := bson.M{
		"message_id":   dm.MessageID,
		"sender_id":    dm.SenderID,
		"recipient_id": dm.RecipientID,
		"content":      dm.Content,
		"created_at":   dm.CreatedAt,
		"type":         dm.Type,
		"status":       "sent", // Default status when first stored
	}

	_, err := collection.InsertOne(ctx, messageDoc)
	if err != nil {
		log.Printf("Error inserting message: %v", err)
		return err
	}

	return nil
}

func (ad *ChatRepository) StoreOrUpdateIndividualChatInConversation(conv domain.Conversation) error {
	collection := ad.MongoClient.Database("chat").Collection("conversations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. FILTER: Look for a conversation containing BOTH users.
	// $all ensures that the order [A, B] or [B, A] doesn't create two different chats.
	filter := bson.M{
		"participants": bson.M{"$all": conv.Participants},
	}

	// 2. UPDATE:
	// $set: Always update these (the latest message info)
	// $setOnInsert: ONLY set these if we are creating a NEW document
	update := bson.M{
		"$set": bson.M{
			"last_message":      conv.LastMessage,
			"last_message_time": conv.LastMessageTime,
		},
		"$setOnInsert": bson.M{
			"conversation_id": conv.ConversationID, // Only saved once
			"participants":    conv.Participants,   // Only saved once
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (ad *ChatRepository) StoreGroupChatInMessages(gm domain.Message) error {
	collection := ad.MongoClient.Database("chat").Collection("messages")

	// Create a context with timeout for the database operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mapping to BSON for MongoDB
	// We use lowercase field names to match standard MongoDB conventions
	messageDoc := bson.M{
		"message_id": gm.MessageID,
		"sender_id":  gm.SenderID,
		"group_id":   gm.GroupID,
		"content":    gm.Content,
		"created_at": gm.CreatedAt,
		"type":       gm.Type,
		"status":     "sent", // Default status when first stored
	}

	_, err := collection.InsertOne(ctx, messageDoc)
	if err != nil {
		log.Printf("Error inserting message: %v", err)
		return err
	}

	return nil
}

func (ad *ChatRepository) StoreOrUpdateGroupChatInConversation(conv domain.Conversation) error {
	collection := ad.MongoClient.Database("chat").Collection("conversations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. FILTER: For groups, we identify the unique conversation by GroupID.
	filter := bson.M{
		"group_id": conv.GroupID,
		"type":     "group",
	}

	// 2. UPDATE:
	// $set: Updates the preview text and time every time a new message is sent.
	// $setOnInsert: Only sets these fields the very first time the group chat is initialized.
	update := bson.M{
		"$set": bson.M{
			"last_message":      conv.LastMessage,
			"last_message_time": conv.LastMessageTime,
		},
		"$setOnInsert": bson.M{
			"conversation_id": conv.ConversationID,
			"participants":    conv.Participants, // All group members
			"type":            "group",
		},
	}

	// 3. Upsert: true tells MongoDB to create the document if it doesn't find the group_id.
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update group conversation: %v", err)
	}

	return nil
}
