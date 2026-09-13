# Guia de desenvolvimento para agentes

Este arquivo define o padrão a preservar neste repositório. Antes de alterar código, identifique a camada responsável pela regra e mantenha o escopo da mudança nela. Não altere ou reverta modificações preexistentes que não façam parte da tarefa.

## Arquitetura

O projeto segue portas e adaptadores (hexagonal), com responsabilidades separadas:

```text
adapter/
  input/                 # Entrada HTTP: routes, middleware, controllers, DTOs
  output/                # Infraestrutura: banco, JWT, R2 e integrações externas
application/
  domain/                # Entidades, regras e erros de domínio
  port/
    input/               # Interfaces dos casos de uso
    output/              # Interfaces de dependências externas
  service/               # Implementação dos casos de uso
configuration/           # Bootstrap e configurações transversais
```

- `domain` não pode depender de Gin, SQL, HTTP ou detalhes de infraestrutura.
- `service` coordena regras de negócio por meio de interfaces em `application/port/output`.
- `adapter/input` converte HTTP em chamadas de caso de uso e converte resultados em HTTP; não concentra regra de negócio.
- `adapter/output` implementa ports para banco, autenticação, armazenamento e integrações.
- A composição de dependências fica em `main.go`; não instancie repositórios ou clientes concretos dentro de controllers ou services.

## Arquivos e nomenclatura

- Use nomes em `kebab-case` e sufixo por responsabilidade: `user-service.go`, `user-controller.go`, `user-repository.go`, `user-port.go`, `user-request.go` e `user-response.go`.
- Mantenha arquivos de teste ao lado do código: `nome_test.go`.
- Um arquivo deve ter responsabilidade coesa; prefira extrair helpers privados para o arquivo/camada apropriados em vez de criar arquivos genéricos sem contexto.
- Use nomes de pacote em minúsculas e sem underscore: `service`, `controller`, `repository`, `request`, `response`.
- Use PascalCase para tipos exportados e camelCase para identificadores não exportados.
- Siglas em tipos exportados usam o padrão idiomático Go: `JWT`, `URL`, `ID`, `CPF`, `R2`; exemplos: `UserID`, `R2Storage`, `JWTClaims`.
- Nomeie entidades e DTOs de forma explícita: `User`, `Fazenda`, `AbateCreateRequest`, `GetAgendasResponse`.

## Domain

- Entidades, constantes de negócio e erros pertencem a `application/domain`.
- Declare erros reutilizáveis como `var Err... = errors.New(...)`; use `errors.Is` nas camadas que traduzem esses erros.
- Não vaze tipos de request/response, `gin.Context`, `*sql.DB` ou clientes de SDK para o domínio.
- Mantenha valores enumerados (por exemplo, papéis e estados) como constantes nomeadas, e valide-os nas fronteiras de entrada adequadas.

## Ports e services

- Interfaces de casos de uso ficam em `application/port/input`; interfaces de dependências externas ficam em `application/port/output`.
- Interfaces devem expressar a necessidade do caso de uso, não expor detalhes de uma biblioteca ou banco. Receba interfaces nos construtores dos services.
- Services têm o formato `XxxService`, construtor `NewXxxService` e implementam o respectivo use case.
- Centralize regras, validações de negócio, autorização de propriedade e mapeamento de erros previsíveis nos services/domínio.
- Receba `context.Context` em operações que realizem I/O ou possam ser canceladas, propagando o contexto até o adapter de saída.
- Não use `panic` para erros esperados de negócio; retorne `error`.

## Controllers, DTOs e rotas

- Controllers têm o formato `XxxController`, construtor `NewXxxController` e dependem apenas de interfaces de input.
- Controllers devem: fazer bind e validação sintática do request, montar entidades/argumentos, chamar o use case e escrever a resposta HTTP.
- Requests ficam em `adapter/input/model/request` e responses em `adapter/input/model/response`. Use tags JSON em `snake_case` e tags `binding` do Gin para validações HTTP.
- Não retorne entidades de domínio diretamente quando houver dados a ocultar, formatar ou estabilizar; use DTOs de response.
- Rotas ficam em `adapter/input/route`. Registre middleware no grupo mais restrito possível e preserve a matriz de autorização existente.
- Use códigos HTTP consistentes: `200` leitura/alteração, `201` criação, `204` exclusão sem resposta, `400` entrada inválida, `401` sem autenticação, `403` sem permissão, `404` ausente, `409` conflito e `500` falha inesperada.

## Adapters de saída e banco

- Implemente persistência em `adapter/output/repository`, usando `database/sql` e `*sql.DB` como padrão de conexão e execução de queries.
- Passe `context.Context` a queries de novas operações (`QueryContext`, `QueryRowContext`, `ExecContext`) quando o port permitir.
- Sempre use parâmetros posicionais; nunca concatene dados de usuário em SQL.
- Feche `rows` e verifique `rows.Err()`. Use transações para operações que precisam ser atômicas, com rollback seguro.
- Não coloque SQL em services, controllers ou domain.
- Mantenha mapeamento entre linha SQL e domínio no repositório e traduza somente erros de infraestrutura que a camada superior precisa conhecer.

## Logs: Zap é prioritário

- Use exclusivamente `configuration/logger` (baseado em Zap) para logs da aplicação: `logger.Info(...)` e `logger.Error(...)`.
- Registre contexto útil da operação e o erro original nas falhas inesperadas; evite logs duplicados para o mesmo erro em múltiplas camadas.
- Nunca registre senha, JWT, cookie `auth`, cabeçalho `Authorization`, chaves de API, credenciais de banco, VAPID ou dados sigilosos.
- Não use `fmt.Println`, `log.Print` ou logging ad hoc em código de produção.

## Chamadas HTTP externas: Resty é prioritário

- Para novas chamadas HTTP de saída, use `github.com/go-resty/resty/v2` como cliente padrão, encapsulado em um adapter de saída e protegido por uma interface de `application/port/output`.
- Configure timeout, contexto, limites de retry quando apropriado e valide códigos HTTP antes de desserializar respostas.
- Não faça chamadas HTTP diretamente em controllers, services ou domain.
- Não registre headers ou corpos que possam conter segredos. Converta falhas externas em erros claros para a camada de serviço.
- Antes de importar Resty, confirme que a tarefa realmente precisa de uma integração HTTP; quando necessário, adicione a dependência com `go get` e atualize `go.mod`/`go.sum` de modo deliberado.

## Qualidade e validação

- Formate arquivos alterados com `gofmt`.
- Inclua ou atualize testes para comportamentos novos e correções, especialmente regras de autorização, validação e erros de negócio.
- Antes de concluir uma alteração Go, execute `go test ./...` e `go vet ./...`, salvo bloqueio externo devidamente reportado.
- Execute `git diff --check` para detectar problemas de whitespace.
- Preserve compatibilidade de API quando não houver autorização explícita para quebrá-la; documente mudanças de rota, payload e permissão em `API_ROUTES.md` e, quando relevante, em `README.md`.
