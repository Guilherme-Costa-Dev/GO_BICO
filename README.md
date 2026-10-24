# Bico API

O Bico API é o backend REST da plataforma que conecta clientes a prestadores de serviços. É escrito em Go e usa o Firebase como infraestrutura: **Cloud Firestore** para persistência e **Firebase Authentication** para credenciais.

## Tecnologias

- **Linguagem:** Go (Golang)
- **Banco de dados:** Google Cloud Firestore (NoSQL)
- **Autenticação:** Firebase Authentication (JWT)
- **Geocodificação:** BrasilAPI (`https://brasilapi.com.br/api/cep/v2/{cep}`)

## Autenticação

A maioria das rotas é protegida pelo middleware `Auth`. Envie o token JWT do Firebase no header:

```http
Authorization: Bearer <SEU_TOKEN_JWT>
```

- O middleware valida o token com `VerifyIDToken`.
- O UID do usuário autenticado fica disponível no contexto (`userUID`).
- Nas rotas do próprio usuário, **o UID vem do token** — não é necessário (nem aceito) enviar o ID no corpo/URL.
- Sem token, formato inválido ou token expirado: **401 Unauthorized** com mensagem em texto puro.

Rotas **públicas** (sem token): `POST /clientes` e `POST /prestadores`. Todas as demais exigem autenticação.

## Convenções da API

- Formato: `application/json`.
- O UID do Firebase é o ID do documento nas coleções `clientes` e `prestadores`.
- Erros usam `http.Error`, portanto o corpo é **texto puro** (não JSON).
- Datas são serializadas no padrão RFC 3339 (ex.: `YYYY-MM-DDTHH:MM:SSZ`).

## Modelos de domínio (Firestore)

### `UserBase` (campos comuns a cliente e prestador)

| Campo | Tipo | Firestore | Descrição |
|---|---|---|---|
| `id` | string | — | UID do Firebase (não é salvo como campo) |
| `nome` | string | `nome` | Nome do usuário |
| `cpf` | string | `cpf` | CPF |
| `cep` | string | `cep` | CEP usado para geocodificação |
| `lat` | float64 | `lat` | Latitude |
| `lng` | float64 | `lng` | Longitude |
| `email` | string | `email` | E-mail |

### `Cliente`

Além de `UserBase`:

| Campo | Tipo | Firestore |
|---|---|---|
| `numero` | string | `numero` |
| `complemento` | string | `complemento` |
| `favoritos` | []string | `favoritos` (UIDs de prestadores) |

### `Prestador`

Além de `UserBase`:

| Campo | Tipo | Firestore |
|---|---|---|
| `username` | string | `username` |
| `tiposServico` | []string | `tiposServico` |
| `localAtuacao` | string | `localAtuacao` |
| `fotoPerfil` | string | `fotoPerfil` |
| `fotoPaginaPerfil` | string | `fotoPaginaPerfil` |
| `fotosServicos` | []string | `fotosServicos` |
| `sobre` | string | `sobre` |
| `notaMedia` | float64 | `notaMedia` |
| `avaliacoes` | []Avaliacao | `avaliacoes` |

### `Avaliacao`

| Campo | Tipo | Descrição |
|---|---|---|
| `nome` | string | Nome do cliente que avaliou |
| `nota` | int | Nota de **0 a 5** |
| `comentario` | string | Comentário |
| `data` | timestamp | Data/hora da avaliação |

## Endpoints

### Resumo

| Método | Rota | Auth | Recebe | Devolve |
|---|---|---|---|---|
| POST | `/clientes` | Não | `CriarClienteDTO` (JSON) | Cliente criado (201) |
| GET | `/clientes` | Sim | — | Cliente autenticado (200) |
| PUT | `/clientes` | Sim | `AtualizarClienteDTO` (JSON, parcial) | Status (200) |
| DELETE | `/clientes` | Sim | — | Status (200) |
| PUT | `/clientes/favoritos?id=` | Sim | Query `id` | Status (200) |
| GET | `/clientes/favoritos` | Sim | — | `{ "favoritos": [PrestadorPublicoDTO] }` |
| DELETE | `/clientes/favoritos?id=` | Sim | Query `id` | Status (200) |
| PUT | `/clientes/avaliacoes` | Sim | `AdicionarAvaliacaoDTO` (JSON) | Status (200) |
| PUT | `/clientes/email?email=` | Sim | Query `email` | Status (200) |
| POST | `/prestadores` | Não | `CriarPrestadorDTO` (JSON) | Prestador criado (201) |
| GET | `/prestadores` | Sim | — | Prestador autenticado (200) |
| PUT | `/prestadores` | Sim | `AtualizarPrestadorDTO` (JSON, parcial) | Status (200) |
| DELETE | `/prestadores` | Sim | — | Status (200) |
| GET | `/prestadores/busca` | Sim | Query params | `[PrestadorPublicoDTO]` |
| PUT | `/prestadores/email?email=` | Sim | Query `email` | Status (200) |

