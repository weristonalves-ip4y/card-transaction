package validation

import (
	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/entity/transaction"
	"card-transaction/internal/domain/vo"
)

type Context struct {
	Transaction           transaction.Transaction
	Card                  card.Card
	IsDuplicate           bool
	CardProductCode       string
	MonthlyTransactionSum vo.Money
	Balance               vo.Money
}

func Validate(ctx Context) transaction.ValidationResult {
	if ctx.IsDuplicate {
		return transaction.Rejected("07", "duplicate transaction")
	}

	if result := ctx.Transaction.ValidateCard(ctx.Card); !result.Approved {
		return result
	}

	if result := ctx.Transaction.ValidateProductCompatibility(ctx.CardProductCode); !result.Approved {
		return result
	}

	if result := ctx.Transaction.ValidateMonthlyLimit(ctx.Card, ctx.MonthlyTransactionSum); !result.Approved {
		return result
	}

	if result := ctx.Transaction.ValidateBalance(ctx.Balance); !result.Approved {
		return result
	}

	return transaction.Approved()
}
