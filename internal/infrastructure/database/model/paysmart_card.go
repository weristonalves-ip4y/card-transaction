package model

import "time"

type PaysmartCard struct {
	ID               int64      `db:"id"`
	AccountID        string     `db:"account_id"`
	UUID             string     `db:"uuid"`
	CardID           string     `db:"card_id"`
	CardStatusID     int        `db:"card_status_id"`
	CardMonthlyLimit int64      `db:"card_monthly_limit"`
	CardCheckLimit   bool       `db:"card_check_limit"`
	PsProductCode    string     `db:"ps_product_code"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at"`
}

func (PaysmartCard) TableName() string {
	return "paysmart_cards"
}
