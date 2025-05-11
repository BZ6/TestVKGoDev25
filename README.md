# TestVKGoDev25

## Задание

Задание состоит из 2х частей.

### Часть 1

Впервой части требуется реализовать пакет subpub. В этой части задания нужно написать простую шину событий, работающую по принципу **Publisher-Subscriber**.
Требование к шине:

- На один subject может подписываться (и отписываться) множество подписчкиков.
- Один медленный подписчик не должен тормозить остальных.
- Нельзя терять порядок сообщений (FIFO очередь).
- Метод Close должен учитывать переданный контекст. Если он отменен - выходим сразу, работающие хендлеры оставляем работать.
- Горутины (если они будут) течь не должны.

Ниже представлен API пакета subpub.

```golang
package subpub

import "context"

// MessageHandler is a callback function that proccesses messages delivered to subscribers.
type MessageHandler func(msg interface{})

// Unsubscribe will remove interest in the current subject subscription is for.
type Subscription interface {
    Unsubscribe()
}

type SubPub interface {
    // Subscribe creates an asynchronous queue subscriber on the given subject.
    Subscribe(subject string, cb MessageHandler) (Subscription, error)

    // Publish publishes the msg argument to the given subject.
    Publish(subject string, msg interface{}) error

    // Close will shutdown sub-pub system.
    // May be blocked by data delivery until the context is canceled.
    Close(ctx context.Context) error
}

func NewSubPub() SubPub {
    panic("Implement me")
}
```

К заданию рекомендуется писать unit-тесты.

### Часть 2

Во второй части задания требуется с использованием пакета subpub из 1 части реализовать сервис подписок. Сервис работает по gRPC. Есть возможность подписаться на события по ключу и опубликовать события по ключу для всех подписчиков.

Protobuf-схема gRPC сервиса:

```Golang
import "google/protobuf/empty.proto";

syntax = "proto3";

service PubSub {
    // Подписка (сервер отправляет поток событий)
    rpc Subscribe(SubscribeRequest) returns (stream Event);

    // Публикация (классический запрос-ответ)
    rpc Publish(PublishRequest) returns (google.protobuf.Empty);
}

message SubscribeRequest {
    string key = 1;
}

message PublishRequest {
    string key = 1;
    string data = 2;
}

message Event {
    string data = 1;
}
```

Также пользуйся стандартными статус-кодами gRPC из пакетов `google.golang.org/grpc/status` и `google.golang.org/grpc/codes` в качестве критериев успешности и неуспешности запросов к сервису. Что еще ожидается в решении:

- Обязательно должно быть описание того, как работает сервис и как его собирать.
- У сервиса должен быть свой конфиг, куда можно написать порты и прочие параметры (на ваше усмотрение).
- Логирование.
- Приветствуется использование известных паттернов при разработке микросервисов на Go (например, dependency injection, graceful shutdown и пр.). Если таковые будут использоваться, то просьбу упомянуть его в описании решения.

## Описание сервиса

Сервис реализует функциональность подписки и публикации событий по ключу. Он использует gRPC для взаимодействия между клиентами и сервером.

### Основные функции

1. **Подписка (Subscribe)**: Клиенты могут подписываться на события по ключу. Сервер отправляет поток событий подписчику.
2. **Публикация (Publish)**: Клиенты могут публиковать события по ключу. Все подписчики, подписанные на этот ключ, получают событие.

### Используемые паттерны

1. **Dependency Injection**:
    - Зависимость `subPub` передаётся в сервер через конструктор:

    ```golang
    pb.RegisterPubSubServer(grpcServer, &server{subPub: subPub})
    ```

2. **Graceful Shutdown**:
    - В методе `Close` пакета `subpub` используется контекст для корректного завершения работы подписчиков:

    ```golang
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-done:
        return nil
    }
    ```

### Сборка и запуск

1. **Сборка сервера**:

    Для генерации кода использовал `buf generate proto`.
    Выполните команду для сборки:

    ```bash
    go build -o server ./server
    ```

    или

    ```pwsh
    go build -o server.exe ./server
    ```

    Не забываем про: `go mod tidy`.

2. **Запуск сервера**:

    Выполните команду для запуска:

    ```bash
    ./server
    ```

    или

    ```pwsh
    ./server.exe
    ```

## Тестирование

В проекте реализован unit-тест для проверки работы пакета `subpub`. Тест проверяет следующие аспекты:

- Подписка на события (`Subscribe`).
- Публикация сообщений (`Publish`) и их доставка подписчикам.
- Корректная обработка сообщений подписчиком.
- Отписка от событий (`Unsubscribe`).
- Закрытие системы **Publisher-Subscriber** (`Close`).

### Запуск тестов

Для запуска тестов выполните следующую команду:

```bash
go test -v ./subpub
```

### Аналоги

Также для тестирования можно использовать grpcurl:

```bash
grpcurl -d '{\"key\": \"test\", \"data\": \"hello\"}' -plaintext localhost:50051 pubsub.PubSub/Publish
```

```bash
grpcurl -d '{\"key\": \"test\"}' -plaintext localhost:50051 pubsub.PubSub/Subscribe
```
