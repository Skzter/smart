package entity

import shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

// DatabaseEntry represents a database record containing a request, response, and associated tags.
type DatabaseEntry struct {
	Request  Request         `json:"request"`
	Response Response        `json:"response"`
	Tags     *shared.TagList `json:"tags"`
	Updated  bool            `json:"updated"`
}
