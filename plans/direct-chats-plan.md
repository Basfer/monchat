# 📋 План реализации: Личные чаты (Direct Messages)

## 🎯 Цель
Реализовать полноценную функциональность личных чатов 1-на-1 между пользователями.

---

## 📊 Текущее состояние
- ✅ Тип `RoomTypeDirect` уже определён в моделях
- ✅ Базовая структура комнат поддерживает разные типы
- ❌ Нет логики автоматического создания личных чатов
- ❌ Нет UI для поиска пользователей и начала личного чата
- ❌ Нет проверки на дубликаты (один личный чат между двумя пользователями)

---

## 🔧 Backend изменения

### 1. Модель данных (server/internal/models/)

#### 1.1 Обновить модель Room
**Файл:** `server/internal/models/room.go`
- ✅ Добавили поле `Participants []string` для явного хранения участников direct чата
- ✅ Добавили метод `IsDirect() bool` для удобства проверки типа
- ✅ Добавили метод `GetOtherParticipant(myUserID string) (string, error)` для получения ID второго участника

#### 1.2 Обновить модель Message
**Файл:** `server/internal/models/message.go`
- ✅ Добавили поле `EditedAt *time.Time` для поддержки будущего редактирования (в коде как UpdatedAt)
- ✅ Добавили поле `Deleted bool` для поддержки будущего удаления
- ✅ Изменили тип MessageType на расширенный enum:
  ```go
  const (
      MessageTypeText    MessageType = "m.text"
      MessageTypeEdited  MessageType = "m.emote" // маркер отредактированного
  )
  ```

### 2. Сервисный слой (server/internal/services/)

#### ✅ 2.1 Новый сервис: DirectChatService - создан
**Файл:** `server/internal/services/direct_chat_service.go` (НОВЫЙ)

```go
type DirectChatService struct {
    DB *sql.DB
}

func NewDirectChatService(db *sql.DB) *DirectChatService
func (s *DirectChatService) GetOrCreateDirectChat(userID1, userID2 string) (*models.Room, error)
func (s *DirectChatService) GetDirectChatsWithUser(userID, otherUserID string) ([]models.Room, error)
func (s *DirectChatService) FindExistingDirectChat(userID1, userID2 string) (*models.Room, error)
```

**Логика GetOrCreateDirectChat:**
1. Проверить существование существующего direct чата между двумя пользователями
2. Если существует — вернуть его
3. Если нет — создать новую комнату с type="direct"
4. Добавить обоих пользователей как members с ролью "member"
5. Вернуть созданную/найденную комнату

**Логика FindExistingDirectChat:**
```sql
SELECT r.* FROM rooms r
JOIN room_members rm1 ON r.id = rm1.room_id AND rm1.user_id = $1
JOIN room_members rm2 ON r.id = rm2.room_id AND rm2.user_id = $2
WHERE r.type = 'direct'
AND (
    (rm1.user_id = $1 AND rm2.user_id = $2) OR
    (rm1.user_id = $2 AND rm2.user_id = $1)
)
LIMIT 1
```

#### 2.2 Обновить RoomService
**Файл:** `server/internal/services/room_service.go`

✅ Добавили методы:
```go
func (s *RoomService) GetDirectChatsForUser(userID string) ([]models.Room, error)
```
- Фильтровать комнаты по типу 'direct'
- Для каждой комнаты получить имя/аватар второго участника
- Вернуть отсортированный список (по последнему сообщению)

✅ Обновили `GetUserRooms`:
- Добавить опциональный параметр `roomType` для фильтрации
- По умолчанию возвращать все комнаты

### ✅ 3. API Handlers (server/internal/api/) - сделано

#### ✅ 3.1 Новые endpoints

✅ **GET /api/v1/users** — Список всех пользователей (для поиска)
```go
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request)
```
- Query параметр: `?q=search_term`
- Возвращает пользователей, чьи username/email содержат search_term
- Исключать текущего пользователя из результатов
- Ограничение: max 20 результатов

