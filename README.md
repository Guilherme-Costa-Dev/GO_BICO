# Bico API
O Bico API é o backend REST responsável por gerenciar a plataforma de conexão entre clientes e prestadores de serviços. Construído em Go (Golang), ele utiliza o Firebase como solução de nuvem, integrando 
o Firestore para banco de dados NoSQL e o Firebase Authentication para o gerenciamento de credenciais dos usuários.

## Tecnologias Utilizadas
* Linguagem: Go  
* Banco de Dados: Google Cloud Firestore  
* Autenticação: Firebase Auth

## Estrutura do Projeto
O projeto segue uma arquitetura modular dividida em pacotes internos:
* **internal/config**: Inicializa e gerencia a conexão com o Firebase (Firestore e Auth).  
* **internal/domain**: Define as estruturas de dados (Modelos) do sistema, como Cliente e Prestador.  
* **internal/handler**: Contém os controladores HTTP que processam as requisições e interagem com o banco de dados.

## Pré-requisitos
* GO
* Um projeto no Firebase configurado.  
* O arquivo de credenciais de serviço do Firebase. O arquivo JSON deve se encontrar na pasta raiz do projeto, com o nome de serviceAccountKey.json

## Modelos de Dados (Domain)
O sistema baseia-se em uma estrutura comum UserBase que contém ID, Nome, Cpf, Email e Senha.  
* **Cliente**: Herda de UserBase e adiciona campos de endereço (Cep, Numero, Complemento).  
* **Prestador**: Herda de UserBase e adiciona o portfólio profissional, incluindo Username, TiposServico, LocalAtuacao, links para fotos (FotoPerfil, FotoPaginaPerfil, FotosServicos), 
um texto Sobre, NotaMedia e TotalAvaliacoes.

## Endpoints
Todas as respostas da API são devolvidas no formato application/json. O UID gerado pelo Firebase Auth é utilizado como o ID principal do documento nas coleções do Firestore.

### Clientes

1. **Criar Cliente**
   * **Rota:** `POST /cliente`
   * **Descrição:** Cria as credenciais no Firebase Auth e salva os dados na coleção `clientes` do Firestore.
   * **Body (JSON):** Deve conter os campos definidos no modelo `Cliente` (incluindo a `senha` para o Auth).

2. **Buscar Cliente**
   * **Rota**: `GET /cliente?id={uid}`
   * **Descrição**: Retorna os dados completos de um cliente específico.

3. **Atualizar Cliente**
   * **Rota**: `PUT /cliente?id={uid}`
   * **Descrição**: Atualiza campos específicos do cliente. Utiliza o método MergeAll do Firestore, o que significa que você só precisa enviar os campos que deseja alterar.

4. **Deletar Cliente**
   * **Rota**: `DELETE /cliente?id={uid}`
   * **Descrição**: Remove permanentemente o usuário do Firebase Auth e exclui seu documento na coleção clientes.

### Prestadores de Serviço

1. **Criar Prestador**
   * **Rota**: `POST /prestador`
   * **Descrição**: Cria as credenciais de acesso no Auth e inicializa o documento na coleção prestadores. Se não forem enviadas fotos de serviços, uma lista vazia é inicializada por padrão.

2. **Buscar Prestador**
   * **Rota**: `GET /prestador?id={uid}`
   * **Descrição**: Retorna o perfil completo de um prestador pelo ID. 

3. **Atualizar Prestador**
   * **Rota**: `PUT /prestador?id={uid}`
   * **Descrição**: Atualiza parcialmente os dados do perfil do prestador.  

4. **Deletar Prestador**
   * **Rota**: `DELETE /prestador?id={uid}`
   * **Descrição**: Deleta a conta de autenticação e os dados do prestador no banco.  

5. **Listar Prestadores**
   * **Rota**: `GET /prestadores`
   * **Descrição**: Retorna uma lista de prestadores ordenados da maior para a menor NotaMedia.
   * **Query Params Suportados**:
      * **inicio (int)**: Offset de paginação (padrão: 0).  
      * **fim (int)**: Limite de itens na página (padrão: 15).  
      * **tipoServico (string)**: Filtra prestadores por categorias específicas, separadas por vírgula (ex: ?tipoServico=Eletricista,Encanador). Limite máximo de 10 categorias por busca.  
