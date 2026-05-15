package service

import (
	"context"
	"encoding/json"
	"fmt"

	"dms-backend/internal/model"
	"dms-backend/internal/repository"
)

// OrgService — сервис для работы с организацией.
type OrgService struct {
	orgRepo *repository.OrgRepository
}

// NewOrgService создаёт сервис с зависимостями.
func NewOrgService(orgRepo *repository.OrgRepository) *OrgService {
	return &OrgService{orgRepo: orgRepo}
}

// Get возвращает текущую организацию (по org_id из контекста).
func (s *OrgService) Get(ctx context.Context) (*model.Organization, error) {
	return s.orgRepo.Get(ctx)
}

// Update обновляет данные организации. Принимает частичное обновление:
// nil-поля не трогают текущие значения.
func (s *OrgService) Update(ctx context.Context, name, inn *string, settings *json.RawMessage) (*model.Organization, error) {
	// Загружаем текущее состояние.
	org, err := s.orgRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("организация не найдена: %w", err)
	}

	// Применяем только переданные поля.
	if name != nil {
		org.Name = *name
	}
	if inn != nil {
		org.INN = *inn
	}
	if settings != nil {
		org.Settings = *settings
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}
