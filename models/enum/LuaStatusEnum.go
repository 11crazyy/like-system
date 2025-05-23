package enum

type LuaStatus int64

const (
	LuaSuccess LuaStatus = 1
	LuaFail    LuaStatus = -1
)

func (s LuaStatus) String() string {
	switch s {
	case LuaSuccess:
		return "成功"
	case LuaFail:
		return "失败"
	default:
		return "未知状态"
	}
}