✅ **POST /api/v1/direct-chat** — Получить или создать личный чат
```go
func (h *Handler) GetOrCreateDirectChat(w http.ResponseWriter, r *http.Request)
```
- Body: `{"user_id": "target_user_id"}`
- Использовать DirectChatService.GetOrCreateDirectChat
- Вернуть существующий или новый чат

✅ **GET /api/v1/direct-chats** — Все личные чаты пользователя
```go
func (h *Handler) GetUserDirectChats(w http.ResponseWriter, r *http.Request)
```
- Возвращает все direct комнаты пользователя
- Для каждой комнаты включить информацию о втором участнике

#### ✅ 3.2 Обновить существующие handlers

✅ **Update CreateRoom handler:**
- Добавить обработку case для type="direct"
- Автоматически добавить второго участника из поля `members[0]`
- Валидировать что ровно один участник указан для direct чата

✅ **Update GetUserRooms handler:**
- Добавить query параметр `?type=direct|group|channel`
- Передать фильтр в RoomService

### ✅ 4. Маршрутизация (server/internal/api/routes.go)

✅ Добавили новые маршруты:
```go
// Direct chat routes
r.HandleFunc("/api/v1/users", h.SearchUsers).Methods("GET")
r.HandleFunc("/api/v1/direct-chat", h.GetOrCreateDirectChat).Methods("POST")
r.HandleFunc("/api/v1/direct-chats", h.GetUserDirectChats).Methods("GET")
```

## ✅ Added database indexes for performance


## 🖥️ Frontend изменения

### 1. Типы (client-web/src/types/index.ts)

#### 1.1 Обновить интерфейс Room
```typescript
export interface Room {
  id: string;
  name: string | null;
  description: string | null;
  type: 'direct' | 'group' | 'channel' | 'public';
  created_by: string;
  is_private: boolean;
  created_at: string;
  updated_at: string;
  member_count: number;
  last_message: Message | null;
  
  // Новое поле для direct чатов
  participant?: {
    id: string;
    username: string;
    email: string;
  };
}
```

#### 1.2 Добавить новый интерфейс
```typescript
export interface UserSearchResult {
  id: string;
  username: string;
  email: string;
}

export interface DirectChatResponse {
  room: Room;
  other_participant: UserSearchResult;
}
```

### 2. Сервисы (client-web/src/services/)

