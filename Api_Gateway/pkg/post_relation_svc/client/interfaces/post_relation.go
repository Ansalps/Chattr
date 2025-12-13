package interfaces

import (
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/responsemodels"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/post_relation_svc/requestmodels"
)

type PostRelationClient interface {
	CreatePost(requestmodels.CreatePostRequest) (responsemodels.CreatePostResponse,error)
}
