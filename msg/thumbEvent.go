package msg

import "time"

type ThumbEvent struct {
	UserId    int64     `json:"user_id"`
	BlogId    int64     `json:"blog_id"`
	Type      EventType `json:"type"`
	EventTime time.Time `json:"event_time"` //事件发生时间
}

type EventType string

const (
	EventTypeINCR EventType = "INCR" //点赞
	EventTypeDECR EventType = "DECR" //取消点赞
)
