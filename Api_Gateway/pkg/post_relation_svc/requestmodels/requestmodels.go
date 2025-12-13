package requestmodels

type CreatePostRequest struct {
	UserID    uint64
	Caption   string
	MediaUrls []string
}
