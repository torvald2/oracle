package main

import (
	"time"

	"github.com/google/uuid"
)

// OpenSeaAttribute представляет атрибут в метаданных OpenSea.
type OpenSeaAttribute struct {
	TraitType   string      `json:"trait_type"`
	Value       interface{} `json:"value"`
	DisplayType string      `json:"display_type,omitempty"`
}

// OpenSeaMetadata представляет полную структуру метаданных для OpenSea.
type OpenSeaMetadata struct {
	Description string             `json:"description"`
	ExternalURL string             `json:"external_url"`
	Image       string             `json:"image"`
	Name        string             `json:"name"`
	Attributes  []OpenSeaAttribute `json:"attributes"`
}

// Well представляет запись о скважине из таблицы wells.
type Well struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key"`
	WellID   string    `gorm:"column:well_id;type:text;unique"`
	WellName string    `gorm:"column:well_name;type:text"`
}

// TableName указывает GORM, что эта структура сопоставляется с таблицей 'wells'.
func (Well) TableName() string {
	return "wells"
}

// Valuation represents the valuation metrics and economic assumptions for a well.
type Valuation struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WellID               uuid.UUID  `gorm:"type:uuid;not null;unique"`
	NpvUsd               *float64   `gorm:"column:npv_usd;type:decimal(15,2)"`
	MarketValueUsd       *float64   `gorm:"column:market_value_usd;type:decimal(15,2)"`
	DiscountPct          *float64   `gorm:"column:discount_pct;type:decimal(5,2)"`
	Confidence           *float64   `gorm:"column:confidence;type:decimal(3,2)"`
	RemainingReservesBbl *float64   `gorm:"column:remaining_reserves_bbl;type:decimal(15,2)"`
	CalculatedAt         *time.Time `gorm:"column:calculated_at;type:timestamp with time zone"`
	OilPriceUsd          *float64   `gorm:"column:oil_price_usd;type:decimal(8,2);default:75.00"`
	OperatingCostPerBbl  *float64   `gorm:"column:operating_cost_per_bbl;type:decimal(8,2);default:15.00"`
	DiscountRate         *float64   `gorm:"column:discount_rate;type:decimal(5,4);default:0.10"`
	RoyaltyRate          *float64   `gorm:"column:royalty_rate;type:decimal(5,4);default:0.20"`
	ValuationDate        *time.Time `gorm:"column:valuation_date;type:date;default:now()"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null;default:now()"`
}
