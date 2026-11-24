# Library Management API (CSE5032)

REST API веб-сервис на языке Go (`net/http`) для автоматизации библиотеки с хранением данных в памяти.

## Запуск и проверка

```bash
go run ./cmd/server
go build ./...
go vet ./...
go test -v ./...
```

## Где используются конструкции Go

- **Массив (`[12]int`)**: `LoansByMonth` в `internal/models/statistics.go` и `internal/services/loan_service.go` — подсчёт выдач по месяцам.
- **Срезы (`[]Book`, `[]Reader`, `[]Loan`)**: списки сущностей и результаты фильтрации (`SearchBooks`, `GetAllBooks`).
- **Карты (`map[int]Book`, `map[int]Reader`, `map[int]Loan`)**: хранилище в `internal/storage/memory.go`, защищённое `sync.RWMutex`.
- **Функции (более 10)**: `ValidateBookInput`, `ValidateReaderInput`, `decrementAvailableCopies`, `incrementAvailableCopies`, `markLoanReturned`, `calculateAvailabilityPercent` и методы сервисов.
- **Указатели (`*Book`, `*Reader`, `*Loan`, `*time.Time`)**: изменение объектов по указателю при выдаче и возврате книг, поле `ReturnedAt *time.Time`.
- **Маршрутизация (`http.ServeMux`)**: паттерны Go 1.22+ (`GET /api/books/{id}`, `PUT /api/loans/{id}/return`) с обработкой `404` и `405` в `internal/router/router.go`.
- **Middleware**: `middleware.Logging` логирует метод, URL, статус и время выполнения каждого запроса.