# MonChat - Matrix Messenger

Мессенджер, реализующий основные функции протокола Matrix с поддержкой разных типов комнат (личные чаты, группы, каналы).

## Архитектура

- **Backend**: Go (Golang) с WebSocket для real-time коммуникации
- **Frontend**: React + TypeScript с Vite
- **База данных**: PostgreSQL
- **Кэш/Сессии**: Redis

## Типы комнат

- **Личные чаты (Direct)**: Переписка 1-на-1
- **Групповые чаты (Group)**: Многостороннее общение с ролями участников
- **Каналы (Channel)**: Односторонняя публикация сообщений (только администраторы могут писать)
- **Публичные комнаты**: Открытые для поиска и вступления

## Быстрый старт

### Требования

- Docker и Docker Compose
- Node.js 18+ (для локальной разработки клиента)
- Go 1.21+ (для локальной разработки сервера)

### Запуск через Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f server
docker-compose logs -f client
```

Сервер будет доступен на `http://localhost:8008`
Клиент будет доступен на `http://localhost:3000`

### Локальная разработка

#### Backend

```bash
cd server
go mod tidy
go run cmd/main.go
```

#### Frontend

```bash
cd client-web
npm install
npm run dev
```

## API Endpoints

### Аутентификация

- `POST /api/v1/register` - Регистрация пользователя
- `POST /api/v1/login` - Вход пользователя

### Комнаты

- `POST /api/v1/rooms` - Создать комнату
- `GET /api/v1/rooms` - Получить список комнат пользователя

### Сообщения

- `GET /api/v1/rooms/{roomId}/messages` - Получить сообщения комнаты
- `POST /api/v1/rooms/{roomId}/messages` - Отправить сообщение

### WebSocket

- `ws://localhost:8008/ws?userId={userId}&roomId={roomId}` - WebSocket соединение

## Структура проекта

```
matrix-messenger/
├── server/                 # Серверная часть на Go
│   ├── cmd/
│   │   └── main.go        # Точка входа
│   ├── internal/
│   │   ├── api/           # HTTP handlers
│   │   ├── database/      # Работа с БД
│   │   ├── models/        # Модели данных
│   │   ├── services/      # Бизнес-логика
│   │   └── websocket/     # WebSocket обработка
│   ├── go.mod
│   └── Dockerfile
├── client-web/            # Веб-клиент на React
│   ├── src/
│   │   ├── components/    # React компоненты
│   │   ├── services/      # API и WebSocket сервисы
│   │   ├── store/         # Zustand store
│   │   └── types/         # TypeScript типы
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## Дальнейшее развитие

- [ ] Поддержка медиафайлов (изображения, документы)
- [ ] End-to-end шифрование
- [ ] Голосовые и видеозвонки (WebRTC)
- [ ] Поиск по сообщениям
- [ ] Эмодзи и реакции
- [ ] Threads/ответы на сообщения
- [ ] Уведомления
- [ ] Мобильное приложение

## Лицензия

MIT