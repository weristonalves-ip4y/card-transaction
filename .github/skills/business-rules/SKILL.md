---
name: business-rules
description: 'Identifica e mapeia regras de negócio do projeto Card API. Use para: entender fluxos de transação, validar implementações, identificar regras críticas de cálculo/inversão de sinal, mapear estratégias de movimento, regras de cache, validações ISO8583. Palavras-chave: regra de negócio, business rule, chargeback, estorno, inversão sinal, voucher vs card, validação transação, fluxo, cálculo saldo.'
argument-hint: 'Opcional: tipo de regra (movimento, validação, cache) ou domínio específico'
---

# Business Rules Mapper - Card API

## Objetivo

Identifica, documenta e valida **regras de negócio** do projeto de transações financeiras com cartões e vouchers. Serve como referência centralizada para entender objetivos, restrições e lógica crítica do domínio.

---

## Quando Usar

Use esta skill quando:
- ✅ **Implementar nova feature** - verificar regras relacionadas
- ✅ **Debugar comportamento** - entender lógica de negócio aplicada
- ✅ **Revisar código** - validar consistência com regras estabelecidas
- ✅ **Onboarding** - mapear domínio financeiro do projeto
- ✅ **Refatorar** - garantir que regras críticas sejam preservadas
- ✅ **Documentar** - criar/atualizar docs de regras

---

## Regras de Negócio Mapeadas

### 1. IDENTIFICAÇÃO: Card vs Voucher

**Regra**: Tipo de transação determinado por `ps_product_code` na tabela `paysmart_cards`

| Campo               | Localização          | Valores                               |
|---------------------|----------------------|---------------------------------------|
| `ps_product_code`   | `paysmart_cards`     | `011202` (CARD), `011401` (VOUCHER)   |
| `request_ps_product_code` | Request ISO8583 | `011201`, `011101`, `011202` (CARD), `013401`, `011401` (VOUCHER) |

**Impacto**:
- Define qual **Strategy** usar (`CardMovementStrategy` vs `VoucherMovementStrategy`)
- Define qual **Job** disparar (`ProcessMovementCardJob` vs `ProcessMovementVoucherJob`)
- Define qual **SP** executar (`sp_insert_movement_card` vs `sp_insert_movement_voucher`)

**Localização no Código**:
- Decisão: `AppServiceProvider` (service locator bindings)
- Validação: `TransactionValidationService`
- Join: `transaction_purchases.card_id = paysmart_cards.card_id` (ambos strings, não FK para `id`)

---

### 2. INVERSÃO DE SINAL: ChargeBack/Estorno

**⚠️ REGRA CRÍTICA**: ChargeBack **inverte sinal** para gerar **CRÉDITO** (devolução ao cliente)

#### CardMovementStrategy
```php
$transactionValue = match (true) {
    $entity instanceof TransactionPurchaseEntity => $baseValue,           // +100 (débito)
    $entity instanceof TransactionWithdrawalEntity => $baseValue,         // +100 (débito)
    $entity instanceof TransactionReverseEntity => -1 * abs($baseValue),  // -100 (crédito)
    $entity instanceof TransactionReverseCancelEntity => abs($baseValue), // +100 (débito)
};
```

**SP Behavior**:
- `sp_insert_movement_card`: **NÃO multiplica** por -1 (valor direto)
  - Purchase: passa +10 → SP registra +10 (débito)
  - Withdrawal: passa +10 → SP registra +10 (débito)
  - ChargeBack: passa -10 → SP registra -10 (crédito)

#### VoucherMovementStrategy
```php
$transactionValue = match (true) {
    $entity instanceof TransactionPurchaseEntity => $baseValue,          // +100
    $entity instanceof TransactionReverseEntity => -1 * $baseValue,      // -100 (inverte)
    $entity instanceof TransactionReverseCancelEntity => $baseValue,     // +100
};
```

**SP Behavior**:
- `sp_insert_movement_voucher`: **MULTIPLICA** por -1 internamente
  - Purchase: passa +100 → SP inverte → -100 (débito)
  - ChargeBack: passa -100 → SP inverte → +100 (crédito)

**Impacto**: Erros na inversão de sinal geram **débitos/créditos incorretos no saldo**.

---

### 3. FLUXOS DE TRANSAÇÃO

#### 3.1 Purchase (Compra)

**Endpoint**: `POST /api/transaction` → `TransactionController`

**Fluxo**:
1. Parse ISO8583 → `PurchaseRequestDTO`
2. Validação → `TransactionValidationService::validate()`
3. Balance check (do cache) → `BalanceService`
4. Criação transação → Repository
5. **Cálculo de saldo em memória** (balance - valor)
6. **Invalidação de cache** (antes de responder)
7. **Dispatch Job assíncrono** (se aprovado: `response_code === '00'`)
8. Retorna resposta ISO8583

**Regra**: Movimento executado **assincronamente** via Job (não bloqueia response)

#### 3.2 Cancel (Cancelamento)

**Endpoint**: `POST /api/cancel` → `CancelController`

