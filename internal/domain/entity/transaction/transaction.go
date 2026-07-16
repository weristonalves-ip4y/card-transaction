package transaction

import (
	"card-transaction/internal/domain/vo"
	"fmt"
	"strconv"
)

type Type string

type ProductType string

const (
	ProductCard    ProductType = "card"
	ProductVoucher ProductType = "voucher"
)

const (
	TypePurchase   Type = "PURCHASE"
	TypeWithdrawal Type = "WITHDRAWAL"
)

type ValidationResult struct {
	Approved bool
	Code     string
	Reason   string
}

type ProcessingCode struct {
	TransactionType        string
	SourceAccountType      string
	CreditCardAccount      *string
	DestinationAccountType string
}

type AdditionalTerminalData struct {
	TerminalType                   *string
	PartialApprovalIndicator       *string
	TerminalLocationIndicator      *string
	CardholderPresenceIndicator    *string
	CardPresenceIndicator          *string
	CardCaptureCapabilityIndicator *string
	TransactionStatusIndicator     *string
	TransactionSecurityIndicator   *string
	TerminalPOSType                *string
	TerminalInputCapability        *string
}

type ISORequest struct {
	MTI   string
	DE002 *string
	DE003 *string
	DE004 *string
	DE005 *string
	DE006 *string
	DE007 *string
	DE008 *string
	DE009 *string
	DE010 *string
	DE011 *string
	DE012 *string
	DE013 *string
	DE014 *string
	DE015 *string
	DE016 *string
	DE018 *string
	DE019 *string
	DE022 *string
	DE023 *string
	DE024 *string
	DE025 *string
	DE026 *string
	DE028 *string
	DE029 *string
	DE032 *string
	DE033 *string
	DE035 *string
	DE036 *string
	DE037 *string
	DE038 *string
	DE039 *string
	DE041 *string
	DE042 *string
	DE043 *string
	DE045 *string
	DE046 *string
	DE047 *string
	DE048 *string
	DE049 *string
	DE050 *string
	DE051 *string
	DE052 *string
	DE053 *string
	DE054 *string
	DE055 *string
	DE056 *string
	DE058 *string
	DE063 *string
	DE122 *string
}

type ISOResponse struct {
	MTI                                 string
	CardNumber                          string
	ProcessingCode                      string
	TransactionAmountLocal              int64
	AmountInCardHolderBilling           int64
	TransmissionDateAndTime             string
	ConversionRate                      string
	SystemTraceAuditNumber              string
	AcquiringInstitutionCountryCode     string
	POSConditionCode                    string
	AquiringInstitutionCode             string
	RetrievalReferenceNumber            string
	AuthorizationIdentificationResponse string
	Code                                string
	CardAcceptorTerminal                string
	CardAcceptorIdentificationCode      string
	CardAcceptorNameLocation            string
	TransactionCurrencyCode             string
	CurrencyCodeCardholderBilling       string
}

type AuthorizationResponse struct {
	Code        string
	Description string
	HTTPStatus  int
	ISO         ISOResponse
}

type CardData struct {
	PaysmartID string
	IssuerID   string
	Pan        string
	PanSeq     string
	Bin        string
}

type AmountData struct {
	Total                vo.Money
	TotalCurrencyCode    int
	Original             vo.Money
	OriginalCurrencyCode int
}

type AuthorizationData struct {
	Code        string
	Description string
}

type EstablishmentData struct {
	MCC     *string
	Name    *string
	City    *string
	Address *string
	State   *string
	Country *string
	ZipCode *string
	Pat     *bool
}

type AuthorizeData struct {
	PurchaseID                   string
	PsProductCode                string
	PsProductName                string
	ProductType                  string
	CountryCode                  string
	Source                       string
	CallingSystemName            string
	Brand                        string
	EntryMode                    string
	HolderValidationMode         string
	Internal                     bool
	PreAuthorization             bool
	IncrementalAuthorization     bool
	Nrid                         string
	Authentication3DSTransaction bool
	Platform                     string
	Internacional                bool
	AdditionalTerminalData       AdditionalTerminalData
	OriginalISO8583              ISORequest
	Card                         CardData
	Amount                       AmountData
	ProcessingCode               ProcessingCode
	Authorization                AuthorizationData
	Establishment                EstablishmentData
	ResolvedCardID               string
	ResolvedProductCode          string
}

