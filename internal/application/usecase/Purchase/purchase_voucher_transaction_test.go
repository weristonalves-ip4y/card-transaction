package purchase

import (
	"testing"

	"card-transaction/internal/domain/entity/card"
	"card-transaction/internal/domain/vo"
)

type voucherMovementRepoSpy struct{}

func (s *voucherMovementRepoSpy) InsertDebitVoucherMovement(
	accountID,
	originID int64,
	movementTypeID int,
	amount float64,
	description string,
	cardID,
	cardExternalID,
	cardPurchaseMcc interface{},
) error {
	return nil
}

type voucherTxRepoSpy struct {
	existsByIdentifier bool
	saveCalls          int
	lastPayload        map[string]any
}

func (r *voucherTxRepoSpy) ExistsByIdentifier(identifier string) (bool, error) {
	return r.existsByIdentifier, nil
}

func (r *voucherTxRepoSpy) GetMonthlySum(cardID string) (vo.Money, error) {
	return vo.Zero(), nil
}

func (r *voucherTxRepoSpy) SaveSerialized(payload map[string]any) (int64, error) {
	r.saveCalls++
	r.lastPayload = payload
	return int64(r.saveCalls), nil
}

type voucherCardRepoStub struct{}

func (s voucherCardRepoStub) FindByPaysmartID(paysmartID string) (card.Card, error) {
	return card.Card{}, nil
}

type voucherBalanceRepoStub struct{}

func (s voucherBalanceRepoStub) GetBalanceVoucher(accountID int64) (vo.Money, error) {
	return vo.Zero(), nil
}

func TestPurchaseVoucherTransactionExecuteDuplicatePersistsRejectedTransaction(t *testing.T) {
	t.Parallel()

	txRepo := &voucherTxRepoSpy{existsByIdentifier: true}
	useCase := NewPurchaseVoucherTransaction(voucherCardRepoStub{}, txRepo, voucherBalanceRepoStub{}, &voucherMovementRepoSpy{})

	input := validAuthorizePurchaseRequest()
	input.PsProductCode = "011401"

	output, err := useCase.Execute(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if output.Approved {
		t.Fatalf("expected rejected output")
	}

	if output.Code != "07" {
		t.Fatalf("expected code 07, got %s", output.Code)
	}

	if txRepo.saveCalls != 1 {
		t.Fatalf("expected 1 persistence call, got %d", txRepo.saveCalls)
	}

	responseCode, ok := txRepo.lastPayload["response_code"].(string)
	if !ok {
		t.Fatalf("expected response_code field in payload")
	}

	if responseCode != "07" {
		t.Fatalf("expected persisted response_code 07, got %s", responseCode)
	}
}

func TestPurchaseVoucherTransactionExecuteDuplicateRepositoryErrorPersists96(t *testing.T) {
	t.Parallel()

	txRepo := &voucherTxRepoSpyWithError{}
	useCase := NewPurchaseVoucherTransaction(voucherCardRepoStub{}, txRepo, voucherBalanceRepoStub{}, &voucherMovementRepoSpy{})

	_, err := useCase.Execute(validAuthorizePurchaseRequest())
	if err == nil {
		t.Fatalf("expected error")
	}

	if txRepo.saveCalls != 1 {
		t.Fatalf("expected 1 persistence call, got %d", txRepo.saveCalls)
	}

	responseCode, ok := txRepo.lastPayload["response_code"].(string)
	if !ok {
		t.Fatalf("expected response_code field in payload")
	}

	if responseCode != "96" {
		t.Fatalf("expected persisted response_code 96, got %s", responseCode)
	}
}

type voucherTxRepoSpyWithError struct {
	saveCalls   int
	lastPayload map[string]any
}

func (r *voucherTxRepoSpyWithError) ExistsByIdentifier(identifier string) (bool, error) {
	return false, assertErr{}
}

func (r *voucherTxRepoSpyWithError) GetMonthlySum(cardID string) (vo.Money, error) {
	return vo.Zero(), nil
}

func (r *voucherTxRepoSpyWithError) SaveSerialized(payload map[string]any) (int64, error) {
	r.saveCalls++
	r.lastPayload = payload
	return int64(r.saveCalls), nil
}

type assertErr struct{}

func (e assertErr) Error() string { return "repository failure" }

var _ error = assertErr{}

var _ CardRepository = voucherCardRepoStub{}
var _ BalanceVoucherRepository = voucherBalanceRepoStub{}
var _ CardVoucherMovementRepository = &voucherMovementRepoSpy{}
var _ TransactionRepository = (*voucherTxRepoSpy)(nil)
var _ TransactionRepository = (*voucherTxRepoSpyWithError)(nil)