**Fluxo**:
1. Parse ISO8583 → `CancelRequestDTO`
2. Busca transação original (via `purchase_id/withdrawal_id/transfer_id`)
3. Validação → `TransactionValidationService::validate()`
4. Balance check (do cache)
5. Criação transação reversa → Repository (`transaction_reverses`)
6. **Cálculo de saldo em memória** (balance + valor) - **CRÉDITO**
7. **Invalidação de cache**
8. **Dispatch Job** com **parent ID** (aguarda transação original)

**Regra**: Cancel **devolve saldo** (crédito) e tem **parent transaction** para garantir ordem

#### 3.3 ChargeBack (Estorno)

**Endpoint**: `POST /api/chargeback` → `ChargeBackController`

**Tipos**:
- **Reversal** (Estorno): Cria transação reversa com `chargeback_id`
- **Cancel Reversal** (Cancelamento de Estorno): Cria transação com `original_chargeback_id`

**Fluxo Reversal**:
1. Parse ISO8583
2. Validação
3. Cria transação reversa (`transaction_reverses`) com `chargeback_id`
4. **Inversão de sinal** para CRÉDITO
5. Dispatch Job com delay de 5s e parent ID

**Fluxo Cancel Reversal**:
1. Cancela um estorno anterior
2. **Valor positivo** (débita novamente, cancela o crédito)
3. Campo `original_chargeback_id` aponta para estorno cancelado

#### 3.4 Withdrawal (Saque)

**Endpoint**: `POST /api/withdrawals` → `WithdrawalController`

**⚠️ REGRA CRÍTICA**: Withdrawal é **APENAS de cartão físico ou virtual** - não existe withdrawal para voucher

**Fluxo**:
1. Parse ISO8583 → `WithdrawalRequestDTO`
2. Validação customizada → `WithdrawalService::validateWithdrawal()`
   - Usa `withdrawal_id` ao invés de `purchase_id` para duplicação
   - Converte DTO para `PurchaseRequestDTO` internamente
   - Reutiliza `TransactionValidationService::validate()`
3. Balance check (do cache) → `BalanceService`
4. Criação transação → Repository (`transaction_purchases` com `withdrawal_id`)
5. **Cálculo de saldo em memória** (balance - valor) - **DÉBITO**
6. **Invalidação de cache** (antes de responder)
7. **Dispatch Job assíncrono** `ProcessWithdrawalCardMovementJob` (se aprovado)
8. Retorna resposta ISO8583

**Características**:
- **Tabela**: `transaction_purchases` (mesma de Purchase)
- **Campo identificador**: `withdrawal_id` (não `purchase_id`)
- **Entity**: `TransactionWithdrawalEntity` (alias de `TransactionPurchaseEntity`)
- **Job**: `ProcessWithdrawalCardMovementJob` (usa `CardMovementStrategy`)
- **Parent ID**: Sempre `null` (withdrawal não tem transação pai)
- **Output**: `WithdrawalOutput` (response codes específicos)

**Validações MTI**:
```php
// MTIs aceitos para withdrawal
if (!in_array($request->original_iso8583['mti'], ['0200', '0100', '0120'])) {
    return response()->json(['message' => 'MTI inválido.'], 404);
}
```

**Regra de Sinal**: 
- Withdrawal sempre **DÉBITA** (valor positivo na Strategy)
- `CardMovementStrategy` trata `TransactionWithdrawalEntity` como débito
- SP `sp_insert_movement_card` registra valor direto (sem multiplicação)

**Cancelamento de Withdrawal**:
- Endpoint: `POST /api/withdrawals/cancel`
- Usa `CancelController` (mesmo de Purchase)
- Identifica via `original_withdrawal_id`
- Devolve saldo (crédito)

**Queries de Withdrawal**:
- Endpoint: `POST /api/withdrawalQueries`
- Usa `QueriesController` com `withdrawal_query_id`

**Diferenças vs Purchase**:

| Característica       | Purchase               | Withdrawal                 |
|----------------------|------------------------|----------------------------|
| Endpoint             | `/api/transaction`     | `/api/withdrawals`         |
| Campo ID             | `purchase_id`          | `withdrawal_id`            |
| Tipos permitidos     | Cartão (físico/virtual) ou Voucher | **Cartão (físico/virtual)** |
| Job                  | `ProcessMovementCardJob` ou `ProcessMovementVoucherJob` | `ProcessWithdrawalCardMovementJob` |
| Localização          | Estabelecimento (loja) | ATM/Caixa eletrônico       |
| MCC                  | Categoria variada      | Específico de ATM          |

**Localização no Código**:
- Service: `app/Services/WithdrawalService.php`
- Controller: `app/Http/Controllers/WithdrawalController.php`
- DTO: `app/DTOs/WithdrawalRequestDTO.php`, `app/DTOs/WithdrawalOutput.php`
- Entity: `app/Entities/TransactionWithdrawalEntity.php`
- Job: `app/Jobs/ProcessWithdrawalCardMovementJob.php`

---

### 3.5 DESCRIÇÕES DE MOVIMENTOS NO EXTRATO

**Regra**: Cada tipo de transação tem padrão específico de descrição para o extrato do cliente

#### Purchase (Compra)

**Tipos permitidos**: 
- Cartão físico
- Cartão virtual  
- Voucher

**Padrões de descrição**:

