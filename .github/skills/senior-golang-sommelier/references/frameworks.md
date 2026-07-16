Frameworks e Stacks em Go
Leia este arquivo antes de gerar código específico de framework, ou quando o usuário perguntar qual escolher.
Esta skill não força uma stack única — adapta ao que o projeto/usuário já usa, ou recomenda com base no contexto (tamanho do time, necessidade de performance, familiaridade).
HTTP Routers
net/http puro (stdlib, Go 1.22+)
Desde Go 1.22, net/http.ServeMux suporta path params e métodos HTTP nativamente (mux.HandleFunc("POST /orders/{id}", handler)). Para projetos que querem zero dependência externa e máximo controle, essa é a opção mais idiomática hoje em dia — evite recomendar um router de terceiros só por hábito quando a stdlib já resolve.
gomux := http.NewServeMux()
mux.HandleFunc("GET /orders/{id}", handler.GetOrder)
mux.HandleFunc("POST /orders", handler.CreateOrder)

var h http.Handler = mux
h = middleware.Recovery(logger)(h)
h = middleware.RequestID(h)
Gin
Popular, rápido, boa ergonomia (c.ShouldBindJSON, grupos de rota, middlewares prontos). Bom para times que já conhecem, ou APIs REST convencionais com muitos endpoints.
gor := gin.New()
r.Use(gin.Recovery(), middleware.RequestID())
orders := r.Group("/orders")
orders.POST("", handler.CreateOrder)
Atenção: evitar deixar lógica de negócio dentro do handler Gin — extrair para use case, o handler só faz bind/validate/chamar aplicação/responder.
Echo
Similar ao Gin em filosofia, API um pouco mais limpa/minimalista, bom suporte a middleware chainável. Escolha equivalente ao Gin — não há motivo forte para preferir um ao outro além de gosto/familiaridade do time.
Fiber
Construído sobre fasthttp em vez de net/http — mais rápido em benchmarks, mas não é compatível com a interface net/http.Handler (não pode reusar middlewares padrão do ecossistema stdlib) e tem particularidades de gerenciamento de memória (reuso de *fiber.Ctx, cuidado ao guardar referências além do handler). Recomendar só quando performance extrema é requisito real e o time entende essas pegadinhas — não é o default seguro.
Acesso a banco de dados
database/sql + sqlc
sqlc gera código Go type-safe a partir de SQL puro (você escreve a query, ele gera a struct e o método). Vantagem: SQL explícito e revisável, zero mágica de ORM, performance previsível. Bom fit para times que gostam de controlar a query exata (comum em fintech, onde performance e previsibilidade de query importam).
pgx (driver Postgres nativo)
Mais rápido que lib/pq/database/sql genérico para Postgres, suporta tipos nativos do Postgres melhor, pooling built-in (pgxpool). Combina bem com sqlc (sqlc pode gerar código usando pgx como driver).
GORM
ORM completo, produtivo para CRUD rápido, migrations, associations. Trade-off: queries geradas automaticamente podem ser ineficientes (N+1 é fácil de introduzir sem perceber — sempre checar uso de Preload vs joins manuais), e abstrai demais para casos de alta performance/queries complexas. Ao revisar código GORM, sempre checar por N+1 e por queries dentro de loop.
Validação

go-playground/validator — tags de struct (validate:"required,email"), padrão de mercado.
Validação manual no construtor de Value Objects (mais alinhado a DDD, ver architecture-ddd.md) quando a regra é de domínio, não só formato.

Injeção de dependência
Go geralmente não precisa de um framework de DI (como Spring). Wiring manual e explícito em cmd/api/main.go é o padrão idiomático:
gofunc main() {
    db := setupDB()
    orderRepo := postgres.NewOrderRepository(db)
    paymentGateway := httpclient.NewPaymentGateway(cfg.PaymentURL)
    createOrderUC := application.NewCreateOrderUseCase(orderRepo, paymentGateway)
    handler := http.NewOrderHandler(createOrderUC)
    // ...
}
Para projetos muito grandes com muitos componentes, google/wire (geração de código compile-time) é aceitável — evitar DI containers com reflection em runtime (não é o estilo Go idiomático e perde type-safety).
Recomendação por padrão (quando o usuário não especificar)
Se o usuário não tiver preferência declarada: net/http (stdlib moderna) + sqlc/pgx + wiring manual + go-playground/validator. É a combinação mais idiomática, com menos dependências e mais alinhada ao que a comunidade Go sênior recomenda hoje. Ajustar imediatamente se o usuário mencionar Gin, Echo, Fiber ou GORM já em uso no projeto.