package decision

type Outcome struct {
	Code       string
	HTTPStatus int
	Message    string
}

var table = map[string]Outcome{
	"00": {Code: "00", HTTPStatus: 200, Message: "Operacao realizada com sucesso."},
	"01": {Code: "01", HTTPStatus: 400, Message: "Saldo insuficiente ou limite excedido."},
	"02": {Code: "02", HTTPStatus: 404, Message: "Cartao nao encontrado."},
	"03": {Code: "03", HTTPStatus: 400, Message: "Cartao bloqueado ou inativo."},
	"04": {Code: "04", HTTPStatus: 400, Message: "Valor invalido baseado na decisao do emissor."},
	"05": {Code: "05", HTTPStatus: 400, Message: "Suspeita de fraude, transacao negada."},
	"06": {Code: "06", HTTPStatus: 400, Message: "Produto nao permitido para este cartao."},
	"07": {Code: "07", HTTPStatus: 409, Message: "Operacao ja realizada."},
	"08": {Code: "08", HTTPStatus: 400, Message: "Valor da transacao excede o permitido."},
	"09": {Code: "09", HTTPStatus: 404, Message: "Transacao original nao encontrada."},
	"10": {Code: "10", HTTPStatus: 400, Message: "Identificadores de estorno nao encontrados ou valor de cancelamento excede o valor da compra."},
	"14": {Code: "14", HTTPStatus: 400, Message: "Cartao invalido."},
	"60": {Code: "60", HTTPStatus: 500, Message: "Erro ao processar transacao."},
	"96": {Code: "96", HTTPStatus: 500, Message: "Sistema indisponivel."},
}

func Resolve(code string) Outcome {
	if result, ok := table[code]; ok {
		return result
	}
	return table["96"]
}