1. **Compra normal**:
   ```
   COMPRA CARTÃO | BANCO DO SEU NEGOCIO BSN
   ```
   - Formato: `COMPRA CARTÃO | {Nome do estabelecimento}`
   - Fonte: ISO8583 DE043 (Card Acceptor Name/Location)

2. **Estorno de compra** (ChargeBack Reversal):
   ```
   ESTORNO | COMPRA CARTÃO | 1029203                GUARULHOS     BRA
   ```
   - Formato: `ESTORNO | COMPRA CARTÃO | {Localização completa}`
   - Inclui: Terminal ID + Cidade + País
   - **⚠️ CRÍTICO**: Local deve vir da transação ORIGINAL

3. **Cancelamento de estorno** (ChargeBack Cancel):
   ```
   CANCELAMENTO DE ESTORNO CARTÃO | LOCAL INFORMADO
   ```
   - Formato: `CANCELAMENTO DE ESTORNO CARTÃO | {Localização completa}`
   - **⚠️ PROBLEMA CONHECIDO**: Local não está sendo passado corretamente
   - Deve incluir dados do estabelecimento da transação original

**Cancelamentos**:
- `POST /api/cancel` (Purchase Cancel) - Devolve saldo
- `POST /api/chargeback` (ChargeBack) - Estorno/Cancelamento de estorno

#### Withdrawal (Saque)

**Tipos permitidos**:
- Cartão físico
- Cartão virtual
- **NÃO** voucher

**Padrões de descrição**:

1. **Saque normal**:
   ```
   SAQUE | 1029202                GUARULHOS     BRA
   ```
   - Formato: `SAQUE | {Terminal ID} {Cidade} {País}`
   - Fonte: ISO8583 DE043 (Card Acceptor Name/Location)

2. **Estorno de saque** (ChargeBack Reversal):
   ```
   ESTORNO | SAQUE | 1029203                GUARULHOS     BRA
   ```
   - Formato: `ESTORNO | SAQUE | {Localização completa}`
   - Inclui: Terminal ID + Cidade + País
   - **⚠️ CRÍTICO**: Local deve vir da transação ORIGINAL

3. **Cancelamento de estorno de saque** (ChargeBack Cancel):
   ```
   CANCELAMENTO DE ESTORNO SAQUE | LOCAL INFORMADO
   ```
   - Formato: `CANCELAMENTO DE ESTORNO SAQUE | {Localização completa}`
   - **⚠️ PROBLEMA CONHECIDO**: Local não está sendo passado corretamente
   - Deve incluir dados do ATM da transação original

**Cancelamentos**:
- `POST /api/withdrawals/cancel` (Withdrawal Cancel) - Devolve saldo
- `POST /api/chargeback` (ChargeBack) - Estorno/Cancelamento de estorno

#### Regra de Construção das Descrições

**Responsável**: `CardMovementStrategy::buildDescription()` e `VoucherMovementStrategy::buildDescription()`

**Campos ISO8583 utilizados**:
- **DE043** - Card Acceptor Name/Location - Nome e localização do estabelecimento/ATM
  - Purchase: "BANCO DO SEU NEGOCIO BSN"
  - Withdrawal: "1029202                GUARULHOS     BRA"

**⚠️ PROBLEMA IDENTIFICADO - Cancelamento de Estorno sem Local**:

Quando processar **ChargeBack Cancel** (cancelamento de estorno):
1. ✅ **Deve** buscar transação original (chargeback) 
2. ✅ **Deve** obter `request_card_acceptor_name_location` da transação original
3. ❌ **Problema**: Atualmente envia "LOCAL INFORMADO" genérico
4. ✅ **Esperado**: Enviar localização completa do estabelecimento/ATM

**Localização do Bug**:
- Strategy: `app/Services/Transaction/Strategies/CardMovementStrategy.php`
- Método: `buildDescription()`
- Verificar: Se `TransactionReverseCancelEntity` está obtendo dados da transação original

**Correção necessária**:
```php
// ChargeBack Cancel deve buscar location da transação ORIGINAL (chargeback)
if ($entity instanceof TransactionReverseCancelEntity) {
    $originalChargeback = findOriginalChargeBack($entity->originalChargebackId);
    $location = $originalChargeback->request_card_acceptor_name_location;
    
    // Purchase
    return "CANCELAMENTO DE ESTORNO CARTÃO | {$location}";
    
    // Withdrawal  
    return "CANCELAMENTO DE ESTORNO SAQUE | {$location}";
}
```

---

### 4. VALIDAÇÕES

**Service**: `TransactionValidationService`

#### Ordem de Validação (Orquestrada)

1. **Duplicação** - Verifica se `purchase_id` já existe
2. **Cartão** - Via `CardValidationService` (status, expiração, bloqueio)
3. **Produto compatível** - Valida `ps_product_code` vs tipo de transação
4. **Limite mensal** - Calcula soma do mês + transação atual vs `card_monthly_limit`
5. **Saldo** - Via `BalanceService` ou `BalanceVoucherService`

#### 4.1 Validação de Limite Mensal

**⚠️ REGRA CRÍTICA**: Limite mensal verificado ANTES de consultar saldo

