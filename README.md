# Bico API

O Bico API é o backend REST responsável por gerenciar a plataforma de conexão entre clientes e prestadores de serviços. Construído em Go (Golang), ele utiliza o Firebase como solução de nuvem, integrando o Firestore para banco de dados NoSQL e o Firebase Authentication para o gerenciamento de credenciais dos usuários.

## Tecnologias Utilizadas
- Linguagem: Go (Golang)
- Banco de Dados: Google Cloud Firestore  
- Autenticação: Firebase Auth

## Autenticação
A maioria dos endpoints da aplicação são protegidos por um middleware de autenticação. Para acessá-los, você deve enviar o token JWT gerado pelo Firebase no lado do cliente no cabeçalho da requisição:

**Extração Automática de Identidade (UID):**  
Em todas as rotas protegidas que manipulam os dados do próprio usuário logado, **o ID do cliente ou prestador é extraído automaticamente do Token de Autenticação**.

## Modelos de Dados (Domain)
O sistema baseia-se em uma estrutura comum `UserBase` que contém ID, Nome, Cpf, Email e Senha.  
- **Cliente**: Herda de UserBase e adiciona campos de endereço (Cep, Numero, Complemento) e uma lista de favoritos (`favorito`).  
- **Prestador**: Herda de UserBase e adiciona o portfólio profissional, incluindo Username, TiposServico, LocalAtuacao, links para fotos (FotoPerfil, FotoPaginaPerfil, FotosServicos), um texto Sobre, NotaMedia e TotalAvaliacoes.

## Endpoints
Todas as respostas da API são devolvidas no formato `application/json`. O UID gerado pelo Firebase Auth é utilizado como o ID principal do documento nas coleções do Firestore.

### Clientes

1. **Criar Cliente**
   * **Rota:** `POST /clientes`
   * **Autenticação:** Não requerida
   * **Descrição:** Cria as credenciais no Firebase Auth e salva os dados na coleção `clientes` do Firestore.
   * **Body (JSON):** Deve conter os campos definidos no modelo `Cliente` (incluindo a `senha` para o Auth).

2. **Buscar Cliente**
   * **Rota**: `GET /clientes`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Retorna os dados completos do cliente autenticado.

3. **Atualizar Cliente**
   * **Rota**: `PUT /clientes`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Atualiza campos específicos do cliente. Utiliza o método MergeAll do Firestore, o que significa que você só precisa enviar os campos que deseja alterar no Body.

4. **Deletar Cliente**
   * **Rota**: `DELETE /clientes`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Remove permanentemente o usuário do Firebase Auth e exclui seu documento na coleção clientes.

5. **Adicionar Favorito**
   * **Rota**: `PUT /clientes/favoritos`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Adiciona um prestador de serviço à lista de favoritos do cliente.
   * **Parâmetro (Query):** `?id=<id_do_prestador>`

6. **Listar Favoritos**
   * **Rota**: `GET /clientes/favoritos`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Retorna um array com os perfis públicos (detalhados) de todos os prestadores de serviço favoritados pelo cliente.

7. **Deletar Favorito**
   * **Rota**: `DELETE /clientes/favoritos`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Remove um prestador de serviço da lista de favoritos do cliente.
   * **Parâmetro (Query):** `?id=<id_do_prestador>`

### Prestadores de Serviço

1. **Criar Prestador**
   * **Rota**: `POST /prestadores`
   * **Autenticação:** Não requerida
   * **Descrição**: Cria as credenciais de acesso no Auth e inicializa o documento na coleção prestadores. Se não forem enviadas fotos de serviços, uma lista vazia é inicializada por padrão.

2. **Buscar Prestador**
   * **Rota**: `GET /prestadores`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Retorna o perfil completo do prestador autenticado.

3. **Atualizar Prestador**
   * **Rota**: `PUT /prestadores`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Atualiza parcialmente os dados do perfil do prestador.

4. **Deletar Prestador**
   * **Rota**: `DELETE /prestadores`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Deleta a conta de autenticação e os dados do prestador no banco.

5. **Listar Prestadores**
   * **Rota**: `GET /prestadores/busca`
   * **Autenticação:** Não requerida
   * **Descrição**: Retorna uma lista pública de prestadores ordenados da maior para a menor NotaMedia.
   * **Parâmetros (Query):**
      * `inicio` (int): Offset de paginação (padrão: 0).  
      * `fim` (int): Limite de itens na página (padrão: 15).  
      * `tipoServico` (string): Filtra prestadores por categorias específicas, separadas por vírgula (ex: `?tipoServico=Eletricista,Encanador`). Limite máximo de 10 categorias por busca.
      * `username` (string): Filtra prestadores pelo seu username. A filtragem é exata e CaseSensitive.
