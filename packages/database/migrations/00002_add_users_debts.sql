-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY,
  email         VARCHAR(255) NOT NULL UNIQUE,
  name          VARCHAR(255) NOT NULL,
  tax_regime    VARCHAR(50) NOT NULL DEFAULT '',
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS incomes (
  id            UUID PRIMARY KEY,
  user_id       UUID NOT NULL REFERENCES users(id),
  amount        NUMERIC(12,2) NOT NULL,
  type          VARCHAR(50) NOT NULL,
  date          DATE NOT NULL,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS debts (
  id            UUID PRIMARY KEY,
  user_id       UUID NOT NULL REFERENCES users(id),
  creditor      VARCHAR(255) NOT NULL,
  balance       NUMERIC(12,2) NOT NULL,
  installment   NUMERIC(12,2) NOT NULL,
  due_date      DATE NOT NULL,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS movements;
CREATE TABLE movements (
  id            UUID PRIMARY KEY,
  user_id       UUID NOT NULL REFERENCES users(id),
  type          VARCHAR(50) NOT NULL,
  amount        NUMERIC(12,2) NOT NULL,
  category      VARCHAR(100) NOT NULL,
  description   TEXT,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS limits;
CREATE TABLE limits (
  id            UUID PRIMARY KEY,
  user_id       UUID NOT NULL REFERENCES users(id),
  category      VARCHAR(100) NOT NULL,
  value         NUMERIC(12,2) NOT NULL,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS limits;
CREATE TABLE limits (
  id SERIAL PRIMARY KEY,
  category VARCHAR(100) NOT NULL,
  value NUMERIC(12,2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS movements;
CREATE TABLE movements (
  id SERIAL PRIMARY KEY,
  type VARCHAR(50) NOT NULL,
  amount NUMERIC(12,2) NOT NULL,
  category VARCHAR(100) NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS debts;
DROP TABLE IF EXISTS incomes;
DROP TABLE IF EXISTS users;
