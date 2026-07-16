Checklist de Segurança e Middlewares em Go
Leia este arquivo ao revisar segurança de um serviço Go, ou ao implementar rate limiting, middlewares de auth/recovery/logging.
Checklist rápido (para revisão de código)

 Toda query SQL é parametrizada (nunca fmt.Sprintf/concatenação formando SQL)
 Todo input externo (body HTTP, query param, header, mensagem de fila) é validado antes de uso
 Toda chamada de rede (HTTP client, DB, cache, fila) tem timeout via context
 Handlers HTTP têm middleware de recover() para panics
 Dados sensíveis (senha, token, PAN, CVV) nunca aparecem em logs, nem em mensagens de erro devolvidas ao cliente
 Secrets vêm de env var/secret manager, nunca hardcoded ou commitados
 Endpoints públicos/sensíveis têm rate limiting
 Respostas de erro não vazam detalhes internos (stack trace, query SQL, path de arquivo) para o cliente final
 Comparação de segredos/tokens usa crypto/subtle.ConstantTimeCompare (evitar timing attack) quando aplicável
 Goroutines disparadas têm ciclo de vida claro (não há "fire and forget" sem contexto de cancelamento)
 Slices/maps recebidos como parâmetro não são mutados inesperadamente sem o caller saber (cuidado com aliasing)

Middleware de recovery (panic)
gofunc Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    logger.Error("panic recovered",
                        "error", err,
                        "path", r.URL.Path,
                        "stack", string(debug.Stack()),
                    )
                    w.WriteHeader(http.StatusInternalServerError)
                    _ = json.NewEncoder(w).Encode(map[string]string{
                        "error": "internal server error",
                    })
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
Rate limiting — single instance (token bucket, golang.org/x/time/rate)
gotype IPRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    r        rate.Limit
    b        int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        r:        r,
        b:        b,
    }
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.limiters[ip] = limiter
    }
    return limiter
}

func (i *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := clientIP(r)
        if !i.getLimiter(ip).Allow() {
            w.WriteHeader(http.StatusTooManyRequests)
            _ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
            return
        }
        next.ServeHTTP(w, r)
    })
}
Atenção em produção: o mapa limiters acima cresce indefinidamente — precisa de expiração/limpeza periódica (goroutine de GC do mapa, ou LRU cache) para não virar memory leak. Para múltiplas instâncias atrás de load balancer, o rate limit precisa ser compartilhado via Redis (redis_rate ou implementação própria com INCR + EXPIRE/sliding window), senão cada instância aplica o limite isoladamente.
Rate limiting distribuído (Redis, esboço)
gofunc (l *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    pipe := l.client.TxPipeline()
    incr := pipe.Incr(ctx, key)
    pipe.Expire(ctx, key, window)
    if _, err := pipe.Exec(ctx); err != nil {
        return false, fmt.Errorf("rate limit check: %w", err)
    }
    return incr.Val() <= int64(limit), nil
}
Middleware de request ID + logging estruturado
gofunc RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.NewString()
        }
        ctx := context.WithValue(r.Context(), requestIDKey{}, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
Usar context.WithValue apenas para dados transversais como request ID/trace ID — nunca para passar dependências de negócio (isso deve ser injeção explícita).
Validação de input
Preferir validação explícita e cedo (fail fast, no handler ou no construtor do Value Object), usando go-playground/validator para structs de request:
gotype CreateOrderRequest struct {
    CustomerID string  `json:"customer_id" validate:"required,uuid4"`
    AmountCents int64  `json:"amount_cents" validate:"required,gt=0"`
    Currency   string  `json:"currency" validate:"required,iso4217"`
}
Nunca confiar em validação só no frontend/client. Toda validação de negócio crítica é revalidada no backend.
Segredos e dados sensíveis em logs
go// ERRADO — loga o cartão inteiro
logger.Info("processing payment", "card_number", card.Number)

// CERTO — mascara, mostra só os últimos 4 dígitos
logger.Info("processing payment", "card_last4", card.Number[len(card.Number)-4:])
Em contexto de pagamentos/PCI-DSS: nunca logar PAN completo, CVV/CVC, trilha magnética, ou PIN — nem em logs de debug, nem em mensagens de erro/exception que possam ir para uma ferramenta de observabilidade.