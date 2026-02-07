package mission

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONOptions []string

func (jo JSONOptions) Value() (driver.Value, error) {
	return json.Marshal(jo)
}

func (jo *JSONOptions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, jo)
}

type Mission struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	CreatorID   uint              `json:"creator_id" gorm:"column:creator_id;not null"`
	Title       string            `json:"title" gorm:"not null"`
	Description string            `json:"description" gorm:"type:text"`
	Type        string            `json:"type" gorm:"type:enum('quiz','task','assignment');not null"`
	Points      int               `json:"points" gorm:"column:points_reward;not null"`
	Deadline    *time.Time        `json:"deadline" gorm:"column:deadline"`
	Status      string            `json:"status" gorm:"type:enum('active','inactive','expired');default:'active'"`
	Questions   []MissionQuestion `json:"questions,omitempty" gorm:"foreignKey:MissionID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type MissionQuestion struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	MissionID uint        `json:"mission_id" gorm:"not null;index"`
	Question  string      `json:"question" gorm:"type:text;not null"`
	Options   JSONOptions `json:"options" gorm:"type:json"` // Array of strings (options)
	Answer    string      `json:"answer" gorm:"not null"`   // Correct answer or index
	Weight    int         `json:"weight" gorm:"default:1"`
}

func (MissionQuestion) TableName() string {
	return "mission_questions"
}

func (Mission) TableName() string {
	return "missions"
}

type MissionSubmission struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	MissionID  uint      `json:"mission_id" gorm:"not null;index"`
	StudentID  uint      `json:"student_id" gorm:"not null;index"`
	Content    string    `json:"content" gorm:"column:submission_content;type:text"`
	FileURL    string    `json:"file_url" gorm:"size:500"`
	Score      int       `json:"score" gorm:"default:0"`
	TimeTaken  int       `json:"time_taken" gorm:"default:0"` // in seconds
	Status     string    `json:"status" gorm:"type:enum('pending','approved','rejected');default:'pending'"`
	ReviewedBy *uint     `json:"reviewed_by" gorm:"column:validated_by"`
	ReviewNote string    `json:"review_note" gorm:"column:validation_note;type:text"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (MissionSubmission) TableName() string {
	return "mission_submissions"
}

type CreateMissionRequest struct {
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description"`
	Type        string            `json:"type" binding:"required,oneof=quiz task assignment"`
	Points      int               `json:"points" binding:"required,gt=0"`
	Deadline    *time.Time        `json:"deadline"`
	Questions   []QuestionRequest `json:"questions"`
}

type QuestionRequest struct {
	Question string   `json:"question" binding:"required"`
	Options  []string `json:"options"`
	Answer   string   `json:"answer" binding:"required"`
}

type UpdateMissionRequest struct {
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Points      int               `json:"points,omitempty" binding:"omitempty,gt=0"`
	Deadline    *time.Time        `json:"deadline,omitempty"`
	Status      string            `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	Questions   []QuestionRequest `json:"questions,omitempty"`
}

type SubmitMissionRequest struct {
	MissionID uint               `json:"mission_id" binding:"required"`
	Content   string             `json:"content"`
	FileURL   string             `json:"file_url"`
	Answers   []AnswerSubmission `json:"answers"`
	TimeTaken int                `json:"time_taken"` // in seconds
}

type AnswerSubmission struct {
	QuestionID uint   `json:"question_id"`
	Answer     string `json:"answer"`
}

type ReviewSubmissionRequest struct {
	Status     string `json:"status" binding:"required,oneof=approved rejected"`
	Score      int    `json:"score" binding:"gte=0"`
	ReviewNote string `json:"review_note"`
}

type MissionWithCreator struct {
	Mission
	CreatorName  string `json:"creator_name"`
	CreatorEmail string `json:"creator_email"`
}

type SubmissionWithDetails struct {
	MissionSubmission
	MissionTitle string `json:"mission_title"`
	StudentName  string `json:"student_name"`
	StudentNim   string `json:"student_nim"`
	ReviewerName string `json:"reviewer_name,omitempty"`
}

type MissionListParams struct {
	Type      string
	Status    string
	CreatedBy uint
	Page      int
	Limit     int
}

type MissionListResponse struct {
	Missions   []MissionWithCreator `json:"missions"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

type SubmissionListParams struct {
	MissionID uint
	StudentID uint
	CreatorID uint // Filter submissions for missions created by this user
	Status    string
	Page      int
	Limit     int
}

type SubmissionListResponse struct {
	Submissions []SubmissionWithDetails `json:"submissions"`
	Total       int64                   `json:"total"`
	Page        int                     `json:"page"`
	Limit       int                     `json:"limit"`
	TotalPages  int                     `json:"total_pages"`
}

type QuizLeaderboardEntry struct {
	StudentID   uint      `json:"student_id"`
	StudentName string    `json:"student_name"`
	StudentNim  string    `json:"student_nim"`
	Score       int       `json:"score"`
	TimeTaken   int       `json:"time_taken"`
	FinishedAt  time.Time `json:"finished_at"`
}

type DosenStatsResponse struct {
	TotalMissions  int64 `json:"total_missions"`
	PendingReviews int64 `json:"pending_reviews"`
	ValidatedTasks int64 `json:"validated_tasks"`
}
