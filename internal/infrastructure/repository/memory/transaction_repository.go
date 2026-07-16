package memory

import "card-transaction/internal/domain/vo"

type TransactionRepository struct {
	duplicates map[string]bool
	monthlySum map[string]vo.Money
	serialized []map[string]any
}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		duplicates: map[string]bool{},
		monthlySum: map[string]vo.Money{},
		serialized: []map[string]any{},
	}
}

func (r *TransactionRepository) ExistsByIdentifier(identifier string) (bool, error) {
	return r.duplicates[identifier], nil
}

func (r *TransactionRepository) GetMonthlySum(cardID string) (vo.Money, error) {
	sum, ok := r.monthlySum[cardID]
	if !ok {
		return vo.Zero(), nil
	}
	return sum, nil
}

func (r *TransactionRepository) SaveSerialized(payload map[string]any) error {
	r.serialized = append(r.serialized, payload)
	return nil
}
