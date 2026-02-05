package mission

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"wallet-point/internal/wallet"

	"gorm.io/gorm"
)

type MissionService struct {
	repo          *MissionRepository
	walletService *wallet.WalletService
	db            *gorm.DB
}

func NewMissionService(repo *MissionRepository, walletService *wallet.WalletService, db *gorm.DB) *MissionService {
	return &MissionService{
		repo:          repo,
		walletService: walletService,
		db:            db,
	}
}

// Mission Management
func (s *MissionService) CreateMission(req *CreateMissionRequest, creatorID uint) (*Mission, error) {
	mission := &Mission{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Points:      req.Points,
		Deadline:    req.Deadline,
		Status:      "active",
		CreatorID:   creatorID,
	}

	if req.Type == "quiz" && len(req.Questions) > 0 {
		for _, q := range req.Questions {
			mission.Questions = append(mission.Questions, MissionQuestion{
				Question: q.Question,
				Options:  q.Options,
				Answer:   q.Answer,
			})
		}
	}

	if err := s.repo.Create(mission); err != nil {
		return nil, err
	}

	return mission, nil
}

func (s *MissionService) GetMissionByID(id uint) (*Mission, error) {
	return s.repo.FindByID(id)
}

func (s *MissionService) GetAllMissions(params MissionListParams) (*MissionListResponse, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	missions, total, err := s.repo.FindAll(params)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &MissionListResponse{
		Missions:   missions,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *MissionService) UpdateMission(id uint, req *UpdateMissionRequest) (*Mission, error) {
	// Check if exists
	_, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Points > 0 {
		updates["points_reward"] = req.Points
	}
	if req.Deadline != nil {
		updates["deadline"] = req.Deadline
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		if err := s.repo.Update(id, updates); err != nil {
			return nil, err
		}
	}

	// Handle questions update if provided
	if req.Questions != nil {
		err := s.db.Transaction(func(tx *gorm.DB) error {
			// Delete existing questions
			if err := tx.Where("mission_id = ?", id).Delete(&MissionQuestion{}).Error; err != nil {
				return err
			}

			// Add new questions
			for _, q := range req.Questions {
				newQ := MissionQuestion{
					MissionID: id,
					Question:  q.Question,
					Options:   q.Options,
					Answer:    q.Answer,
				}
				if err := tx.Create(&newQ).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(id)
}

func (s *MissionService) DeleteMission(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

// Submission Management
func (s *MissionService) SubmitMission(req *SubmitMissionRequest, studentID uint) (*MissionSubmission, error) {
	// Check if mission exists
	mission, err := s.repo.FindByID(req.MissionID)
	if err != nil {
		return nil, err
	}

	// Check deadline
	if mission.Deadline != nil && mission.Deadline.Before(s.db.NowFunc()) {
		return nil, errors.New("mission deadline has passed")
	}

	// Check duplicate submission
	exists, err := s.repo.CheckDuplicateSubmission(req.MissionID, studentID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("you have already submitted this mission")
	}

	submission := &MissionSubmission{
		MissionID: req.MissionID,
		StudentID: studentID,
		Content:   req.Content,
		FileURL:   req.FileURL,
		Status:    "pending",
		TimeTaken: req.TimeTaken,
	}

	// Auto-evaluate if it's a quiz
	if mission.Type == "quiz" && len(req.Answers) > 0 {
		totalWeight := 0
		earnedWeight := 0

		// Create a map for quick answer checking and weights
		type qInfo struct {
			Answer string
			Weight int
		}
		questionData := make(map[uint]qInfo)
		for _, q := range mission.Questions {
			w := q.Weight
			if w <= 0 {
				w = 1
			}
			questionData[q.ID] = qInfo{Answer: q.Answer, Weight: w}
			totalWeight += w
		}

		for _, ans := range req.Answers {
			if data, ok := questionData[ans.QuestionID]; ok {
				if ans.Answer == data.Answer {
					earnedWeight += data.Weight
				}
			}
		}

		// Calculate score (0-100)
		score := 0
		if totalWeight > 0 {
			score = int(float64(earnedWeight) / float64(totalWeight) * 100)
		}

		submission.Score = score
		submission.Status = "approved" // Quiz doesn't need dosen approval

		// Record answers in content as JSON
		answersBytes, _ := json.Marshal(req.Answers)
		submission.Content = string(answersBytes)

		// Use transaction to save submission and reward points
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(submission).Error; err != nil {
				return err
			}

			// Reward points based on score
			// Default conversion: pro-rated points
			// If score is 80 and mission.Points is 100, student gets 80 points.
			pointsReward := int(float64(score) / 100.0 * float64(mission.Points))

			if pointsReward > 0 {
				desc := fmt.Sprintf("Kuis: %s (Skor: %d)", mission.Title, score)
				err = s.walletService.ProcessMissionRewardWithTx(tx, studentID, pointsReward, desc, mission.ID, 0) // 0 for system/automatic
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return nil, err
		}

		return submission, nil
	}

	// Default task/assignment submission
	if len(req.Answers) > 0 {
		answersBytes, _ := json.Marshal(req.Answers)
		submission.Content = string(answersBytes)
	}

	if err := s.repo.CreateSubmission(submission); err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *MissionService) ReviewSubmission(submissionID uint, req *ReviewSubmissionRequest, reviewerID uint) error {
	submission, err := s.repo.FindSubmissionByID(submissionID)
	if err != nil {
		return err
	}

	if submission.Status != "pending" {
		return errors.New("submission has already been reviewed")
	}

	// Start a transaction for the review and potential wallet reward
	return s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status":          req.Status,
			"score":           req.Score,
			"validation_note": req.ReviewNote,
			"validated_by":    reviewerID,
		}

		// Update submission status
		if err := s.repo.UpdateSubmissionWithTx(tx, submissionID, updates); err != nil {
			return err
		}

		// If approved, reward points
		if req.Status == "approved" {
			mission, err := s.repo.FindByID(submission.MissionID)
			if err != nil {
				return err
			}

			// for now we'll assume it handles its own internal transaction if needed,
			// though nested transactions in GORM/MySQL are safe.
			err = s.walletService.ProcessMissionRewardWithTx(tx, submission.StudentID, mission.Points, mission.Title, mission.ID, reviewerID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *MissionService) GetAllSubmissions(params SubmissionListParams) (*SubmissionListResponse, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}

	submissions, total, err := s.repo.FindAllSubmissions(params)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &SubmissionListResponse{
		Submissions: submissions,
		Total:       total,
		Page:        params.Page,
		Limit:       params.Limit,
		TotalPages:  totalPages,
	}, nil
}

func (s *MissionService) GetDosenStats(dosenID uint) (*DosenStatsResponse, error) {
	var totalMissions int64
	var pendingReviews int64
	var validatedTasks int64

	// Count missions created by dosen
	if err := s.db.Model(&Mission{}).Where("creator_id = ?", dosenID).Count(&totalMissions).Error; err != nil {
		return nil, err
	}

	// Count pending submissions for missions created by this dosen
	if err := s.db.Table("mission_submissions").
		Joins("JOIN missions ON missions.id = mission_submissions.mission_id").
		Where("missions.creator_id = ? AND mission_submissions.status = ?", dosenID, "pending").
		Count(&pendingReviews).Error; err != nil {
		return nil, err
	}

	// Count validated submissions by this dosen
	if err := s.db.Model(&MissionSubmission{}).Where("validated_by = ?", dosenID).Count(&validatedTasks).Error; err != nil {
		return nil, err
	}

	return &DosenStatsResponse{
		TotalMissions:  totalMissions,
		PendingReviews: pendingReviews,
		ValidatedTasks: validatedTasks,
	}, nil
}

func (s *MissionService) GetQuizLeaderboard(missionID uint) ([]QuizLeaderboardEntry, error) {
	return s.repo.GetQuizLeaderboard(missionID)
}
