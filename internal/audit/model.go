package audit

import (
	"time"
)

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"` // User who performed the action
	Action    string    `json:"action" gorm:"size:100;not null"`
	Entity    string    `json:"entity" gorm:"size:100"` // e.g., "USER", "WALLET", "MISSION"
	EntityID  uint      `json:"entity_id"`
	Details   string    `json:"details" gorm:"type:text"`
	IPAddress string    `json:"ip_address" gorm:"size:45"`
	UserAgent string    `json:"user_agent" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type CreateAuditParams struct {
	UserID    uint
	Action    string
	Entity    string
	EntityID  uint
	Details   string
	IPAddress string
	UserAgent string
}

type AuditListParams struct {
	UserID int
	Action string
	Date   string
	Page   int
	Limit  int
}

type AuditListResponse struct {
	Logs       []AuditLogWithUser `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

type AuditLogWithUser struct {
	AuditLog
	UserName string `json:"user_name"`
	UserRole string `json:"user_role"`
}
