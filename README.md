# Rate Limiter em Go

Este projeto implementa um rate limiter robusto e flexível em Go que pode limitar requisições baseado em endereço IP ou token de acesso. O rate limiter foi desenvolvido seguindo boas práticas de arquitetura e permite fácil extensibilidade através do padrão Strategy.

## 🚀 Características

- **Limitação por IP**: Controla requisições por endereço IP
- **Limitação por Token**: Controla requisições por token de acesso (API_KEY)
- **Prioridade de Token**: Configurações de token sobrepõem as de IP
- **Middleware HTTP**: Integração fácil como middleware
- **Strategy Pattern**: Fácil troca de mecanismo de persistência
- **Redis Storage**: Persistência em Redis com fallback para outros storages
- **Bloqueio Temporário**: Tempo configurável de bloqueio quando limite é excedido
- **Testes Automatizados**: Cobertura completa de testes unitários e integração
- **Docker Ready**: Configuração completa com Docker e Docker Compose

## 🏗️ Arquitetura

```
├── cmd/                    # Ponto de entrada da aplicação
│   └── main.go
├── internal/               # Código interno da aplicação  
│   ├── config/            # Configurações
│   │   └── config.go
│   └── middleware/        # Middlewares HTTP
│       └── rate_limiter.go
├── pkg/                   # Bibliotecas reutilizáveis
│   └── ratelimiter/       # Core do rate limiter
│       ├── storage.go     # Interface do Storage
│       ├── redis_storage.go # Implementação Redis
│       └── rate_limiter.go # Lógica principal
├── test/                  # Testes de integração
│   └── integration_test.go
├── docker-compose.yml     # Configuração Docker
├── Dockerfile
└── README.md
```

## ⚙️ Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto ou configure as seguintes variáveis:

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiter Configuration
DEFAULT_IP_LIMIT=10                 # Requisições por segundo por IP
DEFAULT_TOKEN_LIMIT=100             # Requisições por segundo por token
BLOCK_DURATION_SECONDS=300          # Tempo de bloqueio em segundos (5 min)

# Server Configuration
SERVER_PORT=8080

# Token-specific limits (opcional)
TOKEN_abc123_LIMIT=50              # Token específico com limite de 50 req/s
TOKEN_premium_user_LIMIT=200       # Token premium com limite de 200 req/s
```

### Exemplo de arquivo `.env`

```bash
cp env.example .env
```

## 🐳 Executando com Docker

### 1. Usando Docker Compose (Recomendado)

```bash
# Subir Redis e aplicação
docker-compose up -d

# Ver logs
docker-compose logs -f app

# Parar serviços
docker-compose down
```

### 2. Executando apenas o Redis

```bash
# Subir apenas o Redis
docker-compose up -d redis

# Executar aplicação localmente
go run cmd/main.go
```

## 🔧 Executando Localmente

### Pré-requisitos

- Go 1.21+
- Redis (pode usar Docker)

### Passos

1. **Clone o repositório**
```bash
git clone <repository-url>
cd rate-limiter
```

2. **Instale dependências**
```bash
go mod download
```

3. **Configure o ambiente**
```bash
cp env.example .env
# Edite o .env conforme necessário
```

4. **Inicie o Redis**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

5. **Execute a aplicação**
```bash
go run cmd/main.go
```

## 📊 Como Funciona

### Lógica do Rate Limiter

1. **Prioridade**: Token tem prioridade sobre IP
2. **Contagem**: Utiliza contador com janela deslizante de 1 segundo
3. **Bloqueio**: Quando limite é excedido, bloqueia por tempo configurado
4. **Recuperação**: Após expirar o bloqueio, permite novas requisições

### Fluxo de Decisão

```
Requisição → Extrair IP/Token → Verificar se está bloqueado →
Incrementar contador → Verificar limite → 
Se excedeu: bloquear e rejeitar → 
Se não: permitir e continuar
```

## 🔌 Como Usar

### 1. Requisições sem Token (Limitação por IP)

```bash
# Primeira requisição (permitida)
curl http://localhost:8080/

# ... continue fazendo requisições até atingir o limite
curl http://localhost:8080/
```

### 2. Requisições com Token (Limitação por Token)

```bash
# Com token específico (limite configurado)
curl -H "API_KEY: abc123" http://localhost:8080/

