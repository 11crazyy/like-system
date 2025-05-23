package enum

type ThumbType int

const (
	ThumbIncr ThumbType = 1  //点赞
	ThumbDecr ThumbType = -1 //取消点赞
	ThumbNon  ThumbType = 0  //不发生改变
)

func (t ThumbType) String() string {
	switch t {
	case ThumbIncr:
		return "点赞"
	case ThumbDecr:
		return "取消点赞"
	case ThumbNon:
		return "不发生改变"
	default:
		return "未知类型"
	}
}
