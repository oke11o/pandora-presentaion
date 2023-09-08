# Scenrio mock

## Описание сервиса

Сервис для тестирования сценариев.

Предполагается простой сценарий

- Пользователь отправляет запрос на авторизацию. В ответ получает токен авторизации.
- С токеном авторизации пользователь отправляет запрос на получение списка данных.
- С токеном авторизации и item_id пользователь отправляет запрос на получение данных по идентификатору.

Есть 2 вспомогательных метода.

- statistic - получаение статистики сервиса по 3м выше описынным методам. (200, 400, 500)
- reset - сброс статистики

#### Сценарий теста

Preconditions: сбрасываем статистику

- Получаем user_id из списка (1,10)
- По нему получаем список item_ids
- Берем рандомный item_id получаем ответ.
- Проверяем, что в ответе есть поле "data.status:'success'"
- FINISH

Postconditions: проверяем статистику

## Сущности

- **user_id** - `range(1,10)`
- **item_id** - `user_id*1000 + range(0,99)`

Пример, для user_id=5, item_id=range(5000, 5099)

## Config

Используется лишь переменная окружения PORT.
При отсутствии этой переменной - дефолтный порт 8091.

## Handlers

### POST {{service}}/auth

#### Request

```
Content-Type: application/json
{"user_id": 5}
```

#### Response

```json
{
  "auth_key": "eWoiCJznmBkTfgPraRPnVlebWBtEQBQJkytkTKJKzZwkZYQlSGoxPlAbkpKmctOs"
}
```

### GET {{service}}/list

#### Request

```
Content-Type: application/json
Authorization: Bearer eWoiCJznmBkTfgPraRPnVlebWBtEQBQJkytkTKJKzZwkZYQlSGoxPlAbkpKmctOs
```

#### Response

```json
{
    "items": [
        5000,
        ...
        5099
    ]
}
```

### POST {{service}}/item

#### Request

```
Content-Type: application/json
Authorization: Bearer eWoiCJznmBkTfgPraRPnVlebWBtEQBQJkytkTKJKzZwkZYQlSGoxPlAbkpKmctOs
{"item_id": 5099}
```

#### Response

```json
{
    "item": 5099
}
```

### GET {{service}}/statistic

#### Request

```
Content-Type: application/json
```

#### Response

```json
{
    "auth": {
        "200": {
            "5": 1,
            "9": 2
        },
        "400": 0,
        "500": 0
    },
    "list": {
        "200": {
            "5": 1,
            "9": 2
        },
        "400": 0,
        "500": 0
    },
    "item": {
        "200": {
            "5": 1,
            "9": 2
        },
        "400": 0,
        "500": 0
    }
}
```

### POST {{service}}/reset

#### Response

```json
{
  "status": "ok"
}
```






