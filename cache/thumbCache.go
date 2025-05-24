package cache

// ThumbCache 定义点赞缓存操作的接口
type ThumbCache interface {
	HasThumb(blogId, userId int64) (bool, error)
	SaveThumb(blogId, userId int64) error
	DeleteThumb(blogId, userId int64) error
	MayHaveThumb(userId, blogId int64) bool
}

const (
	USER_THUMB_KEY_PREFIX = "user:thumb:"
	USER_THUMB_BLOOM_KEY  = "user:thumb:bloom:"
	NU_THUMB_CONSTANT     = 0
)
