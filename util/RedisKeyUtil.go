package util

import (
	"fmt"
	"index/models"
)

func GetTempThumbKey(time string) string {
	return fmt.Sprintf("%s%s", models.TEMP_THUMB_KEY_PREFIX, time)
}

func GetUserThumbKey(loginUserId uint) string {
	return fmt.Sprintf("%s%d", models.USER_THUMB_KEY_PREFIX, loginUserId)
}
