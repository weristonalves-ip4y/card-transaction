package dto

import "fmt"

type AuthorizePurchaseRequest struct {
	PurchaseID                   string                 `json:"purchase_id"`
	AccountID                    string                 `json:"account_id"`
	PsProductCode                string                 `json:"psProductCode"`
	PsProductName                string                 `json:"psProductName"`
	ProductType                  string                 `json:"productType"`
	CountryCode                  string                 `json:"countryCode"`
	Source                       string                 `json:"source"`
	CallingSystemName            string                 `json:"callingSystemName"`
	PreAuthorization             bool                   `json:"preAuthorization"`
	IncrementalAuthorization     bool                   `json:"incrementalAuthorization"`
	Brand                        string                 `json:"brand"`
	Card                         CardInput              `json:"card"`
	TotalAmount                  TotalAmountInput       `json:"total_amount"`
	OriginalAmount               OriginalAmountInput    `json:"original_amount"`
	EntryMode                    string                 `json:"entry_mode"`
	ProcessingCode               ProcessingCodeInput    `json:"processing_code"`
	HolderValidationMode         string                 `json:"holder_validation_mode"`
	Internal                     bool                   `json:"internal"`
	Nrid                         string                 `json:"nrid"`
	Authentication3DSTransaction bool                   `json:"authentication3DSTransaction"`
	Platform                     string                 `json:"platform"`
	AdditionalTerminalData       AdditionalTerminalData `json:"additionalTerminalData"`
	Fees                         []FeesInput            `json:"fees"`
	OriginalIso8583              OriginalIso8583        `json:"original_iso8583"`
	Establishment                EstablishmentInput     `json:"establishment"`
	ForceAccept                  bool                   `json:"forceAccept"`
	Authorization                Authorization          `json:"authorization"`
	Internacional                bool                   `json:"internacional"`
}

type EstablishmentInput struct {
	Mcc     *string `json:"mcc"`
	Name    *string `json:"name"`
	City    *string `json:"city"`
	Address *string `json:"address"`
	State   *string `json:"state"`
	Country *string `json:"country"`
	ZipCode *string `json:"zipCode"`
	Pat     *bool   `json:"pat"`
}

type CardInput struct {
	PaysmartID string `json:"paysmart_id"`
	IssuerID   string `json:"issuer_id"`
	Pan        string `json:"pan"`
	PanSeq     string `json:"panseq	"`
	Bin        string `json:"bin"`
}

type TotalAmountInput struct {
	TotalAmount  int64 `json:"amount"`
	CurrencyCode int   `json:"currency_code"`
}

type OriginalAmountInput struct {
	OriginalAmount int64 `json:"amount"`
	CurrencyCode   int   `json:"currency_code"`
}

type ProcessingCodeInput struct {
	TipoTransacao          string  `json:"tipo_transacao"`
	SourceAccountType      string  `json:"source_account_type"`
	CreditCardAccount      *string `json:"credit_card_account"`
	DestinationAccountType string  `json:"destination_account_type"`
}

type FeesInput struct {
}

type AdditionalTerminalData struct {
	TerminalType                   *string `json:"terminalType"`
	PartialApprovalIndicator       *string `json:"partialApprovalIndicator"`
	TerminalLocationIndicator      *string `json:"terminalLocationIndicator"`
	CardholderPresenceIndicator    *string `json:"cardholderPresenceIndicator"`
	CardPresenceIndicator          *string `json:"cardPresenceIndicator"`
	CardCaptureCapabilityIndicator *string `json:"cardCaptureCapabilityIndicator"`
	TransactionStatusIndicator     *string `json:"transactionStatusIndicator"`
	TransactionSecurityIndicator   *string `json:"transactionSecurityIndicator"`
	TerminalPOSType                *string `json:"terminalPOSType"`
	TerminalInputCapability        *string `json:"terminalInputCapability"`
}

