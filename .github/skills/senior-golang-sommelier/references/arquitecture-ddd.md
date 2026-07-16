Clean Architecture + DDD em Go
Leia este arquivo quando o usuário pedir estrutura de projeto, discutir organização de camadas, ou quando o domínio do problema tiver regras de negócio não triviais que justifiquem táticas de DDD.
Estrutura de pastas de referência
/cmd
  /api            → main.go, wiring/injeção de dependência, bootstrap
/internal
  /domain           → regras de negócio puras, SEM import de framework/DB/HTTP
    /order
      entity.go       → Entities e Value Objects
      repository.go   → interface (porta) do repositório, definida aqui
      service.go      → Domain Services (regra que não pertence a uma única entidade)
      errors.go       → erros de domínio (ErrOrderNotFound, ErrInvalidStatus...)
      events.go       → Domain Events, se aplicável
  /application      → casos de uso / orquestração (Use Cases)
    /order
      create_order.go
      cancel_order.go
  /infrastructure   → implementações concretas das interfaces do domínio
    /postgres
      order_repository.go   → implementa domain/order.Repository
    /httpclient
      payment_gateway.go
  /interfaces (ou /adapters)  → entrega: como o mundo externo fala com a aplicação
    /http
      handler.go
      middleware.go
      router.go
    /grpc
    /worker         → consumers de fila
/pkg                → código exportável/reutilizável entre projetos (raro, use com moderação)
Regra de dependência: as setas de import sempre apontam para dentro. domain não importa nada de infrastructure ou interfaces. application importa domain, nunca o contrário.
Táticas de DDD (usar só quando o domínio justificar)

Entity: tem identidade própria (ID), igualdade por identidade, não por valor. Ex.: Order, Customer.
Value Object: sem identidade, igualdade por valor, imutável. Ex.: Money, CPF, Email. Em Go, geralmente structs pequenas com validação no construtor:

go  type Money struct {
      amountCents int64
      currency    string
  }

  func NewMoney(cents int64, currency string) (Money, error) {
      if cents < 0 {
          return Money{}, fmt.Errorf("amount cannot be negative")
      }
      return Money{amountCents: cents, currency: currency}, nil
  }

Aggregate / Aggregate Root: cluster de entities/VOs tratado como unidade de consistência transacional. Só o Aggregate Root é acessado de fora; regras de invariância são garantidas dentro dele (não permita que código externo mute uma entity filha diretamente e quebre o invariante).
Repository: interface definida no domínio, implementação na infraestrutura. Um repository por Aggregate Root, não por tabela.

go  // domain/order/repository.go
  type Repository interface {
      FindByID(ctx context.Context, id OrderID) (*Order, error)
      Save(ctx context.Context, order *Order) error
  }

Domain Service: regra de negócio que não pertence naturalmente a uma única entity (ex.: cálculo que envolve duas entities diferentes).
Domain Event: fato relevante que aconteceu no domínio (OrderPlaced, PaymentFailed), útil para desacoplar side-effects (notificação, auditoria) da lógica principal. Pode ser publicado via outbox pattern para garantir consistência com a transação principal.
Application Service / Use Case: orquestra: carrega agregados via repository, chama métodos de domínio, persiste, dispara eventos. NÃO deve conter regra de negócio — só orquestração.

Bounded Context
Em sistemas maiores (ex.: um sistema de pagamentos com contexto de "Autorização", "Liquidação", "Antifraude"), cada bounded context deve ter seu próprio modelo de domínio — o mesmo conceito ("Transação") pode ter atributos e regras diferentes em cada contexto. Evite um "modelo de domínio único" compartilhado entre múltiplos contextos; isso é um anti-pattern comum (anemic shared model).
Sinais de que o DDD tático está sendo overengineered

Value Objects para todo campo primitivo sem regra de validação real.
Múltiplas camadas para um CRUD sem lógica de negócio.
Repository genérico (Repository[T any]) que já reintroduz o acoplamento que o padrão tentava evitar.

Nesses casos, prefira sugerir uma estrutura mais simples (handler → service → repository) e explicar por quê.