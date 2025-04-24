-- Таблица для отслеживания сессий пользователей
CREATE TABLE auth.sessions
(
    session_id    SERIAL PRIMARY KEY,                  -- Уникальный идентификатор сессии
    user_id       INT REFERENCES auth.users (user_id), -- Внешний ключ, ссылающийся на таблицу users
    login_time    TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Время начала сессии
    last_activity TIMESTAMP                            -- Время последней активности в сессии
);

-- Таблица для хранения дополнительных учетных данных пользователей
CREATE TABLE auth.credentials
(
    credential_id SERIAL PRIMARY KEY,                  -- Уникальный идентификатор для каждой записи учетных данных
    user_id       INT REFERENCES auth.users (user_id), -- Внешний ключ, ссылающийся на таблицу users
    last_login    TIMESTAMP                            -- Время последнего входа пользователя
);