---

## Clientes

### 1. Criar cliente

`POST /clientes` — **público**

Cria o usuário no Firebase Auth, geocodifica o CEP e salva o documento na coleção `clientes`.

**Recebe (JSON) — `CriarClienteDTO`:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `nome` | string | sim | Nome do cliente |
| `cpf` | string | sim | CPF |
| `cep` | string | sim | CEP (usado para buscar latitude/longitude) |
| `email` | string | sim | E-mail (usado no Firebase Auth) |
| `senha` | string | sim | Senha (usada apenas no Firebase Auth) |
| `numero` | string | não | Número do endereço |
| `complemento` | string | não | Complemento do endereço |

```json
{
  "nome": "string",
  "cpf": "string",
  "cep": "string",
  "email": "string",
  "senha": "string",
  "numero": "string",
  "complemento": "string"
}
```

**Devolve:** `201 Created` com o `Cliente` criado:

```json
{
  "id": "<uid>",
  "nome": "string",
  "cpf": "string",
  "cep": "string",
  "lat": 0,
  "lng": 0,
  "email": "string",
  "numero": "string",
  "complemento": "string",
  "favoritos": []
}
```

**Erros:** `400` JSON inválido; `500` falha ao geocodificar o CEP, criar o usuário ou salvar no Firestore.

---

### 2. Buscar cliente autenticado

`GET /clientes` — **requer token**

**Recebe:** nada (o UID vem do token).

**Devolve:** `200 OK` com o `Cliente` do usuário autenticado (mesmo formato acima).

**Erros:** `401` token ausente/inválido; `404` cliente não encontrado.

---

### 3. Atualizar cliente

`PUT /clientes` — **requer token**

Atualização **parcial**: envie somente os campos que deseja alterar. Se `cep` for enviado, as coordenadas são recalculadas.

**Recebe (JSON) — `AtualizarClienteDTO`:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `nome` | string | não | Nome |
| `cpf` | string | não | CPF |
| `cep` | string | não | CEP (dispara nova geocodificação) |
| `numero` | string | não | Número |
| `complemento` | string | não | Complemento |

```json
{
  "nome": "string",
  "numero": "string"
}
```

**Devolve:** `200 OK`

```json
{ "status": "Cliente atualizado com sucesso" }
```

**Erros:** `400` JSON inválido; `401` token; `404` cliente não encontrado; `500` falha ao geocodificar ou salvar.

---

### 4. Deletar cliente

`DELETE /clientes` — **requer token**

Remove o documento da coleção `clientes` e a conta no Firebase Auth.

**Recebe:** nada.

**Devolve:** `200 OK`

```json
{ "status": "Cliente deletado com sucesso" }
```

**Erros:** `401` token; `500` falha ao deletar no Firestore ou no Auth.

---

### 5. Adicionar favorito

`PUT /clientes/favoritos?id=<prestadorID>` — **requer token**

Adiciona o UID do prestador ao array `favoritos` do cliente (sem duplicar, via `ArrayUnion`).

**Recebe:** query param `id` (UID do prestador).

**Devolve:** `200 OK`

```json
{ "status": "Favorito adicionado com sucesso" }
```

**Erros:** `400` `id` não fornecido; `401` token; `500` falha ao atualizar.

---

### 6. Listar favoritos

`GET /clientes/favoritos` — **requer token**

Busca os documentos dos prestadores favoritados e devolve o perfil público de cada um.

**Recebe:** nada.

**Devolve:** `200 OK`

```json
{
  "favoritos": [
    {
      "id": "<uid>",
      "nome": "string",
      "tiposServico": [],
      "notaMedia": 0,
      "email": "string",
      "localAtuacao": "string",
      "username": "string",
      "fotoPerfil": "string",
      "fotoPaginaPerfil": "string",
      "fotosServicos": [],
      "sobre": "string",
      "totalAvaliacoes": 0,
      "avaliacoes": [],
      "distanciaKm": 0
    }
  ]
}
```

