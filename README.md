# Bico API

O Bico API é o backend REST responsável por gerenciar a plataforma de conexão entre clientes e prestadores de serviços. Construído em Go (Golang), ele utiliza o Firebase como solução de nuvem, integrando o Firestore para banco de dados NoSQL e o Firebase Authentication para o gerenciamento de credenciais dos usuários.

## Tecnologias Utilizadas
* Linguagem: Go (Golang)
* Banco de Dados: Google Cloud Firestore  
* Autenticação: Firebase Auth

## Autenticação
A maioria dos endpoints da aplicação são protegidos por um middleware de autenticação. Para acessá-los, você deve enviar o token JWT gerado pelo Firebase no lado do cliente no cabeçalho da requisição:

```http
Authorization: Bearer <SEU_TOKEN_JWT>
```

As operações de alteração de dados (`PUT`, `DELETE`) extraem a identidade do usuário diretamente desse token por motivos de segurança, eliminando a necessidade de enviar o ID na URL.

## Modelos de Dados (Domain)
O sistema baseia-se em uma estrutura comum `UserBase` que contém ID, Nome, Cpf, Email e Senha.  
* **Cliente**: Herda de UserBase e adiciona campos de endereço (Cep, Numero, Complemento).  
* **Prestador**: Herda de UserBase e adiciona o portfólio profissional, incluindo Username, TiposServico, LocalAtuacao, links para fotos (FotoPerfil, FotoPaginaPerfil, FotosServicos), um texto Sobre, NotaMedia e TotalAvaliacoes.

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
   * **Descrição**: Retorna os dados completos de um cliente específico. O ID do cliente é extraído automaticamente através do Token de Autenticação.

3. **Atualizar Cliente**
   * **Rota**: `PUT /clientes`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Atualiza campos específicos do cliente. O ID do cliente é extraído automaticamente através do Token de Autenticação. Utiliza o método MergeAll do Firestore, o que significa que você só precisa enviar os campos que deseja alterar no Body.

4. **Deletar Cliente**
   * **Rota**: `DELETE /clientes`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Remove permanentemente o usuário do Firebase Auth e exclui seu documento na coleção clientes. O ID é extraído automaticamente do Token.

### Prestadores de Serviço

1. **Criar Prestador**
   * **Rota**: `POST /prestadores`
   * **Autenticação:** Não requerida
   * **Descrição**: Cria as credenciais de acesso no Auth e inicializa o documento na coleção prestadores. Se não forem enviadas fotos de serviços, uma lista vazia é inicializada por padrão.

2. **Buscar Prestador**
   * **Rota**: `GET /prestadores`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Retorna o perfil completo de um prestador. O ID do cliente é extraído automaticamente através do Token de Autenticação.

3. **Atualizar Prestador**
   * **Rota**: `PUT /prestadores`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Atualiza parcialmente os dados do perfil do prestador. O ID do prestador é obtido pelo Token.

4. **Deletar Prestador**
   * **Rota**: `DELETE /prestadores`
   * **Autenticação:** Requerida (`Bearer Token`)
   * **Descrição**: Deleta a conta de autenticação e os dados do prestador no banco. O ID é obtido pelo Token.

5. **Listar Prestadores**
   * **Rota**: `GET /prestadores/lista`
   * **Autenticação:** Não requerida
   * **Descrição**: Retorna uma lista pública de prestadores ordenados da maior para a menor NotaMedia.
   * **inicio (int)**: Offset de paginação (padrão: 0).  
      * **fim (int)**: Limite de itens na página (padrão: 15).  
      * **tipoServico (string)**: Filtra prestadores por categorias específicas, separadas por vírgula (ex: `?tipoServico=Eletricista,Encanador`). Limite máximo de 10 categorias por busca.
