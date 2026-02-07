package audit

import (
	"math"
	"time"
)

type AuditService struct {
	repo *AuditRepository
}

func NewAuditService(repo *AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// LogActivity records a system activity
func (s *AuditService) LogActivity(params CreateAuditParams) error {
	log := &AuditLog{
		UserID:    params.UserID,
		Action:    params.Action,
		Entity:    params.Entity,
		EntityID:  params.EntityID,
		Details:   params.Details,
		IPAddress: params.IPAddress,
		UserAgent: params.UserAgent,
		CreatedAt: time.Now(),
	}

	return s.repo.Create(log)
}

// GetLogs retrieves logs for admin
func (s *AuditService) GetLogs(params AuditListParams) (*AuditListResponse, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	logs, total, err := s.repo.FindAll(params)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &AuditListResponse{
		Logs:       logs,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}
