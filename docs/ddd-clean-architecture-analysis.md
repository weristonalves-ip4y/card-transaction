# Analise DDD e Clean Architecture para Card Transaction em Go

## Diagnostico da base PHP (resumo)
- Regras de negocio estao relativamente corretas e robustas.
- Ha violacao de fronteiras arquiteturais em partes do fluxo.
- Algumas validacoes e mapeamentos de resposta estao espalhados em servicos diferentes.
- Existe risco de drift de regra quando a mesma decisao aparece em multiplos pontos.

## Problemas arquiteturais observados
1. Regra de negocio em camada de orquestracao ou infraestrutura.
2. Mapeamento de response code acoplado ao fluxo especifico.
3. Repeticao de validacoes de produto/cartao.
4. Risco de inconsistencias no uso de valor transacional.

## Direcao adotada no Go
1. Dominio explicito
- Entidade Transaction concentra matriz Debito/Credito.
- Value Object Money usa centavos para evitar float.
- Entidade Card concentra estado minimo para validacao.

2. Aplicacao orquestra, dominio decide
- Pipeline de validacao executa ordem fixa.
- Regras de aprovacao/reprovacao vivem no dominio.

3. Decisao centralizada
- Tabela unica de codigo -> status HTTP -> mensagem.
- Evita hardcode de resposta por endpoint/servico.

## Estrutura criada
- Dominio:
  - internal/domain/money/money.go
  - internal/domain/card/card.go
  - internal/domain/transaction/transaction.go
- Aplicacao:
  - internal/application/validation/pipeline.go
  - internal/application/decision/table.go
  - internal/application/usecase/authorize_transaction.go
- Testes:
  - internal/domain/transaction/transaction_test.go
  - internal/application/validation/pipeline_test.go

## Regras preservadas
- Debitos: purchase, withdrawal, chargeback_cancel validam saldo.
- Creditos: cancel, chargeback_reversal nao exigem saldo suficiente.
- Ordem de validacao: duplicacao -> cartao -> produto -> limite mensal -> saldo.
- Valor monetario em centavos para calculo de negocio.

## Decisoes recomendadas para proxima iteracao
1. Definir tabela canonica final de response codes com negocio.
2. Implementar repositorios reais e adapters de persistencia.
3. Adicionar idempotencia no caso de uso.
4. Criar testes de contrato por fluxo (purchase, cancel, chargeback, withdrawal).
5. Separar comando de autorizacao e comando de movimentacao assincrona.
