package interfaces

import (
	"context"

	"github.com/tsigemariamzewdu/JobMate-backend/domain/models"
)

type SkillGapRepository interface {
	CreateMany(ctx context.Context, gaps []*models.SkillGap) error
	GetByUserID(ctx context.Context, userID string) ([]*models.SkillGap, error)
}
