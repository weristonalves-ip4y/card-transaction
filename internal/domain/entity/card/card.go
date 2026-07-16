package card

import (
	"card-transaction/internal/domain/vo"
)

type Status int

const (
	StatusUnknown  Status = 0
	StatusBlocked  Status = 3
	StatusActive   Status = 4
	StatusCanceled Status = 5
)

type Card struct {
	ID           string
	AccountID    string
	Status       Status
	ProductCode  string
	CheckLimit   bool
	MonthlyLimit vo.Money
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
	return c.Status == StatusActive
}

func (c Card) IsBlocked() bool {
	return c.Status == StatusBlocked
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

	if !c.CheckLimit {
		return false
	}

	return monthlySum.Add(transactionValue).GreaterThan(c.MonthlyLimit)
}

/**
 * transactionProductCode: Código do produto da transação
 * ProductCode: Código do produto do cartão
 * returns: true se o cartão suporta o produto da transação, false caso contrário
 */

func ValidateProductCompatibility(c Card, transactionProductCode string) ValidationResult {

	isVoucherTx := transactionProductCode == "011401"

	if c.ProductCode == "011401" && !isVoucherTx {
		return Rejected("06", "product not allowed for this card")
	}

	if (c.ProductCode == "011202" || c.ProductCode == "") && isVoucherTx {
		return Rejected("06", "product not allowed for this card")
	}

	return Approved()
}

func (c Card) IsVoucher() bool {
	if c.ProductCode == "011401" {
		return true
	}
	return false
}

func (c Card) IsCard() bool {
	if c.ProductCode == "011202" || c.ProductCode == "" {
		return true
	}
	return false
}

func (c Card) HasValidatedAccount() bool {
	if c.ID == "" {
		return false
	}
	if c.AccountID == "" {
		return false
	}
	if !c.IsActive() {
		return false
	}
	return true
}

func ValidateCard(c Card) ValidationResult {
	if c.ID == "" {
		return Rejected("02", "card not found")
	}
	if c.AccountID == "" {
		return Rejected("02", "card account not found")
	}
	if !c.IsActive() {
		return Rejected("03", "card blocked or inactive")
	}
	return Approved()
}