**Campos do Cartão**:
- `card_monthly_limit` (float) - Limite mensal em R$
- `card_daily_limit` (float) - Limite diário (⚠️ não implementado atualmente)
- `card_check_limit` (bool) - Se `0`, ignora validação de limite

**Lógica**:
```php
// 1. Se forceAccept = true → IGNORA validação
if ($request->forceAccept) return VALID;

// 2. Se card_check_limit != 1 → IGNORA validação
if ($card->card_check_limit != 1) return VALID;

// 3. Se limite = 0 ou limite < valor transação → BLOQUEADO
if ($card->card_monthly_limit == 0 || $card->card_monthly_limit < $transactionValue) {
    return INVALID('01', 'Limite mensal bloqueado/excedido');
}

// 4. Calcula soma do mês via fn_get_sum_month_card_transaction()
$monthlySum = getMonthlySum($card->card_id);

// 5. Verifica se limite seria ultrapassado
if ($card->card_monthly_limit < ($monthlySum + $transactionValue)) {
    return INVALID('01', 'Limite mensal excedido');
}
```

**Função SQL**: `fn_get_sum_month_card_transaction(@card_id)`
- Retorna soma de transações aprovadas (`response_code = '00'`) no mês atual
- Considera apenas `transaction_purchases` (não reverses)
- Usa índice em `(card_id, created_at)` para performance

#### 4.2 Validação de Produto (ps_product_code)

**Regra**: Tipo de transação deve ser compatível com produto do cartão

| ps_product_code | Tipo Permitido           | Erro se Incompatível |
|-----------------|--------------------------|----------------------|
| `011401`        | VOUCHER apenas           | `06` - Produto não permitido |
| `011202`        | NÃO-VOUCHER apenas       | `06` - Produto não permitido |
| `null`          | NÃO-VOUCHER apenas       | `06` - Produto não permitido |

**Detecção de Voucher**:
```php
// Via request_ps_product_code em ISO8583
$isVoucher = $request->psProductCode === '013401';
```

#### 4.3 Validação de MCC (Merchant Category Code)

**⚠️ Regra**: MCC está presente mas validação de whitelist/blacklist **não está implementada**

**Campos**:
- `request_mcc` (string) - ISO8583 DE018
- Usado para: Registro em `transaction_purchases.request_mcc`
- **Voucher**: MCC passado para SP `sp_insert_movement_voucher`

**Response Code**:
- `06` - MCC inválido (se validação implementada)

#### 4.4 Response Codes ISO8583

**Aprovações**:
- `00` - Aprovado

**Erros de Cartão**:
- `14` - Cartão inválido/não encontrado
- `54` - Cartão expirado
- `57` - Transação não permitida para este cartão

**Erros de Saldo/Limite**:
- `01` - Saldo insuficiente OU Limite mensal excedido
- `51` - Saldo insuficiente (código alternativo)

**Erros de Validação**:
- `06` - Produto/MCC inválido para este cartão
- `07` - Transação duplicada

**Erros de Fraude/Segurança**:
- `59` - Suspeita de fraude

**Localização**: 
- `app/Services/Validation/TransactionValidationService.php`
- `app/Services/Validation/CardValidationService.php`
- `app/DTOs/PurchaseOutput.php` (mapeamento codes → messages)

---

### 5. CACHE STRATEGY (Regra de Performance)

**Regra**: Dual-layer caching para evitar N+1 queries

#### Request Cache (In-Memory)
```php
// Singleton por HTTP request
$this->cache->remember($key, function() {
    return DB::query(...);
});
```

**Objetivo**: Prevenir queries duplicadas dentro de **1 request**
- Exemplo: 4 validações → 1 query ao DB

**Repositórios**: `CardRepository`, `BalanceRepository`, `TransactionRepository`

#### Balance Cache Invalidation
```php
$this->balanceService->invalidateCache($accountId);
```

**Regra CRÍTICA**: Deve ser chamado **ANTES de responder** após operação que afeta saldo
- Purchase approved → invalida cache
- Cancel → invalida cache
- ChargeBack → invalida cache

**Motivo**: Próximo request precisa de saldo atualizado

---

### 6. STORED PROCEDURES

**Regra**: Movimentos financeiros executados via SPs (não Eloquent)

| SP                            | Uso                        | Multiplicador Interno |
|-------------------------------|----------------------------|-----------------------|
| `sp_insert_movement_card`     | Movimentos de cartão       | Nenhum (valor direto) |
| `sp_insert_movement_voucher`  | Movimentos de voucher      | **× -1**              |
| `fn_get_account_balance`      | Consulta saldo conta       | N/A                   |
| `fn_get_voucher_balance`      | Consulta saldo voucher     | N/A                   |

**Binding Syntax**:
```php
DB::table('fn_get_account_balance(?)')
    ->setBindings([$accountId])
    ->selectRaw('balance')
    ->first();
```

**Regra**: **NÃO usar Eloquent** para balance queries (performance)

---

### 7. DTOs IMUTÁVEIS

**Regra**: Todo dado de entrada parseado via **readonly DTOs**

**Pattern**:
```php
$dto = PurchaseRequestDTO::fromArray($requestData);
// DTO é imutável - não pode ser modificado após criação
```

