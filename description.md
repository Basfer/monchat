## Project Identity: 
    monchat (Matrix-style messenger), Go backend, React/TypeScript frontend; docker.
## Architecture: 
    Monolithic structure with internal packages (api, database, models, services, websocket).
## Critical Security Issues: 
    Passwords hashed with SHA256 (needs bcrypt); no JWT authentication; CORS CheckOrigin allows all origins; missing input sanitization.
## Database Issues:
    AddMember uses ON CONFLICT DO NOTHING incorrectly; missing indexes on room_members; type mismatch between DB VARCHAR and code enum; missing updated_at trigger.
## WebSocket Issues: 
    DefaultHub singleton causes testing issues; no WS auth; reconnect logic creates multiple connections; no heartbeat; no offline buffering.
## Implemented Features: 
    Direct chats (1-on-1) implemented with dedicated service, store, and WebSocket events.
## Missing Features: 
    Message search, edit/delete messages.
## Ports: 
    Server 8008, Client 3000.
## Structure:
    monchat/ - корень проекта (содержит конфигурацию, документацию и поддиректории client-web/server)
    ├── .gitignore - исключения
    ├── .kilo/kilo.jsonc - конфиг Kilo
    ├── .vscode/settings.json - настройки VSCode
    ├── docker-compose.yml - конфигурация Docker Compose
    ├── README.md - описание проекта
    ├── TODO.md - глобальный todo-list
    ├── description.md - информация для работы над проектом
    ├── plans/ - планы и документация
    │   └── direct-chats-plan.md - план реализации прямых чатов
    ├── client-web/ - клиентская часть (React)
    │   ├── Dockerfile - образ для клиента
    │   ├── index.html - главная страница
    │   ├── jest.config.js - конфигурация Jest для тестирования
    │   ├── nginx.conf - конфиг Nginx для клиента
    │   ├── package.json - зависимости клиента
    │   ├── TESTING.md - документация по тестированию
    │   ├── tsconfig.json - конфиг TypeScript для клиента
    │   ├── tsconfig.node.json - конфиг TypeScript для Node
    │   ├── vite.config.ts - конфиг Vite для клиента
    │   └── src/ - исходный код клиента (React компоненты, сервисы, стейт-менеджмент)
    │       ├── App.tsx - главный React компонент приложения
    │       ├── main.tsx - точка входа React приложения
    │       ├── index.css - глобальные стили CSS
    │       ├── vite-env.d.ts - объявления типов Vite
    │       ├── components/ - React компоненты интерфейса
    │       │   ├── ChatApp.tsx - приложение чата
    │       │   ├── ChatWindow.tsx - окно чата
    │       │   ├── Login.tsx - страница входа
    │       │   ├── RoomList.tsx - список комнат
    │       │   └── UserSearchModal.tsx - модальное окно поиска пользователей
    │       ├── services/ - сервисы (API, WebSocket)
    │       │   ├── api.ts - API клиент
    │       │   └── websocket.ts - WebSocket клиент
    │       ├── store/ - управление состоянием (Zustand)
    │       │   ├── authStore.ts - стейт аутентификации
    │       │   ├── directChatStore.ts - стейт прямых чатов
    │       │   └── roomStore.ts - стейт комнат
    │       └── types/ - TypeScript типы
    │           └── index.ts - экспорты типов
    └── server/ - серверная часть (Go)
        ├── Dockerfile - образ для сервера
        ├── go.mod - зависимости Go сервера
        ├── go.sum - контрольные суммы зависимостей
        └── cmd/ - директория с командами CLI
            └── main.go - точка входа сервера
        └── internal/ - внутренняя логика сервера
            ├── api/ - HTTP API слой
            │   ├── handlers.go - обработчики HTTP запросов
            │   └── routes.go - маршрутизация API
            ├── database/ - работа с базой данных
            │   └── postgres.go - работа с PostgreSQL
            ├── models/ - модели данных
            │   ├── message.go - модель сообщения
            │   ├── room.go - модель комнаты
            │   └── user.go - модель пользователя
            ├── services/ - бизнес-логика
            │   ├── auth_service.go - сервис аутентификации
            │   ├── direct_chat_service.go - сервис прямых чатов
            │   ├── direct_chat_service_test.go - тесты сервиса прямых чатов
            │   └── room_service.go - сервис комнат
            └── websocket/ - WebSocket функциональность
                ├── direct_chat_event.go - события прямых чатов WebSocket
                └── hub.go - WebSocket хаб
