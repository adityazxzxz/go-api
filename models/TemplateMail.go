package models

import "time"

type EmailTemplate struct {
	ID           uint       `gorm:"primaryKey;autoIncrement"`
	UUID         string     `gorm:"type:char(36);uniqueIndex;not null"`
	TemplateName string     `gorm:"type:varchar(255);index;not null"`
	Body         string     `gorm:"type:text"`
	PrevData     *string    `gorm:"type:text"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    *time.Time `gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"index"`
	CreatedBy    *uint
	UpdatedBy    *uint
	DeletedBy    *uint
}
