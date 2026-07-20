package card

import (
	"card-transaction/internal/domain/vo"
	"time"
)

type Status int

const (
	StatusUnknown  Status = 0
	StatusBlocked  Status = 3
	StatusActive   Status = 4
	StatusCanceled Status = 5
)

type Card struct {
	ID               int64
	AccountID        int64
	UUID             string
	CardID           string
	CardStatusID     int
	CardMonthlyLimit vo.Money
	CardCheckLimit   bool
	FourLastDigits   string
	PsProductCode    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type ValidationResult struct {
	Approved bool
	Code     string
	Reason   string
}

func Approved() ValidationResult {
	return ValidationResult{Approved: true, Code: "00"}
}

func Rejected(code, reason string) ValidationResult {
	return ValidationResult{Approved: false, Code: code, Reason: reason}
}

func (c Card) IsActive() bool {
	return c.CardStatusID == int(StatusActive)
}

func (c Card) IsBlocked() bool {
	return c.CardStatusID == int(StatusBlocked)
}

/*
*
  - monthlySum: Soma das transações do mês
  - transactionValue: Valor da transação atual
  - forceAccept: Se a transação deve ser aceita mesmo que o limite seja excedido
    returns: true se o limite mensal do cartão for excedido, false caso contrário
*/
func MonthsLimitExceeded(c Card, monthlySum vo.Money, transactionValue vo.Money, forceAccept bool) bool {
	if forceAccept {
		return false
	}

	if !c.CardCheckLimit {
		return false
	}

	return monthlySum.Add(transactionValue).GreaterThan(c.CardMonthlyLimit)
}

/**
 * transactionProductCode: Código do produto da transação
 * ProductCode: Código do produto do cartão
 * returns: true se o cartão suporta o produto da transação, false caso contrário
 */

func ValidateProductCompatibility(c Card, transactionProductCode string) ValidationResult {

	isVoucherTx := transactionProductCode == "011401"

	if c.PsProductCode == "011401" && !isVoucherTx {
		return Rejected("06", "product not allowed for this card")
	}

	if (c.PsProductCode == "011202" || c.PsProductCode == "") && isVoucherTx {
		return Rejected("06", "product not allowed for this card")
	}

	return Approved()
}

func (c Card) IsVoucher() bool {
	if c.PsProductCode == "011401" {
		return true
	}
	return false
}

func (c Card) IsCard() bool {
	if c.PsProductCode == "011202" || c.PsProductCode == "" {
		return true
	}
	return false
}

func (c Card) HasValidatedAccount() bool {
	if c.ID == 0 {
		return false
	}
	if c.AccountID == 0 {
		return false
	}
	if !c.IsActive() {
		return false
	}
	return true
}

func ValidateCard(c Card) ValidationResult {
	if c.ID == 0 {
		return Rejected("02", "card not found")
	}
	if c.AccountID == 0 {
		return Rejected("02", "card account not found")
	}
	if !c.IsActive() {
		return Rejected("03", "card blocked or inactive")
	}
	return Approved()
}
