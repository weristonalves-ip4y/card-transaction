package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"card-transaction/internal/domain/vo"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return TransactionRepository{db: db}
}

func (r TransactionRepository) ExistsByIdentifier(identifier string) (bool, error) {
	const query = `
		SELECT TOP 1 1
		FROM transaction_purchases
		WHERE purchase_id = @identifier
	`

	var found int
	err := r.db.QueryRowContext(context.Background(), query, sql.Named("identifier", identifier)).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("querying duplicate purchase_id %s: %w", identifier, err)
	}

	return found == 1, nil
}

func (r TransactionRepository) GetMonthlySum(accountID int64) (vo.Money, error) {
	const query = `
		SELECT ISNULL(SUM(request_transaction_amount_local), 0)
		FROM transaction_purchases
		WHERE account_id = @accountID
			AND purchase_id IS NOT NULL
			AND response_code = '00'
			AND deleted_at IS NULL
			AND created_at >= DATEFROMPARTS(YEAR(GETDATE()), MONTH(GETDATE()), 1)
	`

	var cents int64
	err := r.db.QueryRowContext(context.Background(), query, sql.Named("accountID", accountID)).Scan(&cents)
	if err != nil {
		return vo.Money{}, fmt.Errorf("querying monthly sum for account %d: %w", accountID, err)
	}

	total, err := vo.NewFromCents(cents)
	if err != nil {
		return vo.Money{}, fmt.Errorf("invalid monthly sum for account %d: %w", accountID, err)
	}

	return total, nil
}

