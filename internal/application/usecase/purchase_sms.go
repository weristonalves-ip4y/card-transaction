package usecase

import (
	"context"
	"fmt"
	"strings"

	domainnotification "card-transaction/internal/domain/notification"
	"card-transaction/internal/domain/vo"
)

func notifyApprovedPurchaseSMS(
	dispatcher domainnotification.Dispatcher,
	phone string,
	productLabel string,
	amount vo.Money,
	balance vo.Money,
) {
	if dispatcher == nil {
		return
	}

	if phone == "" {
		return
	}

	dispatcher.SendSMS(context.Background(), phone, buildApprovedPurchaseSMSMessage(productLabel, amount, balance))
}

func resolveNotificationPhone(phone string) string {
	if phone == "" {
		return ""
	}

	return strings.TrimSpace(phone)
}

func buildApprovedPurchaseSMSMessage(productLabel string, amount vo.Money, balance vo.Money) string {
	return fmt.Sprintf(
		"Compra %s aprovada. Valor: R$ %.2f. Saldo: R$ %.2f.",
		strings.ToUpper(strings.TrimSpace(productLabel)),
		amount.ToFloat(),
		balance.ToFloat(),
	)
}