# Com token genérico (limite padrão de token)
curl -H "API_KEY: my_token" http://localhost:8080/
```

### 3. Testando Limites

```bash
# Script para testar rapidamente
for i in {1..15}; do
  echo "Request $i:"
  curl -s http://localhost:8080/ | jq .
  sleep 0.1
done
```

## 🧪 Testes

### Executar Testes Unitários

```bash
go test ./pkg/ratelimiter -v
```

### Executar Testes de Integração

```bash
go test ./test -v
```

### Executar Todos os Testes

```bash
go test ./... -v
```

### Teste de Carga

```bash
# Instalar hey (ferramenta de benchmark)
go install github.com/rakyll/hey@latest

# Teste com 100 requisições, 10 concorrentes
hey -n 100 -c 10 http://localhost:8080/

# Teste com token
hey -n 100 -c 10 -H "API_KEY: abc123" http://localhost:8080/
```

## 📡 Endpoints de Exemplo

- `GET /` - Endpoint principal
- `GET /health` - Health check
- `POST /api/data` - Endpoint protegido

### Resposta quando Limite Excedido

```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

Status Code: `429 Too Many Requests`

## 🔄 Extensibilidade

### Adicionando Novo Storage

Para trocar o Redis por outro mecanismo de persistência:

1. **Implemente a interface Storage**

```go
type CustomStorage struct {
    // sua implementação
}

func (c *CustomStorage) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
    // implementação
}

func (c *CustomStorage) Get(ctx context.Context, key string) (int64, error) {
    // implementação  
}

// ... outros métodos
```

2. **Use na aplicação**

```go
// Substitua na main.go
storage := NewCustomStorage() // em vez de ratelimiter.NewRedisStorage()
```

### Configurações Avançadas

- **Diferentes janelas de tempo**: Modifique o `time.Second` no `Increment`
- **Algoritmos alternativos**: Implemente token bucket ou sliding window log
- **Métricas**: Adicione instrumentação com Prometheus
- **Logs estruturados**: Integre com logrus ou zap

## 🚨 Tratamento de Erros

O rate limiter lida com:

- **Falha do Redis**: Retorna erro 500
- **IP inválido**: Usa o RemoteAddr como fallback
- **Token malformado**: Trata como ausência de token
- **Concorrência**: Usa transações Redis para consistência

## 📈 Monitoramento

### Headers de Resposta

- `X-RateLimit-Remaining`: Requisições restantes
- `X-RateLimit-Reset`: Timestamp do reset

### Logs

A aplicação registra:
- Configurações no startup
- Erros de conexão Redis
- Bloqueios por limite excedido

## 🔧 Configurações de Produção

### Redis

```bash
# Para produção, configure Redis com:
# - Autenticação
# - SSL/TLS
# - Cluster se necessário
# - Backup/Replicação
```

### Aplicação

```bash
# Configure adequadamente:
# - Timeouts de Redis
# - Pool de conexões
# - Graceful shutdown
# - Health checks
```

## 📈 Exemplos Práticos

### Cenário 1: API Pública

```bash
# Configuração para API pública
DEFAULT_IP_LIMIT=100
DEFAULT_TOKEN_LIMIT=1000
BLOCK_DURATION_SECONDS=3600  # 1 hora
```

### Cenário 2: API Interna

```bash
# Configuração mais permissiva
DEFAULT_IP_LIMIT=1000
DEFAULT_TOKEN_LIMIT=10000
BLOCK_DURATION_SECONDS=300   # 5 minutos
```

### Cenário 3: Diferentes Tiers

```bash
# Tokens com diferentes limites
TOKEN_free_tier_LIMIT=10
TOKEN_basic_plan_LIMIT=100
TOKEN_premium_plan_LIMIT=1000
TOKEN_enterprise_LIMIT=10000
``` 

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para detalhes.

## 👨‍💻 Autor

Desenvolvido por [Danilo Torchio](https://github.com/danilotorchio) como parte do projeto Go Expert.

---

**Nota**: Esta aplicação foi desenvolvida para fins educacionais e de demonstração. Para uso em produção, considere implementar features adicionais como logging estruturado, métricas mais detalhadas e configurações avançadas de rede. 