package purchase

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainnotification "card-transaction/internal/domain/notification"
	"card-transaction/internal/domain/vo"
)

const smsDateTimeLayout = "02/01/2006 15:04:05"

func notifyApprovedPurchaseSMS(
	dispatcher domainnotification.Dispatcher,
	phone string,
	amount vo.Money,
	establishmentName string,
	cardFinalDigits string,
	transactionDate string,
) {
	if dispatcher == nil {
		return
	}

	if phone == "" {
		return
	}

	dispatcher.SendSMS(context.Background(), phone, buildApprovedPurchaseSMSMessage(establishmentName, amount, cardFinalDigits, transactionDate))
}

func resolveNotificationPhone(phone string) string {
	if phone == "" {
		return ""
	}

	return strings.TrimSpace(phone)
}

func buildApprovedPurchaseSMSMessage(establishmentName string, amount vo.Money, cardFinalDigits string, transactionDate string) string {
	return fmt.Sprintf(
		"Transação aprovada no valor de R$ %.2f em %s. Cartão final %s. em %s.",
		amount.ToFloat(),
		establishmentName,
		cardFinalDigits,
		transactionDate,
	)
}

func formatSMSDateTime(transactionTime time.Time) string {
	// In Go, this is a formatting layout template, not a fixed date value.
	return transactionTime.Format(smsDateTimeLayout)
}
