package mission

import (
	"errors"

	"gorm.io/gorm"
)

type MissionRepository struct {
	db *gorm.DB
}

func NewMissionRepository(db *gorm.DB) *MissionRepository {
	return &MissionRepository{db: db}
}

// Mission CRUD
func (r *MissionRepository) Create(mission *Mission) error {
	return r.db.Create(mission).Error
}

func (r *MissionRepository) FindByID(id uint) (*Mission, error) {
	var mission Mission
	err := r.db.Preload("Questions").First(&mission, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mission not found")
		}
		return nil, err
	}
	return &mission, nil
}

func (r *MissionRepository) FindAll(params MissionListParams) ([]MissionWithCreator, int64, error) {
	var missions []MissionWithCreator
	var total int64

	query := r.db.Table("missions").
		Select("missions.*, users.full_name as creator_name, users.email as creator_email").
		Joins("LEFT JOIN users ON users.id = missions.creator_id")

	if params.Type != "" {
		query = query.Where("missions.type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("missions.status = ?", params.Status)
	}
	if params.CreatedBy > 0 {
		query = query.Where("missions.creator_id = ?", params.CreatedBy)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	err := query.Order("missions.created_at DESC").
		Limit(params.Limit).
		Offset(offset).
		Scan(&missions).Error

	return missions, total, err
}

func (r *MissionRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Mission{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MissionRepository) Delete(id uint) error {
	return r.db.Delete(&Mission{}, id).Error
}

// Submission CRUD
func (r *MissionRepository) CreateSubmission(submission *MissionSubmission) error {
	return r.db.Create(submission).Error
}

func (r *MissionRepository) FindSubmissionByID(id uint) (*MissionSubmission, error) {
	var submission MissionSubmission
	err := r.db.First(&submission, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("submission not found")
		}
		return nil, err
	}
	return &submission, nil
}

func (r *MissionRepository) FindAllSubmissions(params SubmissionListParams) ([]SubmissionWithDetails, int64, error) {
	var submissions []SubmissionWithDetails
	var total int64

	query := r.db.Table("mission_submissions").
		Select("mission_submissions.*, missions.title as mission_title, users.full_name as student_name, users.nim_nip as student_nim, reviewers.full_name as reviewer_name").
		Joins("LEFT JOIN missions ON missions.id = mission_submissions.mission_id").
		Joins("LEFT JOIN users ON users.id = mission_submissions.student_id").
		Joins("LEFT JOIN users as reviewers ON reviewers.id = mission_submissions.validated_by")

	if params.MissionID > 0 {
		query = query.Where("mission_submissions.mission_id = ?", params.MissionID)
	}
	if params.StudentID > 0 {
		query = query.Where("mission_submissions.student_id = ?", params.StudentID)
	}
	if params.CreatorID > 0 {
		query = query.Where("missions.creator_id = ?", params.CreatorID)
	}
	if params.Status != "" {
		query = query.Where("mission_submissions.status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	err := query.Order("mission_submissions.created_at DESC").
		Limit(params.Limit).
		Offset(offset).
		Scan(&submissions).Error

	return submissions, total, err
}

func (r *MissionRepository) UpdateSubmission(id uint, updates map[string]interface{}) error {
	return r.db.Model(&MissionSubmission{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MissionRepository) UpdateSubmissionWithTx(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&MissionSubmission{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MissionRepository) CheckDuplicateSubmission(missionID, studentID uint) (bool, error) {
	var count int64
	err := r.db.Model(&MissionSubmission{}).
		Where("mission_id = ? AND student_id = ?", missionID, studentID).
		Count(&count).Error
	return count > 0, err
}

func (r *MissionRepository) GetQuizLeaderboard(missionID uint) ([]QuizLeaderboardEntry, error) {
	var entries []QuizLeaderboardEntry

	err := r.db.Table("mission_submissions").
		Select("mission_submissions.student_id, users.full_name as student_name, users.nim_nip as student_nim, mission_submissions.score, mission_submissions.time_taken, mission_submissions.created_at as finished_at").
		Joins("JOIN users ON users.id = mission_submissions.student_id").
		Where("mission_submissions.mission_id = ? AND mission_submissions.status = ?", missionID, "approved").
		Order("mission_submissions.score DESC, mission_submissions.time_taken ASC, mission_submissions.created_at ASC").
		Limit(50).
		Scan(&entries).Error

	return entries, err
}