**Erros:** `401` token; `500` falha ao buscar cliente ou prestadores.

---

### 7. Deletar favorito

`DELETE /clientes/favoritos?id=<prestadorID>` — **requer token**

Remove o UID do prestador do array `favoritos` (`ArrayRemove`).

**Recebe:** query param `id`.

**Devolve:** `200 OK`

```json
{ "status": "Favorito removido com sucesso" }
```

**Erros:** `400` `id` não fornecido; `401` token; `500` falha ao atualizar.

---

### 8. Adicionar avaliação

`PUT /clientes/avaliacoes` — **requer token**

Registra a avaliação do cliente autenticado em um prestador. O nome do avaliador é obtido do documento do cliente (não do corpo). A `notaMedia` do prestador é recalculada.

**Recebe (JSON) — `AdicionarAvaliacaoDTO`:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `prestadorID` | string | sim | UID do prestador avaliado |
| `nota` | int | sim | Nota de **0 a 5** |
| `comentario` | string | não | Comentário |

```json
{
  "prestadorID": "<uid>",
  "nota": 0,
  "comentario": "string"
}
```

**Devolve:** `200 OK`

```json
{ "status": "Avaliacao adicionada com sucesso" }
```

**Erros:** `400` JSON inválido, `prestadorID` vazio ou nota fora de 0–5; `401` token; `500` falha ao buscar cliente/prestador ou salvar a avaliação.

---

### 9. Alterar e-mail do cliente

`PUT /clientes/email?email=<novoEmail>` — **requer token**

Atualiza o e-mail no Firebase Auth e no Firestore. Se a atualização no Firestore falhar, tenta reverter o e-mail no Auth.

**Recebe:** query param `email`.

**Devolve:** `200 OK`

```json
{ "status": "Email alterado com sucesso" }
```

**Erros:** `400` e-mail não fornecido; `401` token; `500` falha no Auth/Firestore/rollback.

---

## Prestadores de Serviço

### 1. Criar prestador

`POST /prestadores` — **público**

Cria o usuário no Firebase Auth, geocodifica o CEP e salva o documento na coleção `prestadores`. As fotos de serviço começam como lista vazia.

**Recebe (JSON) — `CriarPrestadorDTO`:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `nome` | string | sim | Nome |
| `cpf` | string | sim | CPF |
| `cep` | string | sim | CEP (usado para latitude/longitude) |
| `email` | string | sim | E-mail (Firebase Auth) |
| `senha` | string | sim | Senha (apenas Firebase Auth) |
| `username` | string | sim | Nome de usuário público |
| `tiposServico` | []string | sim | Tipos de serviço |
| `localAtuacao` | string | sim | Local de atuação |
| `sobre` | string | não | Texto de apresentação |

```json
{
  "nome": "string",
  "cpf": "string",
  "cep": "string",
  "email": "string",
  "senha": "string",
  "username": "string",
  "tiposServico": [],
  "localAtuacao": "string",
  "sobre": "string"
}
```

**Devolve:** `201 Created` com o `Prestador` criado (sem `senha`):

```json
{
  "id": "<uid>",
  "nome": "string",
  "cpf": "string",
  "cep": "string",
  "lat": 0,
  "lng": 0,
  "email": "string",
  "username": "string",
  "tiposServico": [],
  "localAtuacao": "string",
  "fotoPerfil": "string",
  "fotoPaginaPerfil": "string",
  "fotosServicos": [],
  "sobre": "string",
  "notaMedia": 0,
  "avaliacoes": null
}
```

**Erros:** `400` JSON inválido; `500` falha ao geocodificar, criar usuário ou salvar.

---

### 2. Buscar prestador autenticado

`GET /prestadores` — **requer token**

**Recebe:** nada (UID do token).

**Devolve:** `200 OK` com o `Prestador` autenticado (mesmo formato acima).

**Erros:** `401` token; `404` prestador não encontrado.

---

### 3. Atualizar prestador

`PUT /prestadores` — **requer token**

Atualização **parcial** (envie só os campos desejados). Se `cep` for enviado, o backend geocodifica novamente.

**Recebe (JSON) — `AtualizarPrestadorDTO`:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `nome` | string | não | Nome |
| `tiposServico` | []string | não | Tipos de serviço |
| `localAtuacao` | string | não | Local de atuação |
| `username` | string | não | Username |
| `cep` | string | não | CEP (dispara geocodificação) |
| `fotoPerfil` | string | não | URL da foto de perfil |
| `fotoPaginaPerfil` | string | não | URL da foto da página |
| `fotosServicos` | []string | não | URLs das fotos de serviços |
| `sobre` | string | não | Texto de apresentação |

