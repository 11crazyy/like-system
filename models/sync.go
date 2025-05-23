package models

import (
	"fmt"
	"time"
)

func GetPreviousTimeSlice() string {
	//获取上一个时间片
	now := time.Now()
	seconds := now.Second()
	sliceSeconds := (seconds/10 - 1) * 10
	if sliceSeconds < 0 { //0-9取上一分钟的50-59
		sliceSeconds = 50
		now = now.Add(-1 * time.Minute)
	}
	return now.Format("15:04:" + fmt.Sprintf("%02d", sliceSeconds))
}