func Approved() ValidationResult {
	return ValidationResult{Approved: true, Code: "00"}
}

func Rejected(code, reason string) ValidationResult {
	return ValidationResult{Approved: false, Code: code, Reason: reason}
}

type Transaction struct {
	ID          string
	Identifier  string
	UUID        string
	AccountID   *int64
	CardID      *string
	Type        Type
	ProductType ProductType
	Value       vo.Money
	ForceAccept bool
	Data        AuthorizeData
	Response    AuthorizationResponse
}

func New(
	identifier string,
	txType Type,
	productType ProductType,
	value vo.Money,
	forceAccept bool,
) (Transaction, error) {
	return Transaction{
		ID:          identifier,
		Identifier:  identifier,
		Type:        txType,
		ProductType: productType,
		Value:       value,
		ForceAccept: forceAccept,
	}, nil
}

func (t Transaction) ValidateBalance(balance vo.Money) ValidationResult {
	if t.ForceAccept {
		return Approved()
	}

	if balance.LessThan(t.Value) {
		return Rejected("01", "insufficient balance")
	}

	return Approved()
}

func (t Transaction) WithAuthorizeData(data AuthorizeData) Transaction {
	t.Data = data
	return t
}

func (t Transaction) WithResolvedCard(cardID string, productCode string) Transaction {
	t.CardID = &cardID
	t.Data.ResolvedCardID = cardID
	t.Data.ResolvedProductCode = productCode
	return t
}

func (t Transaction) WithPersistenceUUID(uuid string) Transaction {
	t.UUID = uuid
	return t
}

func (t Transaction) WithPersistenceBase(uuid string, cardExternalID string, purchaseID *string, withdrawalID *string) Transaction {
	t.UUID = uuid
	t.Data.Card.PaysmartID = cardExternalID
	if purchaseID != nil {
		t.Data.PurchaseID = *purchaseID
	}
	if withdrawalID != nil {
		t.Type = TypeWithdrawal
	}
	return t
}

func (t Transaction) WithResolvedAccountCard(accountID int64, cardID string) Transaction {
	t.AccountID = &accountID
	t.CardID = &cardID
	t.Data.ResolvedCardID = cardID
	return t
}

func (t Transaction) WithAuthorizationResponse(response AuthorizationResponse) Transaction {
	t.Response = response
	return t
}

func (t Transaction) BuildISOResponse(responseCode string) Transaction {
	request := t.Data.OriginalISO8583

	t.Response = AuthorizationResponse{
		Code: responseCode,
		ISO: ISOResponse{
			MTI:                                 "0110", //TODO: Quando for WithDrawal, o MTI deve ser 0210
			CardNumber:                          stringOrEmpty(request.DE002),
			ProcessingCode:                      stringOrEmpty(request.DE003),
			TransactionAmountLocal:              t.Data.Amount.Total.Cents(),
			AmountInCardHolderBilling:           parseISOAmount(request.DE006),
			TransmissionDateAndTime:             stringOrEmpty(request.DE007),
			ConversionRate:                      stringOrEmpty(request.DE010),
			SystemTraceAuditNumber:              stringOrEmpty(request.DE011),
			AcquiringInstitutionCountryCode:     stringOrEmpty(request.DE019),
			POSConditionCode:                    stringOrEmpty(request.DE025),
			AquiringInstitutionCode:             stringOrEmpty(request.DE032),
			RetrievalReferenceNumber:            stringOrEmpty(request.DE037),
			AuthorizationIdentificationResponse: "",
			Code:                                responseCode,
			CardAcceptorTerminal:                stringOrEmpty(request.DE041),
			CardAcceptorIdentificationCode:      stringOrEmpty(request.DE042),
			CardAcceptorNameLocation:            stringOrEmpty(request.DE043),
			TransactionCurrencyCode:             CurrencyCodeToString(t.Data.Amount.TotalCurrencyCode),
			CurrencyCodeCardholderBilling:       stringOrEmpty(request.DE051),
		},
	}

	return t
}

