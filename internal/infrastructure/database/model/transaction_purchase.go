package model

import "time"

type TransactionPurchase struct {
	ID                                          int64      `db:"id"`
	UUID                                        string     `db:"uuid"`
	AccountID                                   string     `db:"account_id"`
	CardID                                      string     `db:"card_id"`
	RequestMTI                                  string     `db:"request_mti"`
	TransferID                                  *string    `db:"transfer_id"`
	WithdrawalID                                *string    `db:"withdrawal_id"`
	PurchaseID                                  *string    `db:"purchase_id"`
	RequestCardNumber                           string     `db:"request_card_number"`
	RequestProcessingCode                       string     `db:"request_processing_code"`
	RequestTransactionAmountLocalOriginal       int64      `db:"request_transaction_amount_local_original"`
	RequestTransactionAmountLocal               int64      `db:"request_transaction_amount_local"`
	RequestTransactionAmountReferencia          int64      `db:"request_transaction_amount_referencia"`
	RequestAmountInCardHolderBilling            int64      `db:"request_amount_in_card_holder_billing"`
	RequestTransmitionDateAndTime               time.Time  `db:"request_transmition_date_and_time"`
	RequestConvertionRateCardHolderBilling      float64    `db:"request_convertion_rate_card_holder_billing"`
	RequestSystemTraceAuditNumber               string     `db:"request_system_trace_audit_number"`
	RequestLocalTransactionTime                 string     `db:"request_local_transaction_time"`
	RequestLocalTransactionDate                 string     `db:"request_local_transaction_date"`
	RequestExpirationDate                       string     `db:"request_expiration_date"`
	RequestMCC                                  string     `db:"request_mcc"`
	RequestAcquiringInstitutionCountryCode      string     `db:"request_acquiring_institution_country_code"`
	RequestPOSEntryMode                         string     `db:"request_pos_entry_mode"`
	RequestPOSConditionCode                     string     `db:"request_pos_condition_code"`
	RequestAquiringInstitutionCode              string     `db:"request_aquiring_institution_code"`
	RequestRetrievalReferenceNumber             string     `db:"request_retrieval_reference_number"`
	RequestAuthorizationResponseCode            string     `db:"request_authorization_response_code"`
	RequestCardAcceptorTerminal                 string     `db:"request_card_acceptor_terminal"`
	RequestCardAcceptorIdentificationCode       string     `db:"request_card_acceptor_identification_code"`
	RequestCardAcceptorNameLocation             string     `db:"request_card_acceptor_name_location"`
	RequestContainsPDSInLTVFormat               bool       `db:"request_contains_pds_in_ltv_format"`
	RequestTransactionCurrencyCode              string     `db:"request_transaction_currency_code"`
	RequestTransactionAmount                    int64      `db:"request_transaction_amount"`
	RequestCurrencyCodeCardholderBilling        string     `db:"request_currency_code_cardholder_billing"`
	ResponseMTI                                 string     `db:"response_mti"`
	ResponseCardNumber                          string     `db:"response_card_number"`
	ResponseProcessingCode                      string     `db:"response_processing_code"`
	ResponseTransactionAmountLocal              int64      `db:"response_transaction_amount_local"`
	ResponseAmountInCardHolderBilling           int64      `db:"response_amount_in_card_holder_billing"`
	ResponseTransmitionDateAndTime              time.Time  `db:"response_transmition_date_and_time"`
	ResponseConversionRate                      float64    `db:"response_conversion_rate"`
	ResponseSystemTraceAuditNumber              string     `db:"response_system_trace_audit_number"`
	ResponseAcquiringInstitutionCountryCode     string     `db:"response_acquiring_institution_country_code"`
	ResponsePOSConditionCode                    string     `db:"response_pos_condition_code"`
	ResponseAquiringInstitutionCode             string     `db:"response_aquiring_institution_code"`
	ResponseRetrievalReferenceNumber            string     `db:"response_retrieval_reference_number"`
	ResponseAuthorizationIdentificationResponse string     `db:"response_authorization_identification_response"`
	ResponseCode                                string     `db:"response_code"`
	ResponseCardAcceptorTerminal                string     `db:"response_card_acceptor_terminal"`
	ResponseCardAcceptorIdentificationCode      string     `db:"response_card_acceptor_identification_code"`
	ResponseCardAcceptorNameLocation            string     `db:"response_card_acceptor_name_location"`
	ResponseTransactionCurrencyCode             string     `db:"response_transaction_currency_code"`
	ResponseCurrencyCodeCardholderBilling       string     `db:"response_currency_code_cardholder_billing"`
	TransactionJustification                    string     `db:"transaction_justification"`
	TransactionJustificationApproved            bool       `db:"transaction_justification_approved"`
	TransactionJustificationApprovedAt          *time.Time `db:"transaction_justification_approved_at"`
	TransactionJustificationApprovedByUserID    *string    `db:"transaction_justification_approved_by_user_id"`
	TransferDataPaymentType                     string     `db:"transfer_data_payment_type"`
	TransferDataUniqueReferenceNumber           string     `db:"transfer_data_unique_reference_number"`
	TransferDataSendersName                     string     `db:"transfer_data_senders_name"`
	TransferDataSendersAddress                  string     `db:"transfer_data_senders_address"`
	TransferDataSendersCity                     string     `db:"transfer_data_senders_city"`
	TransferDataSendersCountryStateCodeIfUS     string     `db:"transfer_data_senders_country_state_code_if_us"`
	TransferDataCardholderZipcode               string     `db:"transfer_data_cardholder_zipcode"`
	TransferDataCardholderIdentificationNumber  string     `db:"transfer_data_cardholder_identification_number"`
	TransferDataOriginOfFunds                   string     `db:"transfer_data_origin_of_funds"`
	TransferDataAdditionalTransferData          string     `db:"transfer_data_additional_transfer_data"`
	TransferDataRecipientCode                   string     `db:"transfer_data_recipient_code"`
	TransferDataFundSenderEmail                 string     `db:"transfer_data_fund_sender_email"`
	TransferDataFundRecipientEmail              string     `db:"transfer_data_fund_recipient_email"`
	TransferDataFundSenderPhone                 string     `db:"transfer_data_fund_sender_phone"`
	TransferDataFundRecipientPhone              string     `db:"transfer_data_fund_recipient_phone"`
	TransferDataDeviceID                        string     `db:"transfer_data_device_id"`
	TransferDataCardholderCpfOrCnpj             string     `db:"transfer_data_cardholder_cpf_or_cnpj"`
	TransferDataBinOrigin                       string     `db:"transfer_data_bin_origin"`
	TransferDataOriginCardLast4Digits           string     `db:"transfer_data_origin_card_last_4_digits"`
	TransferDataTransactionType                 string     `db:"transfer_data_transaction_type"`
	Pat                                         *bool      `db:"pat"`
	CreatedAt                                   time.Time  `db:"created_at"`
	UpdatedAt                                   time.Time  `db:"updated_at"`
	DeletedAt                                   *time.Time `db:"deleted_at"`
}

func (TransactionPurchase) TableName() string {
	return "transaction_purchases"
}
