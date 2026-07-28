package model

import "time"

// Template stores the dynamic HTML/Text content for notifications.
// I'm explicitly using a Composite Primary Key (Name + Version) here so I can safely insert 'v2' 
// without ever overwriting 'v1'. This guarantees I won't break mid-flight tasks!
type Template struct {
	Name            string `gorm:"type:varchar(100);primaryKey"`
	Version         string `gorm:"type:varchar(50);primaryKey"`
	SubjectTemplate string `gorm:"type:text;not null"`
	BodyTemplate    string `gorm:"type:text;not null"`
	CreatedAt       time.Time
}
