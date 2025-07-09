package service

import "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"

type HistoryStorageService interface {
	SaveHistory(entity.SessionSummary) error
}
type historyStorageService struct {
}

func NewHistoryStorageService() HistoryStorageService {
	return &historyStorageService{}
}

// Die Methode, die vom Handler aufgerufen wird
func (s *historyStorageService) SaveHistory(summary entity.SessionSummary) error {
	return nil
}
