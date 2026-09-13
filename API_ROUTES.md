# API do backend — guia para o front-end

Base: `http(s)://<host>/api`. Todas as rotas da versão atual começam por `/v1`.

## Autenticação e permissões

O login retorna o cookie HTTP-only seguro `auth` (validade de 3 dias). No navegador, envie as chamadas autenticadas com credenciais, por exemplo `credentials: 'include'` no `fetch`. Como alternativa, a API aceita `Authorization: Bearer <token>`.

| Perfil | Valor | Acesso |
| --- | ---: | --- |
| Administrador | `1` | Todas as rotas autenticadas |
| Usuário | `2` | Rotas de negócio, exceto qualquer `DELETE`; não acessa gerenciamento de usuários |

- `401`: cookie/token ausente, inválido, sessão revogada ou expirada.
- `403`: usuário inativo ou perfil sem permissão.
- Erros possuem o formato `{ "error": "mensagem" }`.
- Respostas de sucesso de operações simples possuem `{ "message": "mensagem" }` (e, quando criada uma entidade, podem conter `id`).

## Saúde

| Método | Rota | Autenticação | Resposta |
| --- | --- | --- | --- |
| `GET` | `/api/health` | Pública | `200 {"message":"ok"}` |

## Usuários

| Método | Rota | Permissão | Corpo / observações |
| --- | --- | --- | --- |
| `POST` | `/api/v1/user/login` | Pública | `{ "email": "...", "password": "..." }`. Retorna `200` e grava o cookie `auth`. |
| `POST` | `/api/v1/user/logout` | Admin ou usuário autenticado | Sem corpo. Retorna `200` e remove o cookie `auth`. |
| `POST` | `/api/v1/user` | Admin | `{ "nome":"...", "email":"...", "password":"...", "role":1 }`. `role`: `1` admin, `2` usuário. |
| `GET` | `/api/v1/users` | Admin | Retorna `{ "users": [{ "nome", "email", "role", "ativo" }] }`. |
| `PUT` | `/api/v1/user` | Admin | `{ "id": 1, "nome":"...", "email":"...", "role":2 }`. |
| `PUT` | `/api/v1/user/activate/:id` | Admin | Sem corpo. |
| `PUT` | `/api/v1/user/deactivate/:id` | Admin | Sem corpo. |

## Proprietários

Todos os endpoints abaixo aceitam admin e usuário autenticado.

| Método | Rota | Corpo / resposta |
| --- | --- | --- |
| `POST` | `/api/v1/proprietario` | `{ "nome":"...", "cpf":"12345678901", "observacao":"..." }`; CPF tem exatamente 11 caracteres. |
| `GET` | `/api/v1/proprietarios` | `{ "proprietarios": [{ "id", "nome", "cpf", "observacao", "ativo" }] }`. |
| `PUT` | `/api/v1/proprietario` | `{ "id":1, "nome":"...", "cpf":"12345678901", "observacao":"..." }`. |
| `PUT` | `/api/v1/proprietario/activate/:id` | Sem corpo. |
| `PUT` | `/api/v1/proprietario/deactivate/:id` | Sem corpo. |

## Fazendas

Todos os endpoints abaixo aceitam admin e usuário autenticado.

| Método | Rota | Corpo / resposta |
| --- | --- | --- |
| `POST` | `/api/v1/fazenda` | `{ "nome":"...", "cidade":"...", "inscricao_rural":"...", "observacao":"...", "id_proprietario":1 }`. |
| `GET` | `/api/v1/fazendas/:idProprietario` | Retorna `{ "fazendas": [{ "id", "nome", "cidade", "inscricao_rural", "observacao", "id_proprietario", "ativo" }] }`. |
| `PUT` | `/api/v1/fazenda` | `{ "id":1, "nome":"...", "cidade":"...", "inscricao_rural":"...", "observacao":"..." }`. |
| `PUT` | `/api/v1/fazenda/activate/:id` | Sem corpo. Pode retornar `409` se o proprietário estiver inativo. |
| `PUT` | `/api/v1/fazenda/deactivate/:id` | Sem corpo. |

## Abates

Todos aceitam admin e usuário autenticado, exceto os `DELETE`, que são exclusivos de admin. Datas usam `AAAA-MM-DD`.

