# Gestão de Abate e Manejo — Backend

API REST para gestão de proprietários, fazendas, abates, agenda e notificações Web Push. O projeto é escrito em Go, segue uma organização por camadas/portas e adaptadores e usa PostgreSQL, armazenamento compatível com S3 (Cloudflare R2) e JWT em cookie HTTP-only.

## Funcionalidades

- Autenticação de usuários por JWT (`auth`) e controle de acesso por perfil.
- Cadastro e ativação de usuários, proprietários e fazendas.
- Registro, consulta, atualização e exclusão de abates, incluindo fotos.
- Agenda individual por usuário e notificações Web Push.
- Armazenamento de fotos em Cloudflare R2/S3 compatível.

## Perfis de acesso

| Perfil | Valor | Permissões |
| --- | ---: | --- |
| Administrador | `1` | Todas as operações autenticadas, inclusive exclusões e gerenciamento de usuários. |
| Usuário | `2` | Operações de negócio, exceto `DELETE`; não acessa as rotas de gerenciamento de usuários. |

`POST /api/v1/user/login` é público. `POST /api/v1/user/logout` exige autenticação, mas é permitido aos dois perfis.

## Tecnologias

- Go 1.27.1
- Gin
- PostgreSQL (`lib/pq`)
- JWT (`golang-jwt/jwt`)
- Cloudflare R2 / API compatível com S3 (AWS SDK v2)
- Web Push com VAPID
- Zap para logs

## Estrutura

```text
adapter/
  input/       # Controllers, middleware, rotas e modelos HTTP
  output/      # PostgreSQL, JWT, R2 e Web Push
application/
  domain/      # Entidades e regras de domínio
  port/        # Contratos de entrada e saída
  service/     # Casos de uso
configuration/ # Banco e logs
ddl.sql        # DDL disponível no repositório
```

## Pré-requisitos

- Go 1.27.1 ou compatível com o `go.mod`.
- PostgreSQL acessível.
- Bucket R2/S3 e credenciais com permissões de leitura, escrita e exclusão de objetos.
- Par de chaves VAPID para notificações Web Push.

## Configuração

O servidor carrega obrigatoriamente o arquivo `.env` na raiz. Crie-o localmente — ele é ignorado pelo Git — com valores equivalentes aos abaixo:

```dotenv
# Servidor
PORT=8080
GIN_MODE=debug

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=senha-local
DB_NAME=gestao_abate
DB_SSLMODE=disable

# Autenticação
JWT_SECRET=gere-uma-chave-longa-e-aleatoria

# Cloudflare R2 / S3 compatível
R2_BASE_URL=https://<account-id>.r2.cloudflarestorage.com
R2_ACCESS_KEY=<access-key>
R2_SECRET_KEY=<secret-key>
R2_BUCKET_NAME=<nome-do-bucket>

# Web Push
VAPID_PUBLIC_KEY=<chave-publica>
VAPID_PRIVATE_KEY=<chave-privada>
VAPID_SUBJECT=mailto:seu-email@exemplo.com

# Logs (opcionais)
LOG_OUTPUT=stdout
LOG_LEVEL=info
```

`DB_SSLMODE` assume `disable` quando não informado. As variáveis de R2, VAPID e `JWT_SECRET` são obrigatórias para a inicialização do servidor.

## Executar localmente

```powershell
go mod download
go run .
```

Com o exemplo de configuração, a API estará em `http://localhost:8080` e o health check em `GET /api/health`.

A geração dos relatórios PDF de abate usa `chromedp` e requer uma instalação do Chrome ou Chromium acessível no ambiente em que a API estiver em execução. Os valores permitidos para dentição, acabamento de carcaça, classificação do frigorífico e distribuição de peso estão documentados em [API_ROUTES.md](API_ROUTES.md).

Para desenvolvimento com recarregamento, o projeto também inclui a configuração [.air.toml](.air.toml), compatível com o Air:

```powershell
air
```

## Testes e qualidade

```powershell
go test ./...
go vet ./...
```

## Integração com o front-end

A documentação de todas as rotas, permissões, payloads, filtros, uploads e códigos de resposta está em [API_ROUTES.md](API_ROUTES.md).

O login grava um cookie `auth` com as flags `Secure` e `HttpOnly`. Em aplicações web, use chamadas com credenciais (`credentials: 'include'` no `fetch`) e sirva o front-end por HTTPS em ambientes onde o cookie seguro é necessário.

## Códigos HTTP mais comuns

| Código | Significado |
| ---: | --- |
| `200` | Operação concluída. |
| `201` | Recurso criado. |
| `204` | Exclusão concluída sem corpo. |
| `400` | Dados, filtro, data ou arquivo inválido. |
| `401` | Autenticação ausente ou inválida. |
| `403` | Usuário inativo ou sem permissão. |
| `404` | Recurso não encontrado. |
| `409` | Conflito de regra de negócio. |
| `500` | Erro interno. |
