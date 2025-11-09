# Data Service

Сервис для получения данных о цене Bitcoin с биржи.

## Описание

Сервис получает текущую цену Bitcoin из API биржи и предоставляет ее через HTTP endpoint. Инструментирован с помощью OpenTelemetry для метрик и трейсинга.

Source API docs - https://freecryptoapi.com/documentation/

Service API docs - see below

## Требования

- Go 1.21 или выше
- API ключ для `api.freecryptoapi.com`

## Установка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

3. Заполните `.env` файл своими значениями:
```bash
EXCHANGE_API_KEY=your_api_key_here
PORT=8080
```

## Запуск

```bash
export EXCHANGE_API_KEY=your_api_key_here
export PORT=8080
go run .
```

Или используйте `.env` файл с помощью инструмента вроде `godotenv` (если нужно).

## API Endpoints

### GET /health
Проверка здоровья сервиса.

**Примеры запросов:**
```bash
curl "http://localhost:8080/health"
```

**Response:**
```json
{
  "status": "ok"
}
```

### GET /price
Получение данных о цене криптовалюты с биржи.

**Query Parameters:**
- `symbol` (optional) - Символ криптовалюты (например, BTC, ETH). По умолчанию: BTC

**Примеры запросов:**
```bash
# Получение цены Bitcoin (по умолчанию)
curl "http://localhost:8080/price"

# Получение цены Ethereum
curl "http://localhost:8080/price?symbol=ETH"

# Получение цены Bitcoin
curl "http://localhost:8080/price?symbol=BTC"
```

**Response:**
```json
{
  "status": "success",
  "symbols": [
    {
      "symbol": "BTC",
      "last": "101711.63",
      "last_btc": "1",
      "lowest": "101487.25",
      "highest": "104072",
      "date": "2025-11-08 18:49:11",
      "daily_change_percentage": "-2.0047232754466",
      "source_exchange": "binance"
    }
  ]
}
```

## Переменные окружения

| Переменная | Описание | Обязательная | По умолчанию |
|------------|----------|--------------|--------------|
| `EXCHANGE_API_KEY` | API ключ для биржи | Да | - |
| `PORT` | Порт для HTTP сервера | Нет | 8080 |
| `EXCHANGE_API_URL` | URL API биржи | Нет | https://api.freecryptoapi.com/v1/getData |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Endpoint для OpenTelemetry | Нет | localhost:4317 |

## Тестирование

```bash
# Проверка здоровья
curl "http://localhost:8080/health"

# Получение цены Bitcoin (по умолчанию)
curl "http://localhost:8080/price"

# Получение цены Ethereum
curl "http://localhost:8080/price?symbol=ETH"
```

## OpenTelemetry

Сервис настроен для отправки метрик и трейсов в OpenTelemetry Collector. Если Collector не запущен, сервис продолжит работу, но трейсы не будут отправляться.