**DTOs Principais**:
- `PurchaseRequestDTO` - Request de compra
- `WithdrawalRequestDTO` - Request de saque
- `TransactionMessageDTO` - ISO8583 message (50+ campos)
- `PurchaseOutput` - Response padronizada
- `WithdrawalOutput` - Response de saque
- `ValidationResultDTO` - Resultado de validações

**Benefícios**:
- Type safety
- Validação centralizada
- Imutabilidade garante consistência

---

### 8. STRATEGY PATTERN

**Regra**: Estratégias separam lógica de Card vs Voucher

```
MovementStrategyInterface
├── CardMovementStrategy
└── VoucherMovementStrategy
```

**Factory**:
```php
$strategy = MovementStrategyFactory::create('card'); // ou 'voucher'
$strategy->execute($entity, $cardId);
```

**Responsabilidades**:
- `buildDescription()` - Descrição do movimento
- Cálculo de inversão de sinal
- Execução de SP apropriada

**Regra**: **Single Source of Truth** - Jobs delegam para Strategies (não duplicam lógica)

---

### 9. ASYNCHRONOUS MOVEMENTS

**Regra**: Movimentos executados **assincronamente** via Jobs

**Jobs**:
- `ProcessMovementCardJob` - Compras com cartão
- `ProcessMovementVoucherJob` - Compras com voucher
- `ProcessWithdrawalCardMovementJob` - Saques (sempre cartão)

**Fluxo**:
1. Transação criada com `movement_status = 'pending'`
2. Job disparado (apenas se `response_code === '00'`)
3. Job processa:
   - Converte `stdClass` → `Entity` (via `fromArray`)
   - Obtém Strategy via Factory (ou injeta diretamente para Withdrawal)
   - **Delega execução** para Strategy
4. Marca `movement_status = 'completed'`

**Parent Transaction Wait**:
- Cancel/ChargeBack esperam transação original completar
- Retry com delay (5s, 10s, 15s)
- Eventual consistency após 3 tentativas

**Idempotência**: Job verifica `movement_status` antes de processar

---

### 10. ISO8583 MESSAGE STRUCTURE

**Regra**: Todas requests/responses seguem padrão ISO8583

#### Campos Críticos por Categoria

**Identificação da Transação**:
- **MTI** (Message Type Indicator) - Tipo de mensagem (0100, 0200, 0400, etc)
- **DE003** - Processing Code - Código de processamento
- **DE011** - STAN (System Trace Audit Number) - Número de auditoria
- **DE007** - Transmission Date/Time - Data/hora da transmissão
- **DE012/DE013** - Local Transaction Time/Date - Hora/data local

**Valores Monetários**:
- **DE004** - Transaction Amount - Valor da transação (campo crítico)
- **DE005** - Settlement Amount - Valor de liquidação
- **DE006** - Cardholder Billing Amount - Valor em moeda do portador
- **DE009** - Conversion Rate - Taxa de conversão
- **DE049** - Transaction Currency Code - Código moeda transação (986 = BRL)
- **DE050** - Settlement Currency Code - Código moeda liquidação
- **DE051** - Cardholder Billing Currency Code - Código moeda portador

**Dados do Cartão**:
- **DE002** - Primary Account Number (PAN) - Número do cartão (mascarado)
- **DE014** - Expiration Date - Data expiração (YYMM)
- **DE022** - POS Entry Mode - Modo de entrada (chip, contactless, manual)
- **DE025** - POS Condition Code - Condição do POS

**Estabelecimento**:
- **DE018** - MCC (Merchant Category Code) - **Categoria do estabelecimento**
- **DE032** - Acquiring Institution Code - Código da instituição adquirente
- **DE033** - Forwarding Institution Code - Código instituição encaminhadora
- **DE042** - Card Acceptor Identification Code - ID do estabelecimento
- **DE043** - Card Acceptor Name/Location - **Nome e localização** (usado em descrição movimento)

**Resposta**:
- **DE038** - Authorization Identification Response - Código de autorização
- **DE039** - **Response Code** - Código resposta (00 = aprovado, 01 = saldo insuf, etc)

**Dados Adicionais**:
- **DE048** - Additional Data - **Contém ps_product_code** (011201/011101/011202/013401/011401)
- **DE062** - Reserved for Private Use - Dados privativos
- **DE090** - Original Data Elements - Dados originais (chargebacks)

**DTO**: `TransactionMessageDTO` encapsula 50+ Data Elements com parsing e validação

#### Parsing de ps_product_code

**⚠️ CAMPO CRÍTICO**: Define tipo Card vs Voucher

**Localização no Request**:
```php
// Em ISO8583 DE048 (Additional Data)
$psProductCode = $data['iso8583_message']['de048']['ps_product_code'] ?? null;

// ou alternativamente
$psProductCode = $data['product']['code'] ?? null;
```

**Valores**:
- `011201`, `011101` ou `011202` - CARD (transações com cartão)
- `013401` ou `011401` - VOUCHER (transações com voucher)

---

### 11. REGRAS DE BLOQUEIO E FRAUDE

#### 11.1 Status de Bloqueio do Cartão

