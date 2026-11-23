# internal/config

В этом пакете хранятся конфигурации приложения.

Загрузку конфигурации можно реализовать из различных источников, например: файлов, переменных окружения, баз данных и других.

Все настройки можно задать в JSON-файле:

```json
{
    "server_address": "localhost:8080",
    "base_url": "http://localhost:8080",
    "file_storage_path": "/tmp/urls.json",
    "database_dsn": "",
    "auth_secret": "shortener-auth-secret",
    "audit_file": "",
    "audit_url": "",
    "enable_https": false
}
```

Путь задаётся флагом `-c` / `-config` или переменной окружения `CONFIG`:

```bash
go run ./cmd/shortener -c config.json
go run ./cmd/shortener -config=config.json -a localhost:9000
CONFIG=config.json go run ./cmd/shortener
```