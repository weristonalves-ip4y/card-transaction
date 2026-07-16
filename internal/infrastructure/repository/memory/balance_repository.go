package memory

import (
	"card-transaction/internal/domain/vo"
)

type BalanceRepository struct {
	balances map[string]vo.Money
}

func NewBalanceRepository() BalanceRepository {
	return BalanceRepository{balances: map[string]vo.Money{}}
}

func (r BalanceRepository) GetBalance(accountID string) (vo.Money, error) {
	balance, ok := r.balances[accountID]
	if !ok {
		return vo.Zero(), nil
	}
	return balance, nil
}