**Constantes** (`PaysmartCard`):
- `PAYSMARTCARD_STATUS_ACTIVE = 4` - Ativo
- `PAYSMARTCARD_STATUS_BLOCKED = 3` - Bloqueado
- `PAYSMARTCARD_STATUS_AWAIT_UNBLOCK = 6` - Aguardando desbloqueio
- `PAYSMARTCARD_STATUS_ISSUED = 2` - Emitido (não ativado)
- `PAYSMARTCARD_STATUS_EXPIRED = 5` - Expirado

**Campos de Bloqueio**:
- `card_blocking_reason_id` (int) - ID do motivo
- `card_blocking_reason_description` (string) - Descrição do motivo
- `card_blocked_by_user_id` (int) - Quem bloqueou
- `card_blocked_ip` (string) - IP de origem do bloqueio
- `card_blocked_at` (datetime) - Quando foi bloqueado
- `card_unblock_first` (bool) - Primeiro desbloqueio

**Regra**: Cartão bloqueado (`status = 3`) **NÃO pode transacionar**
- Response Code: `57` - Transação não permitida

#### 11.2 Detecção de Fraude

**⚠️ Regra Genérica**: Sistema tem suporte a `response_code = '59'` (suspeita de fraude)

**Possíveis Critérios** (validar implementação):
- Múltiplas transações em curto período
- Valores atipicamente altos
- Mudanças bruscas de localização (DE043)
- Tentativas após negações sucessivas

**Response**:
- Code: `59`
- Message: "Suspeita de fraude, transação negada."

**⚠️ IMPORTANTE**: Validação específica de fraude **não está totalmente documentada** no código atual. Pode estar em regras de negócio externas ou gateway.

---

### 12. LIMITES E RESTRIÇÕES

#### 12.1 Limite Mensal por Cartão

**Campos**:
- `card_monthly_limit` (float) - Valor máximo mensal em R$
- `card_check_limit` (bool) - Flag ativa/desativa verificação

**Regra**: 
```
soma_mes_atual + valor_transação <= card_monthly_limit
```

**Cálculo**:
- Função: `fn_get_sum_month_card_transaction(@card_id)`
- Considera: Transações aprovadas (`response_code = '00'`) do mês atual
- Inclui: Cartão físico + todos virtuais vinculados

**Bypass**:
- `forceAccept = true` → Ignora limite
- `card_check_limit = 0` → Ignora limite

#### 12.2 Limite Diário (Não Implementado)

**Campo**: `card_daily_limit` (float)

**⚠️ Status**: Campo existe em `CardEntity` mas **validação NÃO implementada** atualmente.

**Próximos Passos**: Se implementar, seguir mesmo padrão de `validateMonthlyLimit()`

#### 12.3 Limite de Cancelamento

**Regra**: Existe response code específico para limite de cancelamento ultrapassado

**Response**:
- Code: `01` ou código específico
- Message: "Limite de cancelamento ultrapassado" (em `CancelOutput`)

**⚠️ Regra não documentada**: Validar lógica específica se necessário

---

## Procedimento de Uso

### 1. Identificar Domínio

Determine qual área de negócio você está trabalhando:
- Movimentos financeiros → Strategies, SPs
- Validações → TransactionValidationService
- Cache → RequestCache, BalanceService
- Fluxo de transação → Services (Purchase, Cancel, ChargeBack)

### 2. Consultar Regras Específicas

Use seções acima como referência:
- **Card vs Voucher** → Seção 1
- **Inversão de sinal** → Seção 2 (CRÍTICA)
- **Fluxos** → Seção 3
- **Validações** → Seção 4
- Etc.

### 3. Validar Implementação

Checklist:
- [ ] Inversão de sinal correta (ChargeBack)?
- [ ] Strategy apropriada (Card vs Voucher)?
- [ ] Cache invalidado após operação?
- [ ] Job disparado se aprovado?
- [ ] DTO imutável usado?
- [ ] Parent transaction ID passado (se Cancel/ChargeBack)?

### 4. Atualizar Skill (se necessário)

Se descobrir nova regra não documentada:
1. Adicione na seção apropriada
2. Marque como **⚠️ CRÍTICA** se afeta saldo/consistência
3. Documente localização no código
4. Adicione exemplo de uso

---

## Regras Críticas (Checklist)

**⚠️ NÃO violar estas regras**:

1. **Inversão de sinal em ChargeBack** - SEMPRE inverter para crédito
2. **Cache invalidation** - SEMPRE antes de responder após operação de saldo
3. **Job dispatch condicional** - APENAS se `response_code === '00'`
4. **Parent transaction** - SEMPRE passar para Cancel/ChargeBack (null para Withdrawal)
5. **DTO imutável** - NUNCA modificar após criação
6. **Estratégia correta** - Card vs Voucher baseado em `ps_product_code`
7. **SP correto** - Card = `sp_insert_movement_card`, Voucher = `sp_insert_movement_voucher`
8. **No Eloquent para balance** - SEMPRE usar functions (`fn_get_account_balance`)
9. **Validação de limite mensal** - ANTES de consultar saldo
10. **Produto compatível** - Validar `ps_product_code` vs tipo transação (011202 vs 013401)
11. **Ordem de validação** - Duplicação → Cartão → Produto → Limite → Saldo
12. **Cartão bloqueado** - Status 3 NÃO pode transacionar (code 57)
13. **Limite mensal = 0** - Considerado bloqueado (code 01)
14. **MCC registrado** - SEMPRE salvar em `request_mcc` mesmo sem validação
15. **Withdrawal exclusivo para cartão** - NUNCA criar withdrawal para voucher
16. **Descrição de extrato** - SEMPRE incluir localização completa (DE043)
17. **ChargeBack Cancel com local** - Buscar `request_card_acceptor_name_location` da transação ORIGINAL

