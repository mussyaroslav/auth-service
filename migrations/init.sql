-- Схема для аутентификации и авторизации пользователей
CREATE SCHEMA IF NOT EXISTS auth
    AUTHORIZATION auth;

-- Таблица пользователей
CREATE TABLE auth.users
(
    user_id       UUID PRIMARY KEY,                 -- Уникальный идентификатор пользователя (UUID v4)
    username      VARCHAR(50) UNIQUE,               -- Уникальное имя пользователя для отображения (генерируется автоматически)
    email         VARCHAR(255) UNIQUE NOT NULL,     -- Электронная почта пользователя (уникальная, используется для входа)
    password_hash VARCHAR(255) NOT NULL,            -- Хеш пароля пользователя (с солью, алгоритм bcrypt/scrypt/Argon2)
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Дата и время создания записи
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Дата и время последнего обновления записи
);

-- Функция для генерации username из UUID
CREATE OR REPLACE FUNCTION generate_username()
    RETURNS TRIGGER AS
$$
BEGIN
    -- Генерирует username в формате "user_XXXXXX" из первых 8 символов UUID (без дефисов)
    NEW.username := 'user_' || SUBSTRING(REPLACE(NEW.user_id::TEXT, '-', ''), 1, 8);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматической генерации username при создании пользователя
CREATE TRIGGER set_username
    BEFORE INSERT
    ON auth.users
    FOR EACH ROW
EXECUTE PROCEDURE generate_username();

-- Таблица ролей пользователей
CREATE TABLE auth.roles
(
    role_id          SERIAL PRIMARY KEY,            -- Автоинкрементный идентификатор роли
    role_name        VARCHAR(50) UNIQUE NOT NULL,   -- Уникальное название роли (например, "admin", "user")
    role_description VARCHAR(128)                   -- Описание роли и её прав
);

-- Таблица связи пользователей с ролями (многие-ко-многим)
CREATE TABLE auth.user_roles
(
    user_role_id SERIAL PRIMARY KEY,                -- Автоинкрементный идентификатор связи
    user_id      UUID REFERENCES auth.users (user_id), -- Ссылка на пользователя (UUID)
    role_id      INT REFERENCES auth.roles (role_id), -- Ссылка на роль
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Дата и время назначения роли
);

-- Таблица запросов на сброс пароля
CREATE TABLE auth.password_reset_requests
(
    request_id      SERIAL PRIMARY KEY,             -- Автоинкрементный идентификатор запроса
    user_id         UUID REFERENCES auth.users (user_id), -- Ссылка на пользователя (UUID)
    token_value     UUID NOT NULL,                  -- Уникальный токен для сброса пароля (UUID v4)
    expiration_time TIMESTAMP NOT NULL,             -- Время истечения срока действия токена
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Дата и время создания запроса
);

-- Таблица токенов доступа/обновления
CREATE TABLE auth.tokens
(
    token_id        SERIAL PRIMARY KEY,             -- Автоинкрементный идентификатор токена
    user_id         UUID REFERENCES auth.users (user_id), -- Ссылка на пользователя (UUID)
    token_value     VARCHAR(255) NOT NULL,          -- Значение токена (JWT или случайная строка)
    expiration_time TIMESTAMP NOT NULL,             -- Время истечения срока действия токена
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Дата и время создания токена
);

-- Индексы для ускорения поиска
CREATE INDEX idx_users_email ON auth.users (email); -- Для быстрого поиска по email
CREATE INDEX idx_tokens_user_id ON auth.tokens (user_id); -- Для быстрого поиска токенов пользователя
CREATE INDEX idx_password_reset_token ON auth.password_reset_requests (token_value); -- Для быстрого поиска токена сброса пароля