# PHP -> Go: Business Rules Abstraction

## Fonte primaria
- Skill de dominio no projeto PHP: `/root/projects/ip4y-card-pay-smart-api/.github/skills/business-rules/SKILL.md`

## Objetivo
Traduzir regras de negocio do projeto PHP (Card Pay Smart API) para requisitos implementaveis no projeto Go, preservando comportamento financeiro e codigos de resposta.

## Regras criticas que NAO podem quebrar

### 1) Valor transacional correto em operacoes de saldo
- Debito/credito de saldo deve usar valor transacional completo (equivalente a `totalAmount`).
- Campos de exibicao (equivalentes a `cancelAmount` e `chargeAmount`) nao devem ser usados para calculo de saldo.

Impacto em Go:
- Criar um tipo de valor monetario unico para calculo (`TransactionValue`).
- Expor valores de exibicao separadamente sem permitir uso acidental na conta.

### 2) Regra de sinal por tipo de operacao
- Purchase: debita saldo.
- Withdrawal: debita saldo.
- Cancel: credita saldo.
- Chargeback reversal: credita saldo (inversao de sinal).
- Chargeback cancel: debita saldo (anula credito anterior).

Impacto em Go:
- Centralizar sinal em uma camada de dominio (nao duplicar em handlers/jobs).
- Strategy/Policy por tipo de movimento para evitar inconsistencias.

### 3) Validacao de saldo insuficiente em operacoes de debito
- Operacoes de debito devem validar saldo antes de subtrair.
- Codigo de resposta padrao para saldo insuficiente: `01`.

Impacto em Go:
- Service de dominio deve seguir sequencia: validar -> aplicar movimento -> persistir.

### 4) Card vs Voucher altera processamento
- Tipo definido por `ps_product_code`.
- Decide strategy de movimento e mecanismo de persistencia/execucao.

Impacto em Go:
- Introduzir `ProductType` (Card/Voucher) e resolver strategy por esse tipo.

### 5) Ordem de validacoes
Ordem esperada:
1. Duplicacao
2. Cartao (status/expiracao/restricoes)
3. Produto compativel
4. Limite mensal
5. Saldo

Impacto em Go:
- Implementar pipeline de validacao na mesma ordem.
- Evitar mudar ordem sem decisao explicita de negocio.

### 6) Limite mensal
- Se limite estiver ativo, deve bloquear quando a soma do mes + valor atual excede limite.
- Pode existir bypass por flags de negocio (ex.: force accept).

Impacto em Go:
- Introduzir politica de limite mensal com repositorio/fonte de soma mensal.

### 7) Movimento assincrono
- Movimentos financeiros sao processados assincronamente apos aprovacao.
- Cancel/chargeback podem depender de transacao pai e ordem de processamento.

Impacto em Go:
- Modelar estado de movimento (`pending/completed/failed`) e idempotencia no worker.

### 8) Invalidadacao de cache de saldo
- Sempre invalidar cache de saldo apos operacao que altera saldo.

Impacto em Go:
- Definir contrato explicito de cache invalidation no service de saldo.

## Mapeamento recomendado para arquitetura Go

### Camadas
- Handler: parse/validacao estrutural do request, mapeia resposta.
- Service (dominio): aplica regras de negocio e orquestra fluxo.
- Repository/Gateway: leitura/escrita de dados e integracao externa.

### Componentes de dominio sugeridos
- `Money` (inteiro em centavos; sem float para calculo).
- `TransactionType` (purchase, withdrawal, cancel, chargeback_reversal, chargeback_cancel).
- `MovementDirection` (debit/credit).
- `ProductType` (card/voucher).
- `ValidationPipeline` com ordem fixa.
- `MovementStrategy` por produto.

### Contratos minimos
- `BalanceRepository`: get balance, persist movement, get monthly sum.
- `IdempotencyRepository`: lock/check por chave de negocio.
- `TransactionRepository`: gravar transacao e vinculo com parent.
- `Cache`: invalidate balance por conta/cartao.

## Matriz de comportamento esperada

| Operacao | Direcao | Valida saldo antes? | Efeito saldo |
|---|---|---|---|
| Purchase | Debito | Sim | -valor |
| Withdrawal | Debito | Sim | -valor |
| Cancel | Credito | Nao | +valor |
| Chargeback Reversal | Credito | Nao | +valor |
| Chargeback Cancel | Debito | Sim | -valor |

## Criterios de aceite para portar para Go
- Mesmos codigos de resposta para os mesmos cenarios de negocio.
- Nenhuma operacao de saldo usando valor de exibicao.
- Regra de sinal validada por testes de dominio para todos os tipos.
- Pipeline de validacao mantendo a ordem acordada.
- Testes minimos por fluxo: 1 caminho feliz + 1 erro principal.

