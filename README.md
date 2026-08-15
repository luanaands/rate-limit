# Weather API com Rate Limit

API REST em Go para consultar dados climáticos a partir de um CEP. A aplicação
integra o ViaCEP e a WeatherAPI e usa o Redis para controlar o limite de
requisições.

## Funcionalidades

- Consulta de clima pelo endpoint `GET /weather?cep=01001000`.
- Integração com ViaCEP e WeatherAPI.
- Rate limit por `API_KEY` ou por endereço IP quando o token não é enviado.
- Janela de contagem de 1 segundo.
- Bloqueio temporário após exceder o limite, retornando HTTP `429` e o header
  `Retry-After`.
- Redis obrigatório para armazenar as contagens e os bloqueios.
- Documentação Swagger em `/docs/index.html`.

## Pré-requisitos

- Go 1.25 ou superior.
- Redis 7 ou superior.
- Docker e Docker Compose, caso queira executar toda a aplicação em containers.
- Uma chave da WeatherAPI. É possível obter uma chave gratuita em
  <https://www.weatherapi.com/>.

## Configuração

Copie o arquivo de exemplo:

```bash
cp cmd/server/.env.example cmd/server/.env
```

Edite `cmd/server/.env` e informe a chave da WeatherAPI. As principais
configurações são:

```env
VIA_CEP_API_HOST=https://viacep.com.br/ws
API_WEATHER_HOST=https://api.weatherapi.com/v1/current.json
API_WEATHER_KEY=sua_chave_de_api_aqui
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
RATE_LIMIT_TOKEN_LIMIT=100
RATE_LIMIT_IP_LIMIT=10
RATE_LIMIT_BLOCK_DURATION=1m
```

`REDIS_ADDR=redis:6379` é o endereço usado pelo Docker Compose. Para executar o
binário diretamente na máquina host, use o endereço do Redis local, normalmente
`REDIS_ADDR=localhost:6379`.

## Execução com Docker Compose

Essa é a forma recomendada, pois inicia a API e o Redis juntos:

```bash
docker compose up --build
```

O servidor ficará disponível em <http://localhost:8080>. Para encerrar os
containers:

```bash
docker compose down
```

## Execução local

Inicie um Redis local, ajuste `REDIS_ADDR` para `localhost:6379` no arquivo
`.env` e execute:

```bash
go mod download
cd cmd/server
go run .
```

O servidor ficará disponível em <http://localhost:8080>.

## Testes automatizados

Os testes do rate limit usam `miniredis`, portanto não precisam de um servidor
Redis externo:

```bash
# Todos os testes
go test ./...

# Testes do middleware de rate limit
go test -v ./internal/infra/webserver/middleware

# Testes com detector de condições de corrida
go test -race ./...

# Cobertura
go test -cover ./...
```

Os testes do middleware verificam a seleção entre token e IP, a contagem de
requisições, as chaves independentes e a expiração do bloqueio.

## Teste manual do Rate Limit

Para testar rapidamente, reduza temporariamente os limites no `.env`:

```env
RATE_LIMIT_TOKEN_LIMIT=2
RATE_LIMIT_IP_LIMIT=2
RATE_LIMIT_BLOCK_DURATION=30s
```

Reinicie a aplicação após alterar o arquivo. Em seguida, envie quatro
requisições simultâneas usando o mesmo token:

```bash
TOKEN="teste-local-$(date +%s)"

for i in $(seq 1 4); do
  curl -sS -o /dev/null \
    -w "requisicao $i: HTTP %{http_code}\n" \
    -H "API_KEY: $TOKEN" \
    "http://localhost:8080/weather?cep=01001000" &
done
wait
```

As duas primeiras requisições passam pelo rate limit. A partir da terceira, a
resposta esperada é `HTTP 429`, com `Retry-After: 30`. O status das duas
primeiras pode variar conforme a chave da WeatherAPI e os serviços externos,
mas elas não devem ser rejeitadas pelo rate limit.

Para testar o limite por IP, repita o comando sem o header `API_KEY`. Nesse
caso, as requisições serão contabilizadas pelo endereço IP do cliente.

## Endpoint e Swagger

Consulta de clima:

```http
GET http://localhost:8080/weather?cep=01001000
```

Com a aplicação em execução, a documentação está disponível em:

<http://localhost:8080/docs/index.html>

Também é possível usar o arquivo `test/cep.http` com a extensão REST Client do
VS Code.

## Estrutura principal

- `cmd/server`: inicialização do servidor e carregamento do `.env`.
- `configs`: leitura das configurações.
- `internal/infra/webserver/middleware`: rate limit baseado em Redis.
- `internal/infra/service`: integrações com ViaCEP e WeatherAPI.
- `internal/infra/webserver/handlers`: handlers HTTP.
- `docs`: documentação gerada do Swagger.

## Contato

Desenvolvido por Luana Andrade - luanaands@gmail.com