---

## Localização no Código

### Strategies
- `app/Services/Transaction/Strategies/CardMovementStrategy.php`
- `app/Services/Transaction/Strategies/VoucherMovementStrategy.php`
- `app/Services/Transaction/Strategies/MovementStrategyFactory.php`

### Services (Fluxos)
- `app/Services/PurchaseService.php`
- `app/Services/PurchaseVoucherService.php`
- `app/Services/WithdrawalService.php`
- `app/Services/Cancel/CancelCardService.php`
- `app/Services/Cancel/CancelVoucherService.php`
- `app/Services/ChargeBack/ChargeBackReversalCardService.php`
- `app/Services/ChargeBack/ChargeBackReversalVoucherService.php`

### Validações
- `app/Services/Validation/TransactionValidationService.php`
- `app/Services/Validation/CardValidationService.php`

### DTOs
- `app/DTOs/PurchaseRequestDTO.php`
- `app/DTOs/WithdrawalRequestDTO.php`
- `app/DTOs/TransactionMessageDTO.php`
- `app/DTOs/PurchaseOutput.php`
- `app/DTOs/WithdrawalOutput.php`

### Entities
- `app/Entities/TransactionPurchaseEntity.php`
- `app/Entities/TransactionWithdrawalEntity.php` - Alias de PurchaseEntity
- `app/Entities/TransactionReverseEntity.php`
- `app/Entities/TransactionReverseCancelEntity.php`
- `app/Entities/CardEntity.php`

### Jobs
- `app/Jobs/ProcessMovementCardJob.php`
- `app/Jobs/ProcessMovementVoucherJob.php`
- `app/Jobs/ProcessWithdrawalCardMovementJob.php`

### Controllers
- `app/Http/Controllers/TransactionController.php` - Purchase (POST /api/transaction)
- `app/Http/Controllers/WithdrawalController.php` - Withdrawal (POST /api/withdrawals)
- `app/Http/Controllers/CancelController.php` - Cancel (POST /api/cancel, /api/withdrawals/cancel)
- `app/Http/Controllers/ChargeBackController.php` - ChargeBack (POST /api/chargeback)
- `app/Http/Controllers/QueriesController.php` - Queries (POST /api/withdrawalQueries)

### Stored Procedures
- `database/003_sp_insert_movement_card.sql`
- `database/003_sp_insert_movement_voucher.sql`
- `database/001_fn_get_account_balance.sql`
- `database/001_fn_get_voucher_balance.sql`

---

## Exemplos de Uso da Skill

### Exemplo 1: Validar regra de withdrawal

```
Você: /business-rules withdrawal
     Preciso implementar withdrawal. Quais regras devo seguir?

Skill retorna:
- Seção 3.4: Withdrawal
- Cartão físico OU virtual (NUNCA voucher)
- Usar WithdrawalRequestDTO
- Campo identificador: withdrawal_id (não purchase_id)
- Usa CardMovementStrategy (débito direto)
- ProcessWithdrawalCardMovementJob
- Parent ID sempre null (não tem transação pai)
- MTI aceitos: 0200, 0100, 0120
- Cancelamento via /api/withdrawals/cancel
```

### Exemplo 2: Debugar descrição de extrato sem local

```
Você: /business-rules extrato
     Cancelamento de estorno está aparecendo "LOCAL INFORMADO" 
     ao invés da localização real. Como corrigir?

Skill retorna:
- Seção 3.5: Descrições de Movimentos no Extrato
- ChargeBack Cancel DEVE buscar location da transação ORIGINAL
- Campo: request_card_acceptor_name_location
- Purchase: "CANCELAMENTO DE ESTORNO CARTÃO | {location}"
- Withdrawal: "CANCELAMENTO DE ESTORNO SAQUE | {location}"
- Verificar CardMovementStrategy::buildDescription()
- Entity: TransactionReverseCancelEntity
- Deve usar originalChargebackId para buscar dados
```

### Exemplo 3: Debugar valor incorreto

```
Você: /business-rules movimento
     ChargeBack está debitando ao invés de creditar. Por quê?

Skill retorna:
- Verificar seção 2 (Inversão de Sinal)
- ChargeBack DEVE inverter: -1 * abs($baseValue)
- Verificar Strategy usada (Card vs Voucher)
- Verificar SP behavior (Card direto, Voucher multiplica -1)
- Checar tipo de Entity (TransactionReverseEntity)
```

### Exemplo 4: Validar limite mensal