## Lacunas a confirmar no codigo PHP
- Regras exatas de fraude (code `59`) e gatilhos de bloqueio.
- Detalhes finais de MCC (whitelist/blacklist) se houver implementacao parcial.
- Formato final de descricoes de extrato para todos os subtipos de chargeback.

## Proximo passo recomendado
Ler os services no PHP e gerar uma tabela "Regra -> Arquivo/Metodo -> Comportamento" para fechar gaps antes de implementar definitivamente no Go.

## Validacoes auditadas no codigo PHP (resultado)

### Fluxo orquestrado de validacao (confirmado)
Sequencia confirmada no fluxo principal:
1. Duplicacao
2. Cartao
3. Compatibilidade de produto
4. Limite mensal
5. Saldo

Decisao para Go:
- Implementar pipeline explicito e imutavel de validacao nessa ordem.

### Regras de cartao (confirmadas)
- Cartao inexistente, sem conta vinculada, bloqueado/inativo retornam negacao de validacao.
- Compatibilidade de produto valida Card x Voucher.

Decisao para Go:
- Centralizar em validador de cartao unico.
- Nao duplicar validacao de produto em mais de um ponto do fluxo.

### Limite mensal e saldo (confirmados)
- Limite mensal pode ser ignorado por flag de bypass.
- Limite mensal bloqueia com codigo de saldo/limite.
- Saldo insuficiente bloqueia com codigo de negacao.

Decisao para Go:
- Manter bypass explicito por policy de risco (nao por if espalhado).
- Avaliar modelagem de limite e saldo no mesmo modulo de risco, mas com regras separadas.

### Cancelamento e chargeback (confirmados)
- Cancelamento e estorno (reversal) operam como credito no saldo final.
- Cancelamento de estorno (chargeback cancel) volta a debitar.
- Em chargeback cancel existe validacao de saldo antes de aprovar o debito.

Decisao para Go:
- Formalizar matriz Debito/Credito por tipo de operacao e forcar por testes de dominio.

### Divergencias de codigo de resposta (ponto critico)
- O material de regra descreve alguns codigos historicos.
- A implementacao auditada usa mapeamentos diferentes em pontos de cartao/fraude.

Decisao para Go:
- Definir uma tabela canonica de response codes antes de portar.
- Separar claramente:
	- codigo ISO de negocio
	- status HTTP
	- mensagem para cliente

## Melhores decisoes arquiteturais para as validacoes no Go

1. Criar um modulo `validation` com pipeline declarativo e ordem fixa.
2. Criar `DecisionTable` central para response codes e mensagens, sem hardcode por fluxo.
3. Usar `Money` em centavos em todas as validacoes (sem float).
4. Separar validacao sincrona (request path) de efeitos assincronos (movimento).
5. Tornar idempotencia obrigatoria para purchase, withdrawal, cancel e chargeback.
6. Exigir testes minimos por validador: 1 feliz + 1 erro principal.
7. Adicionar testes de regressao para matriz Debito/Credito e para ordem do pipeline.

## Checklist de decisao antes de implementar
- Tabela canonica de response codes aprovada pelo negocio.
- Definicao oficial de quando retornar balance no payload de erro.
- Regra de bypass (force accept) formalizada e auditavel.
- Regra final de MCC definida (ativa, parcial ou fora de escopo).
- Politica de fraude definida (ou explicitamente adiada com feature flag).

## Foco final: Performance

Motivacao principal da migracao:
- Reduzir latencia no caminho critico de autorizacao.
- Reduzir carga no banco por excesso de leituras sequenciais.
- Preservar regra de negocio com menor custo por transacao.

Decisoes de performance adotadas no Go:
1. Validacao em estagios com fast-fail:
	- duplicacao
	- cartao/produto
	- limite mensal (somente quando aplicavel)
	- saldo (somente operacoes de debito)
2. Carregamento lazy de dados de risco:
	- nao consulta saldo para operacoes de credito
	- nao consulta soma mensal para operacoes que nao exigem limite mensal
3. Dominio centralizado para evitar retrabalho e ramificacoes duplicadas.
4. Value Object monetario em centavos para eliminar custo e erro de float.

Impacto esperado:
- Menos round-trips ao banco por request.
- Menor p95 e p99 em autorizacao.
- Menor risco de lock/contencao em pico.

Proximo passo orientado a performance:
- Criar um adapter de repositorio com query consolidada para carregar snapshot de validacao (card, limite mensal agregado, saldo) em 1 chamada quando fizer sentido operacional.