| Método | Rota | Corpo / observações |
| --- | --- | --- |
| `POST` | `/api/v1/abate` | Cria um abate. Corpo completo abaixo; retorna `201 { "id", "message" }`. |
| `GET` | `/api/v1/abate/:id` | Retorna o abate completo. |
| `GET` | `/api/v1/abates` | Filtros opcionais: `proprietario_id`, `fazenda_id`, `numero_lote`, `data_inicio`, `data_fim`, `pagina` (padrão 1) e `limite` (padrão 50, máximo 100). Retorna `{ "abates": [], "pagina", "limite" }`. |
| `PUT` | `/api/v1/abate/:id/dados-gerais` | Corpo `dados_gerais` do exemplo abaixo. |
| `PUT` | `/api/v1/abate/:id/etapa-fazenda` | Corpo `etapa_fazenda` do exemplo abaixo. |
| `PUT` | `/api/v1/abate/:id/etapa-frigorifico` | Corpo `etapa_frigorifico` do exemplo abaixo. |
| `DELETE` | `/api/v1/abate/:id` | **Admin**. Sem corpo; retorna `204`. |
| `POST` | `/api/v1/abate/:id/fotos/:etapa` | `multipart/form-data`, campo de arquivo `foto`; `:etapa` deve ser `FAZENDA` ou `FRIGORIFICO`. Máximo de 10 MiB. Retorna `201` com `{ "id", "etapa", "nome_original", "content_type", "tamanho", "sha256" }`. |
| `GET` | `/api/v1/abate/fotos/:fotoID` | Baixa o arquivo, com `Content-Disposition: attachment`. |
| `DELETE` | `/api/v1/abate/fotos/:fotoID` | **Admin**. Sem corpo; retorna `204`. |

Exemplo para criar o abate (campos numéricos não podem ser negativos; IDs e lote devem ser positivos):

```json
{
  "dados_gerais": {
    "data_abate": "2026-09-13",
    "fazenda_id": 1,
    "numero_lote": 10,
    "nome_frigorifico": "Frigorífico Exemplo",
    "distancia_frigorifico": 25.5,
    "categoria_animal": "Bovino",
    "preco_funrural": 10.5,
    "preco_sem_funrural": 11.0
  },
  "etapa_fazenda": {
    "peso_total": 1000,
    "quantidade_animal": [{ "denticao": 2, "qtd_animais": 10 }]
  },
  "etapa_frigorifico": {
    "peso_total": 900,
    "balancao": 890,
    "acabamento_carcaca": [{ "acabamento": "3", "qtd_animais": 10 }],
    "classificacao_frigorifico": [{ "classificacao": "A", "qtd_animais": 10 }],
    "distribuicao_peso": [{ "classificacao": "A", "qtd_animais": 10, "peso_total": 900 }]
  }
}
```

O retorno de um abate contém `id`, `proprietario_id`, `nome_proprietario`, `nome_fazenda`, `dados_gerais`, `etapa_fazenda` e `etapa_frigorifico`. Ambas as etapas incluem `fotos`; cada foto tem `id`, `etapa`, `nome_original`, `content_type`, `tamanho` e `sha256`.

## Agenda

Todos aceitam admin e usuário autenticado, mas a agenda é associada ao usuário da sessão: as operações de leitura e alteração só alcançam seus próprios registros. Somente o admin pode excluir. `data_hora` deve estar em RFC 3339, por exemplo `2026-09-13T14:30:00Z`.

| Método | Rota | Corpo / observações |
| --- | --- | --- |
| `POST` | `/api/v1/agenda` | `{ "fazenda_id":1, "data_hora":"2026-09-13T14:30:00Z", "observacao":"..." }`; retorna `201 { "id", "message" }`. |
| `GET` | `/api/v1/agenda/:id` | Retorna `{ "id", "fazenda_id", "data_hora", "observacao" }`. |
| `GET` | `/api/v1/agendas` | Filtros opcionais: `fazenda_id`, `data_inicio`, `data_fim`, `pagina` (padrão 1) e `limite` (padrão 20, máximo 100). Retorna `{ "agendas": [], "pagina", "limite", "total" }`. |
| `PUT` | `/api/v1/agenda/:id` | `{ "fazenda_id":1, "data_hora":"2026-09-13T14:30:00Z", "observacao":"..." }`. |
| `DELETE` | `/api/v1/agenda/:id` | **Admin**. Retorna `200 { "message" }`. |

## Assinaturas Web Push

Todos exigem autenticação; o cadastro pertence ao usuário da sessão. A remoção é exclusiva de admin.

| Método | Rota | Corpo / observações |
| --- | --- | --- |
| `POST` | `/api/v1/push/subscriptions` | `{ "endpoint":"https://...", "expirationTime": 1760000000000, "keys": { "p256dh":"...", "auth":"..." } }`. `expirationTime` é opcional e usa Unix em milissegundos. Retorna `201 { "id", "message" }`. |
| `DELETE` | `/api/v1/push/subscriptions/:id` | **Admin**. Retorna `200 { "message" }`. |

## Códigos de resposta usuais

| Código | Significado |
| ---: | --- |
| `200` | Operação concluída. |
| `201` | Recurso criado. |
| `204` | Exclusão concluída sem corpo. |
| `400` | Corpo, parâmetro, filtro, data ou arquivo inválido. |
| `401` | Autenticação inválida ou ausente. |
| `403` | Sem permissão ou usuário inativo. |
| `404` | Recurso não encontrado. |
| `409` | Conflito de regra de negócio. |
| `500` | Erro interno. |
