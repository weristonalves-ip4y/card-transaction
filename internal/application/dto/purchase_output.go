package dto

type PurchaseOutput struct {
	StatusCode int
	Data       map[string]any
}

func PurchaseOutputFromAuthorizationCode(code string, message string) PurchaseOutput {
	spec, ok := purchaseOutputSpecs[code]
	if !ok {
		spec = purchaseOutputSpecs["96"]
		code = "96"
	}

	responseMessage := spec.Message
	if responseMessage == "" {
		responseMessage = message
	}

	data := map[string]any{
		"message":            responseMessage,
		"code":               spec.Code,
		"authorization_code": code,
	}

	if spec.IncludeBalance {
		data["balance"] = map[string]any{
			"amount":        nil,
			"currency_code": nil,
		}
	}

	if spec.IncludeAuthorizationID {
		data["authorization_id"] = nil
	}

	if spec.IncludeUseVoucher {
		data["useVoucher"] = false
	}

	if spec.IncludePurchaseFields {
		data["purchaseOnlyApproval"] = nil
		data["purchaseOnlyPartialAmountApproved"] = nil
		data["cashbackOnlyPartialAmountApproved"] = nil
	}

	return PurchaseOutput{StatusCode: spec.HTTPStatus, Data: data}
}

type purchaseOutputSpec struct {
	HTTPStatus             int
	Code                   int
	Message                string
	IncludeBalance         bool
	IncludeAuthorizationID bool
	IncludeUseVoucher      bool
	IncludePurchaseFields  bool
}

var purchaseOutputSpecs = map[string]purchaseOutputSpec{
	"00": {
		HTTPStatus:             200,
		Code:                   0,
		Message:                "Operacao realizada com sucesso.",
		IncludeBalance:         true,
		IncludeAuthorizationID: true,
		IncludePurchaseFields:  true,
	},
	"01": {
		HTTPStatus:        400,
		Code:              530,
		Message:           "Saldo insuficiente ou limite excedido.",
		IncludeBalance:    true,
		IncludeUseVoucher: true,
	},
	"02": {HTTPStatus: 404, Code: 111, Message: "Cartao nao encontrado."},
	"03": {
		HTTPStatus:     400,
		Code:           530,
		Message:        "Cartao bloqueado ou inativo.",
		IncludeBalance: true,
	},
	"04": {HTTPStatus: 400, Code: 530, Message: "Valor invalido baseado na decisao do emissor."},
	"05": {HTTPStatus: 400, Code: 530, Message: "Suspeita de fraude, transacao negada."},
	"06": {HTTPStatus: 400, Code: 530, Message: "Produto nao permitido para este cartao."},
	"07": {HTTPStatus: 409, Code: 140, Message: "Operacao ja realizada."},
	"08": {HTTPStatus: 400, Code: 530, Message: "Valor da transacao excede o permitido."},
	"09": {HTTPStatus: 404, Code: 404, Message: "Transacao original nao encontrada."},
	"10": {
		HTTPStatus: 400,
		Code:       530,
		Message:    "Identificadores de estorno nao encontrados ou valor de cancelamento excede o valor da compra.",
	},
	"14": {HTTPStatus: 400, Code: 530, Message: "Cartao invalido."},
	"60": {HTTPStatus: 500, Code: 60, Message: "Erro ao processar transacao."},
	"96": {HTTPStatus: 500, Code: 900, Message: "Sistema indisponivel. Erro ao processar transacao."},
}
