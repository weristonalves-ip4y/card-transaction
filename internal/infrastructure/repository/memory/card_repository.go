package memory

import "card-transaction/internal/domain/entity/card"

type CardRepository struct {
	cardsByPaysmartID map[string]card.Card
}

func NewCardRepository() CardRepository {
	return CardRepository{cardsByPaysmartID: map[string]card.Card{}}
}

func (r CardRepository) FindByPaysmartID(paysmartID string) (card.Card, error) {
	c, ok := r.cardsByPaysmartID[paysmartID]
	if !ok {
		return card.Card{}, nil
	}
	return c, nil
}