func (r TransactionRepository) SaveSerialized(payload map[string]any) (int64, error) {
	const query = `
		INSERT INTO transaction_purchases (
			uuid,
			account_id,
			card_id,
			transfer_id,
			withdrawal_id,
			purchase_id,
			request_mti,
			request_card_number,
			request_processing_code,
			request_transaction_amount_local_original,
			request_transaction_amount_local,
			request_transaction_amount_referencia,
			request_amount_in_card_holder_billing,
			request_transmition_date_and_time,
			request_convertion_rate_card_holder_billing,
			request_system_trace_audit_number,
			request_local_transaction_time,
			request_local_transaction_date,
			request_expiration_date,
			request_mcc,
			request_acquiring_institution_country_code,
			request_pos_entry_mode,
			request_pos_condition_code,
			request_aquiring_institution_code,
			request_retrieval_reference_number,
			request_authorization_response_code,
			request_card_acceptor_terminal,
			request_card_acceptor_identification_code,
			request_card_acceptor_name_location,
			request_contains_pds_in_ltv_format,
			request_transaction_currency_code,
			request_transaction_amount,
			request_currency_code_cardholder_billing,
			response_mti,
			response_card_number,
			response_processing_code,
			response_transaction_amount_local,
			response_amount_in_card_holder_billing,
			response_transmition_date_and_time,
			response_conversion_rate,
			response_system_trace_audit_number,
			response_acquiring_institution_country_code,
			response_pos_condition_code,
			response_aquiring_institution_code,
			response_retrieval_reference_number,
			response_authorization_identification_response,
			response_code,
			response_card_acceptor_terminal,
			response_card_acceptor_identification_code,
			response_card_acceptor_name_location,
			response_transaction_currency_code,
			response_currency_code_cardholder_billing,
			transaction_justification,
			transaction_justification_approved,
			transfer_data_payment_type,
			transfer_data_unique_reference_number,
			transfer_data_senders_name,
			transfer_data_senders_address,
			transfer_data_senders_city,
			transfer_data_senders_country_state_code_if_us,
			transfer_data_cardholder_zipcode,
			transfer_data_cardholder_identification_number,
			transfer_data_origin_of_funds,
			transfer_data_additional_transfer_data,
			transfer_data_recipient_code,
			transfer_data_fund_sender_email,
			transfer_data_fund_recipient_email,
			transfer_data_fund_sender_phone,
			transfer_data_fund_recipient_phone,
			transfer_data_device_id,
			transfer_data_cardholder_cpf_or_cnpj,
			transfer_data_bin_origin,
			transfer_data_origin_card_last_4_digits,
			transfer_data_transaction_type,
			pat,
			created_at,
			updated_at
		)
		OUTPUT INSERTED.id
		VALUES (
			@uuid,
			@accountID,
			@cardID,
			@transferID,
			@withdrawalID,
			@purchaseID,
			@requestMTI,
			@requestCardNumber,
			@requestProcessingCode,
			@requestAmountLocalOriginal,
			@requestAmountLocal,
			@requestAmountReferencia,
			@requestAmountBilling,
			@requestTransmissionDateTime,
			@requestConversionRate,
			@requestSTAN,
			@requestLocalTime,
			@requestLocalDate,
			@requestExpirationDate,
			@requestMCC,
			@requestCountryCode,
			@requestPOSEntryMode,
			@requestPOSConditionCode,
			@requestAquiringInstitutionCode,
			@requestRetrievalRef,
			@requestAuthorizationResponseCode,
			@requestAcceptorTerminal,
			@requestAcceptorIdentificationCode,
			@requestAcceptorNameLocation,
			@requestContainsPDS,
			@requestTransactionCurrencyCode,
			@requestTransactionAmount,
			@requestCurrencyCodeBilling,
			@responseMTI,
			@responseCardNumber,
			@responseProcessingCode,
			@responseAmountLocal,
			@responseAmountBilling,
			@responseTransmissionDateTime,
			@responseConversionRate,
			@responseSTAN,
			@responseCountryCode,
			@responsePOSConditionCode,
			@responseAquiringInstitutionCode,
			@responseRetrievalRef,
			@responseAuthorizationID,
			@responseCode,
			@responseAcceptorTerminal,
			@responseAcceptorIdentificationCode,
			@responseAcceptorNameLocation,
			@responseTransactionCurrencyCode,
			@responseCurrencyCodeBilling,
			@transactionJustification,
			@transactionJustificationApproved,
			@transferDataPaymentType,
			@transferDataUniqueReferenceNumber,
			@transferDataSendersName,
			@transferDataSendersAddress,
			@transferDataSendersCity,
			@transferDataSendersCountryStateCodeIfUS,
			@transferDataCardholderZipcode,
			@transferDataCardholderIdentificationNumber,
			@transferDataOriginOfFunds,
			@transferDataAdditionalTransferData,
			@transferDataRecipientCode,
			@transferDataFundSenderEmail,
			@transferDataFundRecipientEmail,
			@transferDataFundSenderPhone,
			@transferDataFundRecipientPhone,
			@transferDataDeviceID,
			@transferDataCardholderCpfOrCnpj,
			@transferDataBinOrigin,
			@transferDataOriginCardLast4Digits,
			@transferDataTransactionType,
			@pat,
			@createdAt,
			@updatedAt
		)
	`

	now := time.Now().UTC()

	var insertedID int64
	err := r.db.QueryRowContext(
		context.Background(),
		query,
		sql.Named("uuid", getString(payload, "uuid")),
		sql.Named("accountID", getNullableInt64(payload, "account_id")),
		sql.Named("cardID", getNullableInt64(payload, "card_id")),
		sql.Named("transferID", getNullableString(payload, "transfer_id")),
		sql.Named("withdrawalID", getNullableString(payload, "withdrawal_id")),
		sql.Named("purchaseID", getNullableString(payload, "purchase_id")),
		sql.Named("requestMTI", getString(payload, "request_mti")),
		sql.Named("requestCardNumber", getString(payload, "request_card_number")),
		sql.Named("requestProcessingCode", getString(payload, "request_processing_code")),
		sql.Named("requestAmountLocalOriginal", getInt64(payload, "request_transaction_amount_local_original")),
		sql.Named("requestAmountLocal", getInt64(payload, "request_transaction_amount_local")),
		sql.Named("requestAmountReferencia", getInt64(payload, "request_transaction_amount_referencia")),
		sql.Named("requestAmountBilling", getInt64(payload, "request_amount_in_card_holder_billing")),
		sql.Named("requestTransmissionDateTime", now),
		sql.Named("requestConversionRate", getFloat64(payload, "request_convertion_rate_card_holder_billing")),
		sql.Named("requestSTAN", getString(payload, "request_system_trace_audit_number")),
		sql.Named("requestLocalTime", getString(payload, "request_local_transaction_time")),
		sql.Named("requestLocalDate", getString(payload, "request_local_transaction_date")),
		sql.Named("requestExpirationDate", getString(payload, "request_expiration_date")),
		sql.Named("requestMCC", getString(payload, "request_mcc")),
		sql.Named("requestCountryCode", getString(payload, "request_acquiring_institution_country_code")),
		sql.Named("requestPOSEntryMode", getString(payload, "request_pos_entry_mode")),
		sql.Named("requestPOSConditionCode", getString(payload, "request_pos_condition_code")),
		sql.Named("requestAquiringInstitutionCode", getString(payload, "request_aquiring_institution_code")),
		sql.Named("requestRetrievalRef", getString(payload, "request_retrieval_reference_number")),
		sql.Named("requestAuthorizationResponseCode", getString(payload, "request_authorization_response_code")),
		sql.Named("requestAcceptorTerminal", getString(payload, "request_card_acceptor_terminal")),
		sql.Named("requestAcceptorIdentificationCode", getString(payload, "request_card_acceptor_identification_code")),
		sql.Named("requestAcceptorNameLocation", getString(payload, "request_card_acceptor_name_location")),
		sql.Named("requestContainsPDS", getBool(payload, "request_contains_pds_in_ltv_format")),
		sql.Named("requestTransactionCurrencyCode", getString(payload, "request_transaction_currency_code")),
		sql.Named("requestTransactionAmount", getInt64(payload, "request_transaction_amount")),
		sql.Named("requestCurrencyCodeBilling", getString(payload, "request_currency_code_cardholder_billing")),
		sql.Named("responseMTI", getString(payload, "response_mti")),
		sql.Named("responseCardNumber", getString(payload, "response_card_number")),
		sql.Named("responseProcessingCode", getString(payload, "response_processing_code")),
		sql.Named("responseAmountLocal", getInt64(payload, "response_transaction_amount_local")),
		sql.Named("responseAmountBilling", getInt64(payload, "response_amount_in_card_holder_billing")),
		sql.Named("responseTransmissionDateTime", now),
		sql.Named("responseConversionRate", getFloat64(payload, "response_conversion_rate")),
		sql.Named("responseSTAN", getString(payload, "response_system_trace_audit_number")),
		sql.Named("responseCountryCode", getString(payload, "response_acquiring_institution_country_code")),
		sql.Named("responsePOSConditionCode", getString(payload, "response_pos_condition_code")),
		sql.Named("responseAquiringInstitutionCode", getString(payload, "response_aquiring_institution_code")),
		sql.Named("responseRetrievalRef", getString(payload, "response_retrieval_reference_number")),
		sql.Named("responseAuthorizationID", getNullableString(payload, "response_authorization_identification_response")),
		sql.Named("responseCode", getString(payload, "response_code")),
		sql.Named("responseAcceptorTerminal", getString(payload, "response_card_acceptor_terminal")),
		sql.Named("responseAcceptorIdentificationCode", getString(payload, "response_card_acceptor_identification_code")),
		sql.Named("responseAcceptorNameLocation", getString(payload, "response_card_acceptor_name_location")),
		sql.Named("responseTransactionCurrencyCode", getString(payload, "response_transaction_currency_code")),
		sql.Named("responseCurrencyCodeBilling", getString(payload, "response_currency_code_cardholder_billing")),
		sql.Named("transactionJustification", ""),
		sql.Named("transactionJustificationApproved", false),
		sql.Named("transferDataPaymentType", ""),
		sql.Named("transferDataUniqueReferenceNumber", ""),
		sql.Named("transferDataSendersName", ""),
		sql.Named("transferDataSendersAddress", ""),
		sql.Named("transferDataSendersCity", ""),
		sql.Named("transferDataSendersCountryStateCodeIfUS", ""),
		sql.Named("transferDataCardholderZipcode", ""),
		sql.Named("transferDataCardholderIdentificationNumber", ""),
		sql.Named("transferDataOriginOfFunds", ""),
		sql.Named("transferDataAdditionalTransferData", ""),
		sql.Named("transferDataRecipientCode", ""),
		sql.Named("transferDataFundSenderEmail", ""),
		sql.Named("transferDataFundRecipientEmail", ""),
		sql.Named("transferDataFundSenderPhone", ""),
		sql.Named("transferDataFundRecipientPhone", ""),
		sql.Named("transferDataDeviceID", ""),
		sql.Named("transferDataCardholderCpfOrCnpj", ""),
		sql.Named("transferDataBinOrigin", ""),
		sql.Named("transferDataOriginCardLast4Digits", ""),
		sql.Named("transferDataTransactionType", ""),
		sql.Named("pat", getNullableBool(payload, "pat")),
		sql.Named("createdAt", now),
		sql.Named("updatedAt", now),
	).Scan(&insertedID)
	if err != nil {
		return 0, fmt.Errorf("inserting transaction_purchases row: %w", err)
	}

	return insertedID, nil
}

