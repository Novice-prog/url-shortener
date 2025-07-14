# URL Shortener

## Features

* Сокращение HTTP/HTTPS‑ссылок
* Мгновенный 302‑редирект
* Подсчёт количества переходов
* Защита от SQL‑инъекций благодаря параметризованным запросам
* Абстрактный слой репозитория: отсутствие жёсткой привязки к SQLite, можно заменить на Postgres, MySQL и другие СУБД без изменения бизнес‑логики

## Prerequisites

* **Go** ≥ 1.22
* **SQLite** 3 (встроенная база поставляется автоматически)

## Setup and Running

```bash
# Клонируем репозиторий
git clone https://github.com/Novice-prog/url-shortener.git
cd url-shortener

# Запускаем локально
go run .
```

## Project Structure

```
.
├── cmd/            # точка входа
├── internal/
│   ├── handler/    # HTTP‑ручки (Gin)
│   ├── service/    # бизнес‑логика
│   └── repository/ # работа с БД
├── pkg/shortener/  # генератор коротких ID
└── go.mod
```