```
Você: /business-rules limites
     Transação negada com code 01 mas saldo está OK. O que pode ser?

Skill retorna:
- Seção 12.1: Limite Mensal
- Code 01 pode ser saldo insuficiente OU limite mensal excedido
- Verificar card_monthly_limit vs soma do mês
- Usar fn_get_sum_month_card_transaction(@card_id)
- Verificar se card_check_limit = 1 (ativo)
- Se limite = 0 → considerado bloqueado
```

### Exemplo 5: Validar cache strategy

```
Você: /business-rules cache
     Saldo não atualiza após compra. Revisar regras de cache?

Skill retorna:
- Seção 5: Cache Strategy
- Verificar invalidateCache() chamado ANTES de responder
- Checar RequestCache vs Balance Cache
- Balance invalidation CRÍTICA após operações de saldo
```

### Exemplo 6: Implementar validação de MCC

```
Você: /business-rules mcc
     Preciso bloquear MCCs específicos (ex: 4816 = telecom). Como fazer?

Skill retorna:
- Seção 4.3: Validação de MCC
- Atualmente MCC é apenas registrado, não validado
- Response code: '06' - MCC inválido
- Implementar whitelist/blacklist em CardValidationService
- Campo: request_mcc (ISO8583 DE018)
- Para voucher: MCC passado para SP
```

---

## Próximos Passos

Após mapear regras de negócio:
1. **Implementar** seguindo padrões identificados
2. **Testar** validando regras críticas (use checklist acima)
3. **Documentar** atualizações nesta skill (manter sincronizado com código)
4. **Revisar** código existente para consistência

### Regras Pendentes de Implementação

**⚠️ Identificadas mas NÃO implementadas**:
1. **Limite diário** - Campo existe (`card_daily_limit`) mas validação ausente
2. **MCC whitelist/blacklist** - Apenas registrado, não validado
3. **Regras de fraude específicas** - Código 59 existe mas critérios não documentados
4. **Horários permitidos** - Não identificado no código
5. **Limites por tipo de transação** - Apenas limite mensal genérico

**🐛 Bugs Identificados**:
1. **Cancelamento de estorno sem local** - ChargeBack Cancel exibe "LOCAL INFORMADO" genérico
   - Deve buscar `request_card_acceptor_name_location` da transação original
   - Localização: `CardMovementStrategy::buildDescription()`
   - Entity afetada: `TransactionReverseCancelEntity`
   - Correção: Buscar dados via `originalChargebackId`

Se implementar alguma destas, **atualizar esta skill** com:
- Localização no código
- Lógica de validação
- Response codes
- Exemplos de uso

---

## Sumário de Regras por Categoria

| Categoria             | Regras Principais                                         | Status        |
|-----------------------|-----------------------------------------------------------|---------------|
| **Identificação**     | ps_product_code (Card vs Voucher)                         | ✅ Implementado |
| **Inversão Sinal**    | ChargeBack inverte para crédito                           | ✅ Crítico     |
| **Fluxos**            | Purchase, Withdrawal, Cancel, ChargeBack (Reversal/Cancel) | ✅ Implementado |
| **Withdrawal**        | APENAS cartão físico/virtual, withdrawal_id, sem parent   | ✅ Implementado |
| **Descrições Extrato**| Padrões por tipo (Purchase/Withdrawal/ChargeBack)         | ⚠️ Bug no Cancel |
| **Validações**        | Duplicação, Cartão, Produto, Limite, Saldo               | ✅ Implementado |
| **Limite Mensal**     | card_monthly_limit vs soma do mês                         | ✅ Implementado |
| **Limite Diário**     | card_daily_limit                                          | ⚠️ Pendente    |
| **MCC Validation**    | Whitelist/blacklist de categorias                         | ⚠️ Pendente    |
| **Cache Strategy**    | Request Cache + Balance Invalidation                      | ✅ Crítico     |
| **Stored Procedures** | sp_insert_movement_* + fn_get_*_balance                   | ✅ Implementado |
| **DTOs Imutáveis**    | Parsing via fromArray()                                   | ✅ Implementado |
| **Strategy Pattern**  | Card vs Voucher (Single Source of Truth)                 | ✅ Implementado |
| **Async Movements**   | Jobs + Parent Wait + Idempotência                         | ✅ Implementado |
| **ISO8583**           | 50+ Data Elements parseados                               | ✅ Implementado |
| **Bloqueio Cartão**   | Status-based (3 = bloqueado)                              | ✅ Implementado |
| **Fraude**            | Code 59 (critérios não documentados)                      | ⚠️ Parcial     |

---

**Versão**: 1.3 (adicionada seção de Descrições de Extrato + bug do local identificado)  
**Última atualização**: 28/05/2026  
**Autor**: Gerado via análise do codebase Card API  
**Próxima revisão**: Após correção do bug de local em ChargeBack Cancel

---

## Feedback e Contribuição

Esta skill é **living documentation** - deve evoluir com o projeto.

**Como contribuir**:
1. Encontrou regra não documentada? → Adicione nova seção
2. Regra mudou? → Atualize seção correspondente
3. Implementou regra pendente? → Mova de ⚠️ para ✅
4. Código não reflete documentação? → Investigue e corrija

**Mantenha sincronizado** com código para servir como fonte confiável.
