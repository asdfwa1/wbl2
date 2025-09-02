## Запуск сервиса
Заполнить .env файл переменными

```
LEVEL=local
PORT=8080
LOG_FILE_PATH=http_requests.log
LOG_TO_FILE=true
READ_TIMEOUT=10
WRITE_TIMEOUT=10
IDLE_TIMEOUT=60
  ```
- Выполнить go run main.go

### Протестировать до запуска go test ./...

### Посмотреть модели данных и проверить функционал через http://localhost:8080/swagger/index.html