func (t Transaction) ToPersistenceMap() map[string]any {
	req := t.Data.OriginalISO8583
	res := t.Response.ISO
	purchaseID := t.Data.PurchaseID

	return map[string]any{
		"uuid":                    t.UUID,
		"account_id":              t.AccountID,
		"card_id":                 t.CardID,
		"card_external_id":        t.Data.Card.PaysmartID,
		"purchase_id":             purchaseID,
		"withdrawal_id":           nil,
		"transfer_id":             nil,
		"request_mti":             req.MTI,
		"request_card_number":     stringOrNil(req.DE002),
		"request_processing_code": stringOrNil(req.DE003),
		"request_transaction_amount_local_original":      t.Data.Amount.Total.Cents(),
		"request_transaction_amount_local":               t.Data.Amount.Total.ToFloat(),
		"request_transaction_amount_referencia":          parseISOAmount(req.DE005),
		"request_amount_in_card_holder_billing":          parseISOAmount(req.DE006),
		"request_transmition_date_and_time":              stringOrNil(req.DE007),
		"request_convertion_rate_card_holder_billing":    stringOrNil(req.DE010),
		"request_system_trace_audit_number":              stringOrNil(req.DE011),
		"request_local_transaction_time":                 stringOrNil(req.DE012),
		"request_local_transaction_date":                 stringOrNil(req.DE013),
		"request_expiration_date":                        stringOrNil(req.DE014),
		"request_mcc":                                    stringOrNil(req.DE018),
		"request_acquiring_institution_country_code":     stringOrNil(req.DE019),
		"request_pos_entry_mode":                         stringOrNil(req.DE022),
		"request_pos_condition_code":                     stringOrNil(req.DE025),
		"request_aquiring_institution_code":              stringOrNil(req.DE032),
		"request_retrieval_reference_number":             stringOrNil(req.DE037),
		"request_authorization_response_code":            BoolToISOFlag(t.Data.IncrementalAuthorization),
		"request_card_acceptor_terminal":                 stringOrNil(req.DE041),
		"request_card_acceptor_identification_code":      stringOrNil(req.DE042),
		"request_card_acceptor_name_location":            stringOrNil(req.DE043),
		"request_contains_pds_in_ltv_format":             stringOrNil(req.DE048),
		"request_transaction_currency_code":              stringOrNil(req.DE049),
		"request_transaction_amount":                     stringOrNil(req.DE050),
		"request_currency_code_cardholder_billing":       stringOrNil(req.DE051),
		"response_mti":                                   res.MTI,
		"response_card_number":                           res.CardNumber,
		"response_processing_code":                       res.ProcessingCode,
		"response_transaction_amount_local":              res.TransactionAmountLocal,
		"response_amount_in_card_holder_billing":         res.AmountInCardHolderBilling,
		"response_transmition_date_and_time":             res.TransmissionDateAndTime,
		"response_conversion_rate":                       res.ConversionRate,
		"response_system_trace_audit_number":             res.SystemTraceAuditNumber,
		"response_acquiring_institution_country_code":    res.AcquiringInstitutionCountryCode,
		"response_pos_condition_code":                    res.POSConditionCode,
		"response_aquiring_institution_code":             res.AquiringInstitutionCode,
		"response_retrieval_reference_number":            res.RetrievalReferenceNumber,
		"response_authorization_identification_response": nil,
		"response_code":                                  res.Code,
		"response_card_acceptor_terminal":                res.CardAcceptorTerminal,
		"response_card_acceptor_identification_code":     res.CardAcceptorIdentificationCode,
		"response_card_acceptor_name_location":           res.CardAcceptorNameLocation,
		"response_transaction_currency_code":             res.TransactionCurrencyCode,
		"response_currency_code_cardholder_billing":      res.CurrencyCodeCardholderBilling,
		"pat": t.Data.Establishment.Pat,
	}
}

func CurrencyCodeToString(code int) string {
	return fmt.Sprintf("%d", code)
}

func BoolToISOFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func parseISOAmount(value *string) int64 {
	if value == nil || *value == "" {
		return 0
	}
	amount, err := strconv.ParseInt(*value, 10, 64)
	if err != nil {
		return 0
	}
	return amount
}

func stringOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
