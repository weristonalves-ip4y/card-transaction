---
name: senior-golang-sommelier
description: 'Implementa features de transacoes de cartao em Go com padrao senior. Use para criar handlers/services/repositories com separacao obrigatoria de responsabilidades, aplicar regras de negocio com validacoes explicitas, tratar erros com contexto e entregar no minimo 1 teste de caminho feliz e 1 teste de erro principal.'
argument-hint: '<feature ou endpoint a implementar>'
user-invocable: true
---

# Senior Golang Sommelier

## Quando Usar
- Implementar nova feature de transacoes de cartao em APIs, workers ou jobs.
- Aplicar obrigatoriamente o padrao em camadas handler -> service -> repository.
- Traduzir regra de negocio em codigo com validacoes explicitas.
- Entregar codigo com barra minima de teste e criterios de pronto.

## Entradas Esperadas
- Objetivo da feature no fluxo de cartao (compra, cancelamento, estorno, saque, ajuste).
- Contrato de entrada/saida (payload, status, codigos de resposta e erros esperados).
- Restricoes importantes (latencia, idempotencia, consistencia de saldo, compatibilidade).

## Checklist Rapida
1. Definir escopo exato da feature em 3-5 bullets antes de codar.
2. Escolher ponto de entrada (HTTP handler, consumer, cron/job) e mapear dependencias.
3. Modelar contrato: structs de request/response, validacoes e codigos de erro.
4. Implementar regra de negocio no service sem acoplamento a transporte.
5. Isolar I/O externo no repository/client com interfaces pequenas.
6. Garantir consistencia de saldo e idempotencia quando a operacao puder repetir.
7. Tratar erros com contexto (fmt.Errorf("...: %w", err)) sem perder causa raiz.
8. Criar no minimo 1 teste de caminho feliz e 1 teste de erro principal da feature.
9. Validar criterios de pronto (compila, testes passam, logs uteis, sem comportamento ambiguo).

## Decisoes-Chave (Branching)
- Se a feature altera contrato externo:
  - Preservar compatibilidade quando possivel.
  - Se breaking change for inevitavel, explicitar versao/migracao.
- Se a operacao movimenta saldo:
  - Aplicar validacoes de saldo no service antes de debito.
  - Garantir que creditos/debitos usem o valor transacional correto definido pela regra de negocio.
- Se a operacao pode repetir (retry/webhook/evento):
  - Definir estrategia de idempotencia antes da implementacao.
- Se ha mais de uma fonte de dados:
  - Centralizar orquestracao no service.
  - Evitar regra de negocio espalhada em repositories.

## Criterios de Qualidade
- Clareza de responsabilidades entre camadas.
- Erros com contexto e sem swallow silencioso.
- Nomes sem ambiguidades e sem siglas obscuras.
- Testes cobrindo no minimo caminho feliz e erro principal.
- Sem acoplamento desnecessario a framework/biblioteca.

## Definicao de Pronto
- Build local ok.
- Minimo de 2 testes por feature: 1 feliz + 1 erro principal.
- Contrato de entrada/saida consistente com implementacao.
- Principais caminhos de erro tratados.
- Diff pequeno, legivel e facil de revisar.

## Prompt de Exemplo
- `/senior-golang-sommelier implementar endpoint POST /transactions com validacao e persistencia`
- `/senior-golang-sommelier criar service de cancelamento com idempotencia e testes`
- `/senior-golang-sommelier adicionar regra de negocio de limite diario em saques`

### referencias

 use a pasta /references para obter todas as referenicas complementares