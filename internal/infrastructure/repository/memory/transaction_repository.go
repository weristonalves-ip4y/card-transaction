package memory

import "card-transaction/internal/domain/vo"

type TransactionRepository struct {
	duplicates map[string]bool
	monthlySum map[int64]vo.Money
	serialized []map[string]any
}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		duplicates: map[string]bool{},
		monthlySum: map[int64]vo.Money{},
		serialized: []map[string]any{},
	}
}

func (r *TransactionRepository) ExistsByIdentifier(identifier string) (bool, error) {
	return r.duplicates[identifier], nil
}

func (r *TransactionRepository) GetMonthlySum(cardID int64) (vo.Money, error) {
	sum, ok := r.monthlySum[cardID]
	if !ok {
		return vo.Zero(), nil
	}
	return sum, nil
}

func (r *TransactionRepository) SaveSerialized(payload map[string]any) (int64, error) {
	r.serialized = append(r.serialized, payload)
	return int64(len(r.serialized)), nil
}
