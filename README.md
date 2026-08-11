# Desafio weather API - Go

API para tempo real, integrando dados de **WeatherAPI** e **ViaCEP**.

## 🚀 Descrição

Projeto desenvolvido em Go que implementa uma API RESTful para buscar informações de dados climáticos em tempo real através do código de endereçamento postal (CEP). A aplicação integra **ViaCEP** para dados de endereços e **WeatherAPI** para informações climáticas, realizando requisições paralelas para otimizar o tempo de resposta.

## 📋 Pré-requisitos

- Go 1.21 ou superior
- Git
- Docker (opcional, para execução via container)

## 🏃 Como Executar o Servidor


1. Instale as dependências
   ```bash
   go mod tidy
   ```

2. Vá até o caminho /cmd/server 
 ```
 cd cmd/server
 ```

3. Configure as variáveis de ambiente (crie um arquivo `.env`) no caminho /cmd/server
   ```env
   VIA_CEP_API_HOST=https://viacep.com.br/ws
   API_WEATHER_HOST=https://api.weatherapi.com/v1/current.json
   API_WEATHER_KEY=sua_chave_de_api_aqui
   ```

   **Nota:** API_WEATHER_KEY: Obtenha sua chave gratuita fazendo login em https://www.weatherapi.com/

4. Execute o servidor
   ```bash
   go run main.go
   ```

O servidor estará disponível em `http://localhost:8080`

## 🐳 Como Executar com Docker

1. Construa a imagem Docker:
   ```bash
   docker build -t weather-api .
   ```

2. Execute o container:
   ```bash
   docker run -p 8080:8080 --env-file cmd/server/.env weather-api
   ```

O servidor estará disponível em `http://localhost:8080`

## 🧪 Como Rodar os Testes

Para executar todos os testes do projeto:

```bash
go test ./...
```

Para rodar testes de um pacote específico:

```bash
go test ./internal/infra/service/...
```

Para rodar testes com cobertura de código:

```bash
go test -cover ./...
```

## 📚 Como Abrir o Swagger

Com o servidor executando, acesse a documentação da API no seu navegador:

```
http://localhost:8080/docs/index.html
```

Lá você encontrará:
- ✅ Todos os endpoints disponíveis
- ✅ Modelos de requisição e resposta
- ✅ Exemplos de uso
- ✅ Possibilidade de testar os endpoints diretamente

## 🔌 Como Usar Extensão HTTP REST

### Usando a extensão REST Client

1. **Instale a extensão** no VS Code:
   - Procure por "REST Client" (publicada por Huachao Mao)
   - Ou execute: `ext install humao.rest-client`

2. **Use o arquivo** `test/cep.http` incluído no projeto:
   - Abra o arquivo `test/cep.http`
   - Clique em "Send Request" (ou use `Ctrl+Alt+R`)
   - Veja a resposta no painel de output

3. **Exemplo de requisição**:
   ```http
   GET http://localhost:8080/weather?cep=01001000 HTTP/1.1
   ```

## 🌐 Como Testar a API Implantada

A API está implantada no Google Cloud Run e pode ser testada diretamente:

- **Swagger**: https://deploy-cloud-run-mepu6h3qaa-uc.a.run.app/docs/index.html

- **Exemplo de requisição**:
  ```http
  GET https://deploy-cloud-run-mepu6h3qaa-uc.a.run.app/weather?cep=01001000 HTTP/1.1
  ```

## 📝 Endpoints Disponíveis

### Buscar Dados Climáticos
```
GET /weather?cep=01001000
```

Retorna informações dos dados climáticos (temperatura em graus Celsius, Fahrenheit e Kelvin) em tempo real da localidade via WeatherAPI.


## 📞 Contato

Desenvolvido por Luana Andrade - luanaands@gmail.com

---

**Aproveite! 🚀**
