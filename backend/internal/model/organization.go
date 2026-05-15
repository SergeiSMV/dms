package model

import (
	"encoding/json"
	"time"
)

// Organization — компания-клиент, использующая DMS.
// На on-premise инсталляции всегда одна организация.
type Organization struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	INN       string          `json:"inn"`
	Settings  json.RawMessage `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