type Authorization struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type OriginalIso8583 struct {
	RequestMTI                               string  `json:"mti"`
	RequestCardNumber                        *string `json:"de002"`
	RequestProcessingCode                    *string `json:"de003"`
	RequestTransactionAmountLocal            *string `json:"de004"`
	RequestTransactionAmountReferencia       *string `json:"de005"`
	RequestAmountInCardHolderBilling         *string `json:"de006"`
	RequestTransmitionDateAndTime            *string `json:"de007"`
	RequestAmountCardholderBillingFee        *string `json:"de008"`
	RequestConversionRateSettlement          *string `json:"de009"`
	RequestConvertionRateCardHolderBilling   *string `json:"de010"`
	RequestSystemTraceAuditNumber            *string `json:"de011"`
	RequestLocalTransactionTime              *string `json:"de012"`
	RequestLocalTransactionDate              *string `json:"de013"`
	RequestExpirationDate                    *string `json:"de014"`
	RequestSettlementDate                    *string `json:"de015"`
	RequestConversionDate                    *string `json:"de016"`
	RequestMcc                               *string `json:"de018"`
	RequestAcquiringInstitutionCountryCode   *string `json:"de019"`
	RequestPosEntryMode                      *string `json:"de022"`
	RequestCardSequenceNumber                *string `json:"de023"`
	RequestNetworkInternationalId            *string `json:"de024"`
	RequestPosConditionCode                  *string `json:"de025"`
	RequestPosPinCaptureCode                 *string `json:"de026"`
	RequestTransactionFeeAmount              *string `json:"de028"`
	RequestSettlementFeeAmount               *string `json:"de029"`
	RequestAquiringInstitutionCode           *string `json:"de032"`
	RequestForwardingInstitutionCode         *string `json:"de033"`
	RequestTrack2Data                        *string `json:"de035"`
	RequestTrack3Data                        *string `json:"de036"`
	RequestRetrievalReferenceNumber          *string `json:"de037"`
	RequestAuthorizationIdResponse           *string `json:"de038"`
	RequestResponseCode                      *string `json:"de039"`
	RequestCardAcceptorTerminal              *string `json:"de041"`
	RequestCardAcceptorIdentificationCode    *string `json:"de042"`
	RequestCardAcceptorNameLocation          *string `json:"de043"`
	RequestTrack1Data                        *string `json:"de045"`
	RequestAdditionalDataIso                 *string `json:"de046"`
	RequestAdditionalDataNational            *string `json:"de047"`
	RequestContainsPdsInLtvFormat            *string `json:"de048"`
	RequestTransactionCurrencyCode           *string `json:"de049"`
	RequestTransactionAmount                 *string `json:"de050"`
	RequestCurrencyCodeCardholderBilling     *string `json:"de051"`
	RequestPersonalIdentificationNumberData  *string `json:"de052"`
	RequestSecurityRelatedControlInformation *string `json:"de053"`
	RequestAdditionalAmounts                 *string `json:"de054"`
	RequestIccSystemRelatedData              *string `json:"de055"`
	RequestOriginalDataElements              *string `json:"de056"`
	RequestAuthorizationLifeCycleCode        *string `json:"de058"`
	RequestAuthorizingAgentIdentifierVoucher *string `json:"de063"`
	RequestAuthorizationAddressVoucher       *string `json:"de122"`
}

type AuthorizeResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (r AuthorizePurchaseRequest) Validate() error {
	if r.PurchaseID == "" {
		return fmt.Errorf("purchase_id is required")
	}
	if r.AccountID == "" {
		return fmt.Errorf("account_id is required")
	}
	if r.PsProductCode == "" {
		return fmt.Errorf("ps_product_code is required")
	}
	if r.PsProductName == "" {
		return fmt.Errorf("ps_product_name is required")
	}
	if r.ProductType == "" {
		return fmt.Errorf("ps_product_type is required")
	}
	if r.CountryCode == "" {
		return fmt.Errorf("country_code is required")
	}
	if r.Source == "" {
		return fmt.Errorf("source is required")
	}
	if r.CallingSystemName == "" {
		return fmt.Errorf("calling_system_name is required")
	}
	if r.Card.PaysmartID == "" {
		return fmt.Errorf("card.paysmart_id is required")
	}
	if r.Card.IssuerID == "" {
		return fmt.Errorf("card.issuer_id is required")
	}
	if r.Card.Pan == "" {
		return fmt.Errorf("card.pan is required")
	}
	if r.Card.PanSeq == "" {
		return fmt.Errorf("card.pan_seq is required")
	}
	if r.TotalAmount.TotalAmount <= 0 {
		return fmt.Errorf("total_amount.amount must be greater than 0")
	}
	if r.TotalAmount.CurrencyCode <= 0 {
		return fmt.Errorf("total_amount.currency_code must be greater than 0")
	}
	if r.OriginalAmount.OriginalAmount <= 0 {
		return fmt.Errorf("original_amount.amount must be greater than 0")
	}
	if r.OriginalAmount.CurrencyCode <= 0 {
		return fmt.Errorf("original_amount.currency_code must be greater than 0")
	}
	if r.ProcessingCode.TipoTransacao == "" {
		return fmt.Errorf("processing_code.tipo_transacao is required")
	}
	if r.ProcessingCode.DestinationAccountType == "" {
		return fmt.Errorf("processing_code.destination_account_type is required")
	}
	if r.Authorization.Code == "" {
		return fmt.Errorf("authorization.code is required")
	}
	if r.Authorization.Description == "" {
		return fmt.Errorf("authorization.description is required")
	}

	return nil
}
