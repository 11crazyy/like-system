package util

import (
	"fmt"
	"time"
)

func GetTimeSLice() string {
	now := time.Now()
	seconds := now.Second()
	slice := (seconds / 10) * 10
	return now.Format("15:04:") + fmt.Sprintf("%02d", slice)
}
