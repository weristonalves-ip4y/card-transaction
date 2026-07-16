Testes e Mocks em Go
Leia este arquivo ao gerar ou revisar testes em Go.
Princípios

Table-driven tests são o padrão idiomático para cobrir múltiplos cenários sem duplicar código de teste.
Testar comportamento público, não implementação interna — testes não devem quebrar ao refatorar o interior de uma função sem mudar seu contrato.
Sempre cobrir caminho de erro, não só o caminho feliz.
Dependências externas (DB, HTTP client, fila) são abstraídas por interface no domínio/aplicação e mockadas no teste — isso é o principal motivo de "aceitar interfaces, retornar structs".
Testes de integração (batendo em DB/serviço real, ex. via testcontainers-go) ficam separados dos testes unitários — usar build tag (//go:build integration) ou nomenclatura clara (_integration_test.go).

Table-driven test — exemplo
gofunc TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name    string
        amount  int64
        tier    string
        want    int64
        wantErr bool
    }{
        {"gold tier gets 10% off", 10000, "gold", 9000, false},
        {"no tier no discount", 10000, "", 10000, false},
        {"negative amount is invalid", -100, "gold", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateDiscount(tt.amount, tt.tier)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
Mocks via interface (manual ou com testify/mock)
Definir a interface no pacote que consome (aplicação/use case), gerar/mockar a implementação para teste.
go// application/order/create_order.go
type OrderRepository interface {
    Save(ctx context.Context, o *domain.Order) error
}

type PaymentGateway interface {
    Charge(ctx context.Context, amount int64) (string, error)
}
Com testify/mock:
gotype MockPaymentGateway struct {
    mock.Mock
}

func (m *MockPaymentGateway) Charge(ctx context.Context, amount int64) (string, error) {
    args := m.Called(ctx, amount)
    return args.String(0), args.Error(1)
}

func TestCreateOrder_PaymentFails(t *testing.T) {
    mockGateway := new(MockPaymentGateway)
    mockGateway.On("Charge", mock.Anything, int64(5000)).
        Return("", errors.New("gateway timeout"))

    uc := NewCreateOrderUseCase(mockRepo, mockGateway)
    _, err := uc.Execute(context.Background(), CreateOrderInput{AmountCents: 5000})

    require.Error(t, err)
    mockGateway.AssertExpectations(t)
}
Alternativa sem dependência externa: mocks manuais com structs implementando a interface e campos-função (útil em projetos que evitam dependências extras):
gotype stubPaymentGateway struct {
    chargeFn func(ctx context.Context, amount int64) (string, error)
}

func (s *stubPaymentGateway) Charge(ctx context.Context, amount int64) (string, error) {
    return s.chargeFn(ctx, amount)
}
Geração automática de mocks
Para interfaces maiores, usar mockery (//go:generate mockery --name=OrderRepository) para gerar mocks consistentes automaticamente em vez de escrever tudo à mão — reduz erro humano em projetos grandes.
Testando handlers HTTP com httptest
gofunc TestCreateOrderHandler(t *testing.T) {
    body := `{"customer_id":"...", "amount_cents":5000, "currency":"BRL"}`
    req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    handler := NewOrderHandler(mockUseCase)
    handler.Create(w, req)

    resp := w.Result()
    require.Equal(t, http.StatusCreated, resp.StatusCode)
}
Testando concorrência

Rodar testes suspeitos de race condition com go test -race.
Para código com goroutines/channels, usar sync.WaitGroup no teste para aguardar conclusão antes de assertar, e nunca depender de time.Sleep para sincronização (flaky).

O que apontar em revisão quando testes estão fracos

Só testa caminho feliz, nenhum teste de erro.
Mocka dependências concretas em vez de depender de interface (sinal de que a abstração está no lugar errado).
Testes que dependem de ordem de execução ou estado global compartilhado entre eles.
Ausência de -race no CI para pacotes com concorrência.
Cobertura alta em número mas testando getters/setters triviais em vez de regra de negócio.