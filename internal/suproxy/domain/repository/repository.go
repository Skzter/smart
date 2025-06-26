package repository

import (
	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

type Repository interface {
	CreateRequest(entity.DatabaseEntry) error
	ReadRequest(string) (entity.DatabaseEntry, error)
	UpdateRequest(string, entity.DatabaseEntry) error
	DeleteRequest(string) error
}