#### 2.1 Обновить api.ts
```typescript
export const directChatService = {
  // Поиск пользователей для начала личного чата
  searchUsers: async (query: string, userId: string): Promise<UserSearchResult[]> => {
    const response = await fetch(`${API_BASE}/users?q=${encodeURIComponent(query)}&current_user=${userId}`, {
      headers: { 'X-User-ID': userId },
    });
    if (!response.ok) throw new Error('Failed to search users');
    return response.json();
  },

  // Получить или создать личный чат
  getOrCreateDirectChat: async (targetUserId: string, currentUserId: string): Promise<DirectChatResponse> => {
    const response = await fetch(`${API_BASE}/direct-chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': currentUserId,
      },
      body: JSON.stringify({ user_id: targetUserId }),
    });
    if (!response.ok) throw new Error('Failed to get/create direct chat');
    return response.json();
  },

  // Получить все личные чаты
  getDirectChats: async (userId: string): Promise<DirectChatResponse[]> => {
    const response = await fetch(`${API_BASE}/direct-chats`, {
      headers: { 'X-User-ID': userId },
    });
    if (!response.ok) throw new Error('Failed to get direct chats');
    return response.json();
  },
};
```

#### 2.2 Обновить roomService.ts
- Добавить метод `getDirectChats(userId: string)` для загрузки только direct комнат при инициализации

### 3. Store (client-web/src/store/)

#### 3.1 Новый store: directChatStore.ts
**Файл:** `client-web/src/store/directChatStore.ts` (НОВЫЙ)

```typescript
interface DirectChatState {
  searchQuery: string;
  searchResults: UserSearchResult[];
  isSearching: boolean;
  selectedChat: Room | null;
  setSearchQuery: (query: string) => void;
  setSearchResults: (results: UserSearchResult[]) => void;
  setIsSearching: (isSearching: boolean) => void;
  setSelectedChat: (chat: Room | null) => void;
  startSearch: (query: string) => Promise<void>;
  createDirectChat: (targetUserId: string) => Promise<Room>;
}
```

**Функциональность:**
- Debounced поиск пользователей (300ms задержка)
- Кэширование результатов поиска
- Управление состоянием выбранного чата

#### 3.2 Обновить roomStore.ts
- Добавить метод `loadDirectChats()` для загрузки личных чатов
- При создании direct чата — добавлять его в список комнат
- Сортировка: direct чаты отображаются выше групповых

### 4. Компоненты

#### 4.1 Новый компонент: UserSearchModal
**Файл:** `client-web/src/components/UserSearchModal.tsx` (НОВЫЙ)

**UI:**
- Модальное окно с полем поиска
- Список найденных пользователей с аватарами
- Кнопка "Начать чат" рядом с каждым пользователем
- Пустое состояние: "Введите имя пользователя"
- Состояние загрузки: спиннер
- Обработка ошибок

**Поведение:**
- Поиск запускается через debounce (300ms после ввода)
- При клике "Начать чат" — вызывается createDirectChat
- После создания чата — модальное закрывается, чат открывается
- Закрытие по клику вне модалки или Escape

#### 4.2 Обновить RoomList.tsx
**Файл:** `client-web/src/components/RoomList.tsx`

**Изменения:**
- Добавить кнопку "New Direct Message" рядом с "New Room"
- Открытие UserSearchModal при клике
- Визуальное разделение: direct чаты в отдельной секции
- Иконка 👤 для direct чатов (уже есть в getRoomIcon)
- Отображение username собеседника вместо имени комнаты для direct чатов

**Новая структура:**
```
┌─────────────────────┐
│ Rooms               │
│ [+ New DM] [+ Room] │
├─────────────────────┤
│ 👤 John Doe         │  <- Direct chats
│ 👤 Jane Smith       │
├─────────────────────┤
│ 👥 Project Team     │  <- Group chats
│ 📢 Announcements    │  <- Channels
└─────────────────────┘
```

#### 4.3 Обновить ChatWindow.tsx
**Файл:** `client-web/src/components/ChatWindow.tsx`

**Изменения:**
- Для direct чатов показывать username собеседника в заголовке
- Показывать статус "online/offline" (заготовка для будущего пункта 5.4)
- Аватар собеседника в заголовке

#### 4.4 Обновить ChatApp.tsx
**Файл:** `client-web/src/components/ChatApp.tsx`

**Изменения:**
- При инициализации загружать direct чаты отдельно
- Обрабатывать создание нового direct чата из UserSearchModal
- Обновлять список комнат после создания чата

---

## 🗄️ База данных

### SQL миграции

#### ✅ 1. Индекс для быстрого поиска direct чатов - сделанс
```sql
CREATE INDEX IF NOT EXISTS idx_rooms_type ON rooms(type);
CREATE INDEX IF NOT EXISTS idx_room_members_user_id ON room_members(user_id);
CREATE INDEX IF NOT EXISTS idx_room_members_room_id ON room_members(room_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

#### 2. Опционально: таблица для кэширования пар пользователей
```sql
-- Для оптимизации поиска существующих direct чатов
CREATE TABLE IF NOT EXISTS direct_chat_pairs (
    user_id_1 UUID NOT NULL,
    user_id_2 UUID NOT NULL,
    room_id UUID NOT NULL,
    PRIMARY KEY (user_id_1, user_id_2),
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

-- Триггер для автообновления при создании direct чата
CREATE OR REPLACE FUNCTION update_direct_chat_pair()
RETURNS TRIGGER AS $$BEGIN
    IF NEW.type = 'direct' THEN
        INSERT INTO direct_chat_pairs (user_id_1, user_id_2, room_id)
        VALUES (
            LEAST(
                (SELECT user_id FROM room_members WHERE room_id = NEW.id AND role = 'member' LIMIT 1),
                (SELECT user_id FROM room_members WHERE room_id = NEW.id AND role = 'member' OFFSET 1 LIMIT 1)
            ),
            GREATEST(...),
            NEW.id
        );
    END IF;
    RETURN NEW;
END;$$ LANGUAGE plpgsql;
```

> **Примечание:** Таблица direct_chat_pairs опциональна. На начальном этапе можно обойтись без неё, используя JOIN запросы.

---

## 🔄 WebSocket изменения

### Обновление hub.go
**Файл:** `server/internal/websocket/hub.go`

**Изменения:**
- При подключении по WS проверять, является ли комната direct чатом
- Отправлять только сообщения от участников этой комнаты
- Для direct чатов — фильтровать сообщения строже (только от другого участника)

### Новые WS события
```json
{
  "type": "direct_chat_created",
  "room_id": "uuid",
  "other_participant": {
    "id": "user_uuid",
    "username": "username"
  }
}
```
- Отправлять обоим участникам при создании нового direct чата
- Позволяет обновить UI в реальном времени

## 🧪 Тестирование
### Backend тесты
**Файл:** server/internal/services/direct_chat_service_test.go (НОВЫЙ)

Тестовые случаи:
- Создание нового direct чата между двумя пользователями
- Возврат существующего direct чата при повторном запросе
- Проверка что порядок userID не важен (A+B == B+A)
- Валидация что нельзя создать direct чат с самим собой
- Получение списка всех direct чатов пользователя

### Frontend тесты
**Файл:** client-web/src/components/__tests__/UserSearchModal.test.tsx (НОВЫЙ)

Тестовые случаи:
- Рендеринга компонента
- Поиск пользователей с debounce
- Отображение результатов поиска
- Создание чата при клике
- Обработка ошибок

## 📝 Порядок реализации

### Этап 1: Backend Core (2-3 часа)
- [x] Создать DirectChatService с методами GetOrCreateDirectChat и FindExistingDirectChat
- [x] Обновить модели Room и Message
- [x] Добавить новые API endpoints (SearchUsers, GetOrCreateDirectChat, GetUserDirectChats)
- [x] Обновить маршрутизацию
- [x] Написать backend unit тесты

### Этап 2: Frontend Services & Store (1-2 часа)
- [x] Добавить новые интерфейсы TypeScript
- [x] Создать directChatStore
- [x] Добавить методы в api.ts
- [x] Обновить roomStore для работы с direct чатами

### Этап 3: Frontend UI (2-3 часа)
- [x] Создать компонент UserSearchModal
- [x] Обновить RoomList с разделением на Direct/Groups
- [x] Обновить ChatWindow для отображения имени собеседника
- [x] Обновить ChatApp для инициализации direct чатов

### Этап 4: Интеграция и полировка (1-2 часа)
- [x] Добавить WebSocket событие direct_chat_created
- [x] Добавить индексы в БД
- [x] Протестировать полный flow: поиск → создание чата → отправка сообщения
- [x] Обработка edge cases (создание чата с самим собой, уже существующий чат)

## ⚠️ Потенциальные проблемы и решения

### Проблема 1: Дубликаты чатов
**Решение:** Строгая проверка через FindExistingDirectChat перед созданием нового чата. Использование уникального ограничения в БД на пару пользователей.

### Проблема 2: Производительность поиска пользователей

**Решение:**

- выдавать результаты пачками до 50, догружать при необходимости; 
- Index по username и email
- Debounce на клиенте

### Проблема 3: Синхронизация списка комнат

**Решение:**
- При создании direct чата — добавлять его в локальный стейт
- При перезагрузке страницы — загружать direct чаты отдельно
- WebSocket event для real-time обновления

### Проблема 4: Имя комнаты для direct чата

**Решение:**
- Для direct чатов имя комнаты игнорируется
- UI показывает username собеседника динамически
- Описание комнаты не требуется

## 📊 Критерии готовности (Definition of Done)

- [x] Backend API endpoints работают корректно
- [x] Unit тесты написаны и проходят
- [x] Frontend компоненты реализованы
- [x] WebSocket событие direct_chat_created отправляется
- [x] Индексы в БД добавлены
- [x] Полный flow протестирован: поиск → создание чата → отправка сообщения
- [x] Edge cases обработаны (создание чата с самим собой, уже существующий чат)
- [x] Build проходит без ошибок
- [x] Код ревью выполнен
