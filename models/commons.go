package models

import (
	"gorm.io/gorm"
)

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
	)
}

func CheckDefaultValues(db *gorm.DB) {
	//carrot.CheckValue(db, KEY_NOTIFY_URL, "", carrot.ConfigFormatText, false, false)
	//carrot.CheckValue(db, KEY_SMTP_HOST, "", carrot.ConfigFormatText, false, false)
	//carrot.CheckValue(db, KEY_SMTP_PORT, "25", carrot.ConfigFormatInt, false, false)
	//carrot.CheckValue(db, KEY_SMTP_USER, "", carrot.ConfigFormatInt, false, false)
	//carrot.CheckValue(db, KEY_SMTP_PASSWORD, "", carrot.ConfigFormatSecurity, false, false)
	//carrot.CheckValue(db, KEY_VOICE_SERVER_HOST, "", carrot.ConfigFormatText, false, false)
}