func getString(m map[string]any, key string) string {
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func getNullableString(m map[string]any, key string) any {
	value, ok := m[key]
	if !ok || value == nil {
		return nil
	}
	if str, ok := value.(string); ok {
		if str == "" {
			return nil
		}
		return str
	}
	return fmt.Sprintf("%v", value)
}

func getNullableInt64(m map[string]any, key string) any {
	value, ok := m[key]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case *int64:
		if typed == nil {
			return nil
		}
		return *typed
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return nil
		}
		return parsed
	default:
		return nil
	}
}

func getInt64(m map[string]any, key string) int64 {
	value, ok := m[key]
	if !ok || value == nil {
		return 0
	}

	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func getFloat64(m map[string]any, key string) float64 {
	value, ok := m[key]
	if !ok || value == nil {
		return 0
	}

	switch typed := value.(type) {
	case float32:
		return float64(typed)
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func getBool(m map[string]any, key string) bool {
	value, ok := m[key]
	if !ok || value == nil {
		return false
	}

	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
		return typed == "1"
	default:
		return false
	}
}

func getNullableBool(m map[string]any, key string) any {
	value, ok := m[key]
	if !ok || value == nil {
		return nil
	}
	if typed, ok := value.(*bool); ok {
		if typed == nil {
			return nil
		}
		return *typed
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	return nil
}
