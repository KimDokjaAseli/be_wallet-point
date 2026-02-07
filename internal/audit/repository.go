package audit

import (
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) FindAll(params AuditListParams) ([]AuditLogWithUser, int64, error) {
	var logs []AuditLogWithUser
	var total int64

	query := r.db.Model(&AuditLog{})

	if params.UserID > 0 {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Action != "" {
		query = query.Where("action = ?", params.Action)
	}
	if params.Date != "" {
		query = query.Where("DATE(created_at) = ?", params.Date)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit

	// Join with users to get name/role
	err := r.db.Table("audit_logs").
		Select("audit_logs.*, users.full_name as user_name, users.role as user_role").
		Joins("LEFT JOIN users ON users.id = audit_logs.user_id").
		Order("audit_logs.created_at DESC").
		Limit(params.Limit).
		Offset(offset).
		Scan(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