```json
{
  "nome": "string",
  "tiposServico": [],
  "localAtuacao": "string",
  "username": "string",
  "cep": "string",
  "fotoPerfil": "string",
  "fotoPaginaPerfil": "string",
  "fotosServicos": [],
  "sobre": "string"
}
```

**Devolve:** `200 OK`

```json
{ "status": "Prestador atualizado com sucesso" }
```

**Erros:** `400` JSON inválido; `401` token; `404` prestador não encontrado; `500` falha ao geocodificar/salvar.

---

### 4. Deletar prestador

`DELETE /prestadores` — **requer token**

Remove o documento da coleção `prestadores` e a conta no Firebase Auth.

**Recebe:** nada.

**Devolve:** `200 OK`

```json
{ "status": "Prestador deletado com sucesso" }
```

**Erros:** `401` token; `500` falha ao deletar.

---

### 5. Buscar prestadores

`GET /prestadores/busca` — **requer token**

Retorna uma lista pública de prestadores. Sem `busca=proximos`, ordena por `notaMedia` (decrescente). Com `busca=proximos`, calcula a distância em relação ao cliente autenticado, filtra pelo raio e ordena do mais próximo para o mais distante.

**Recebe (query params) — `BuscarPrestadorDTO`:**

| Param | Tipo | Padrão | Descrição |
|---|---|---|---|
| `username` | string | — | Filtro exato por username (case-sensitive) |
| `tipo` | string | — | Filtro por um tipo de serviço (`array-contains`) |
| `inicio` | int | `0` | Offset da paginação |
| `limite` | int | `20` | Itens por página (máximo `100`) |
| `busca` | string | — | Use `proximos` para busca por distância |
| `raio` | float | `10` | Raio em km (somente com `busca=proximos`) |

Exemplos:

```http
GET /prestadores/busca?tipo=<tipo>&inicio=0&limite=20
Authorization: Bearer <token>
```

```http
GET /prestadores/busca?busca=proximos&raio=<raio>&inicio=0&limite=20
Authorization: Bearer <token>
```

> Em `busca=proximos`, o cliente é identificado pelo **UID do token**; o cliente precisa ter `lat`/`lng` (obtidos no cadastro via CEP).

**Devolve:** `200 OK` com um array de `PrestadorPublicoDTO`:

```json
[
  {
    "id": "<uid>",
    "nome": "string",
    "tiposServico": [],
    "notaMedia": 0,
    "email": "string",
    "localAtuacao": "string",
    "username": "string",
    "fotoPerfil": "string",
    "fotoPaginaPerfil": "string",
    "fotosServicos": [],
    "sobre": "string",
    "totalAvaliacoes": 0,
    "avaliacoes": [
      {
        "nome": "string",
        "nota": 0,
        "comentario": "string",
        "data": "<timestamp RFC 3339>"
      }
    ],
    "distanciaKm": 0
  }
]
```

> `distanciaKm` é `0` quando a busca **não** é por proximidade.

**Erros:** `400` `inicio`/`limite`/`raio` inválidos, `busca=proximos` sem localização do cliente; `401` token; `500` falha na consulta.

---

### 6. Alterar e-mail do prestador

`PUT /prestadores/email?email=<novoEmail>` — **requer token**

Atualiza o e-mail no Firebase Auth e no Firestore. Se o Firestore falhar, tenta reverter o Auth.

**Recebe:** query param `email`.

**Devolve:** `200 OK`

```json
{ "status": "Email alterado com sucesso" }
```

**Erros:** `400` e-mail não fornecido; `401` token; `500` falha no Auth/Firestore/rollback.

---

## DTOs de resposta

### `PrestadorPublicoDTO`

Usado em `GET /prestadores/busca` e `GET /clientes/favoritos`.

| Campo | Tipo |
|---|---|
| `id` | string |
| `nome` | string |
| `tiposServico` | []string |
| `notaMedia` | float64 |
| `email` | string |
| `localAtuacao` | string |
| `username` | string |
| `fotoPerfil` | string |
| `fotoPaginaPerfil` | string |
| `fotosServicos` | []string |
| `sobre` | string |
| `totalAvaliacoes` | int |
| `avaliacoes` | []Avaliacao |
| `distanciaKm` | float64 |
