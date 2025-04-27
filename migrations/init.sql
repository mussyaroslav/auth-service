-- Создание схемы auth
CREATE SCHEMA IF NOT EXISTS auth
    authorization auth;

-- Таблица пользователей, хранящая основную информацию о пользователях
CREATE TABLE auth.users
(
    user_id       SERIAL PRIMARY KEY,                 -- Уникальный идентификатор пользователя
    username      VARCHAR(50) UNIQUE,                 -- Уникальное имя пользователя
    email         VARCHAR(255) UNIQUE NOT NULL,       -- Электронная почта пользователя, уникальная и обязательная
    password_hash VARCHAR(255)        NOT NULL,       -- Хешированный пароль, обязательный
    first_name    VARCHAR(100),                       -- Имя пользователя, обязательное
    last_name     VARCHAR(100),                       -- Фамилия пользователя, обязательная
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Время создания учетной записи
);

-- Функция триггера для автоматической генерации username
CREATE OR REPLACE FUNCTION generate_username()
    RETURNS TRIGGER AS
$$
BEGIN
    -- Формирование username в формате 'id' + user_id, дополненное нулями до 10 символов
    NEW.username := 'id' || LPAD(NEW.user_id::text, 10, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер, вызывающий функцию generate_username перед вставкой новой записи в users
CREATE TRIGGER set_username
    BEFORE INSERT
    ON auth.users
    FOR EACH ROW
EXECUTE PROCEDURE generate_username();

-- Таблица ролей, определяющая различные роли пользователей
CREATE TABLE auth.roles
(
    role_id          SERIAL PRIMARY KEY,          -- Уникальный идентификатор роли
    role_name        VARCHAR(50) UNIQUE NOT NULL, -- Имя роли, уникальное и обязательное
    role_description VARCHAR(128)                 -- Краткое описание роли
);

-- Таблица, связывающая пользователей с их ролями
CREATE TABLE auth.user_roles
(
    user_role_id SERIAL PRIMARY KEY,                  -- Уникальный идентификатор для каждой записи
    user_id      INT REFERENCES auth.users (user_id), -- Внешний ключ, ссылающийся на таблицу users
    role_id      INT REFERENCES auth.roles (role_id)  -- Внешний ключ, ссылающийся на таблицу roles
);

-- Таблица для хранения информации о запросах на сброс пароля
CREATE TABLE auth.password_reset_requests
(
    request_id      SERIAL PRIMARY KEY,                  -- Уникальный идентификатор запроса на сброс пароля
    user_id         INT REFERENCES auth.users (user_id), -- Внешний ключ, ссылающийся на таблицу users
    token_value     VARCHAR(255) NOT NULL,               -- Токен для сброса пароля, обязательный
    expiration_time TIMESTAMP    NOT NULL                -- Время истечения токена, обязательное
);

-- Таблица для хранения токенов аутентификации
CREATE TABLE auth.tokens
(
    token_id        SERIAL PRIMARY KEY,                  -- Уникальный идентификатор токена
    user_id         INT REFERENCES auth.users (user_id), -- Внешний ключ, ссылающийся на таблицу users
    token_value     VARCHAR(255) NOT NULL,               -- Значение токена для аутентификации, обязательное
    expiration_time TIMESTAMP    NOT NULL                -- Время истечения токена, обязательное
);