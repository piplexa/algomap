# REST API сервер

## Документация

Документируем API с помощью Swagger.

1.

```bash
go install github.com/swaggo/swag/cmd/swag@latest
cd backend
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/files
```

После установки swag, не забыть прописать его в path или сделать линк.

2.

Добавляем комментарии к основным параметрам API в `backend/cmd/api/main.go` (пример в файле).
Добавляем комментарии к методам API в хендлерах (пример в `internal/handlers/auth.go`).

3.

Добавить ответы и запросы в `internal/handlers/responses.go` (пример в файле).

4.

Сгенерировать документацию:

```bash
swag init -g ./cmd/api/main.go -o ./docs
```
Может понадобится еще:
```
go mod tidy
```

5.

http://localhost:8080/swagger/index.html
