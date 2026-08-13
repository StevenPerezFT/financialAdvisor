# Project Log

This file tracks decisions and progress. Updated at each work session.

## Product goal

- **Phase 1 (current) — B2C**: help individuals with multiple income sources and multiple debts understand their financial situation and tax obligations.
- **Phase 2 (future) — B2B**: package the above as a tool for financial advisory companies (multi-client, payroll/employees). Not implemented yet, but Phase 1 decisions must leave room for it.

## Architecture decisions

- **No multi-tenancy yet**, but a `User` entity is introduced now, and everything (`incomes`, `debts`, `movimientos`, `limites`) is related through `user_id`. This way, adding a `company_id` later is an additive migration, not a rewrite of the model.
- **Incomes and debts are collections of their own entities**, not accumulated `float64` fields. The business needs a breakdown per income source and per creditor (to compute payment capacity, prioritize debts, and report each income separately for tax purposes).
- The existing layered architecture in `services/core` (`domain` → `usecase` → `delivery` → `infra`) is kept as-is; it's being completed, not changed.
- **Schema migrations are versioned with [goose](https://github.com/pressly/goose)**, adopted before any real environment/data exists specifically so the product can grow into deployments with real data (including, eventually, per-company environments) without a painful cutover later. Postgres no longer auto-applies SQL on container init (`docker-entrypoint-initdb.d` mount removed from `docker-compose.yml`); the migrations folder is the single source of truth, applied via `make migrate-up` / `migrate-down` / `migrate-status`. Files follow goose's `NNNNN_description.sql` naming with `-- +goose Up` / `-- +goose Down` sections. Requires the `goose` CLI (`go install github.com/pressly/goose/v3/cmd/goose@latest`).

## Data model (Phase 1)

Target schema for task 6 (migration `00002_*.sql`, incremental on top of `00001_init.sql`):

```sql
users
  id            SERIAL PRIMARY KEY
  email         VARCHAR(255) NOT NULL UNIQUE
  name          VARCHAR(255) NOT NULL
  tax_regime    VARCHAR(50) NOT NULL DEFAULT ''
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP

incomes                              -- mirrors domain.Income
  id            SERIAL PRIMARY KEY
  user_id       INTEGER NOT NULL REFERENCES users(id)
  amount        NUMERIC(12,2) NOT NULL
  type          VARCHAR(50) NOT NULL
  date          DATE NOT NULL
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP

debts                                -- mirrors domain.Debt
  id            SERIAL PRIMARY KEY
  user_id       INTEGER NOT NULL REFERENCES users(id)
  creditor      VARCHAR(255) NOT NULL
  balance       NUMERIC(12,2) NOT NULL
  installment   NUMERIC(12,2) NOT NULL
  due_date      DATE NOT NULL
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

Plus `user_id INTEGER NOT NULL REFERENCES users(id)` added to the existing `movements` and `limits` tables.

Note: `domain.Income.Date` and `domain.Debt.DueDate` are Go `string` fields, but the DB columns are real `DATE` — the string↔date conversion happens in the `pgrepo` layer, not by storing dates as text.

**Primary keys are `UUID`, not `SERIAL`** (decided for all tables, including `movements`/`limits`, for schema consistency). Reasoning: non-guessable IDs (no row-count leakage through URLs/APIs) and no collision risk if data from separate environments/companies is ever merged or exported — both matter given the B2B growth goal. Prefer `UUID v7` (time-ordered) over `v4` (random) for better index locality; Postgres 16 (the version pinned in `docker-compose.yml`) has no native `uuidv7()`, so **the application generates the ID**, not the database — no `DEFAULT` on `id` columns. This is still pending wiring (likely `github.com/google/uuid`'s `uuid.NewV7()`) once the `pgrepo`/`delivery` layers are implemented; until then, inserting rows requires supplying `id` explicitly.

## Phase 1 — Tasks

Status as of 2026-08-06:

1. [x] Model `Debt` in `domain` (creditor, balance, installment, due date)
2. [x] Change `Advisor.Income float64` to `Advisor.Incomes []Income` (the `Income` type already exists in `asesor.go`, just needs to be wired in)
3. [x] Add `Advisor.Debts []Debt` plus aggregation methods: total income, total debt, payment capacity (income - fixed expenses - debt installments)
4. [x] Update `AdvisorUseCase` (`RegisterIncome`, `CalculateTax`) to operate on the collections instead of the accumulator — done as part of task 2's work
5. [x] Unit tests (`calculator_test.go`, plus a new one for `usecase`)
6. [x] Migration (goose): `users`, `incomes`, `debts` tables, plus `user_id` added to `movements` and `limits`

## Change history

- **2026-08-06**: Created the log. Defined the architecture decision: B2C first, base prepared for B2B (`user_id` from day one, no multi-tenancy yet).
- **2026-08-06**: `Advisor.Income` (single `float64`) replaced by `Advisor.Incomes []Income`. Added `Advisor.TotalIncome()` domain method to sum the collection. Updated `RegisterIncome` (now appends a batch of incomes) and `CalculateTax` (now uses `TotalIncome()`) in `AdvisorUseCase` accordingly.
- **2026-08-06**: Added `Debt` entity in `domain/deuda.go` (creditor, balance, installment, due date). Not wired into `Advisor` yet — that's task 3. Regime names were also extracted by the user into `internal/const/advisor/taxes` as string constants (`SmallContributor`, `OptionalSimplified`, `ProfitsTax`), used now in `CalculateTax`.
- **2026-08-06**: Full English pass across the project — no Spanish left in file names, identifiers, DB schema, or frontend routes/text:
  - Go domain/usecase/infra files renamed (`asesor.go`→`advisor.go`, `limites.go`→`limit.go`, `tarifario.go`→`tariff.go`, `regimen.go`→`regime.go`, `calculadoras.go`→`calculator.go`, `deuda.go`→`debt.go`, `movimiento.go`→`movement.go`, plus matching `_repo.go`/`_test.go`/`_usecase.go` files).
  - `Advisor.Regimen` renamed to `Advisor.TaxRegime`.
  - SQL migration (`001_init.sql`): tables/columns translated (`movimientos`→`movements`, `limites`→`limits`, `tipo`→`type`, `monto`→`amount`, `categoria`→`category`, `descripcion`→`description`, `valor`→`value`). Edited in place since no DB volume had been created yet.
  - Frontend routes renamed (`finanzas`→`finance`, `limites`→`limits`, `movimientos`→`movements`) and all visible UI text translated to English.
- **2026-08-06**: `Advisor.Debts []Debt` wired in, plus `TotalDebt()` (sum of balances, for debt-to-income ratio), `TotalDebtService()` (sum of installments, the actual monthly outflow), and `PaymentCapacity()` (`TotalIncome() - Expenses - TotalDebtService()`). `RegisterDebt` added to `AdvisorUseCase`, mirroring `RegisterIncome`.
- **2026-08-06**: Task 4 confirmed already satisfied by task 2's work (`RegisterIncome`/`CalculateTax` already operate on `Incomes`/`TotalIncome()`).
- **2026-08-06**: Unit tests written — `domain/calculator_test.go` covers `TotalIncome`, `TotalDebt`, `TotalDebtService`, `PaymentCapacity`; new `usecase/advisor_usecase_test.go` covers `CalculateTax` (all regimes + unknown/no-income), `RegisterIncome`, `RegisterDebt`, and `RegisterExpense` (limit notification, per-category isolation, accumulation). All pass (`go test ./...`).
- **2026-08-06**: Adopted goose for schema migrations (see architecture decisions). Converted the existing schema (`movements`, `limits`) into `00001_init.sql` with proper `Up`/`Down` sections; removed the `docker-entrypoint-initdb.d` auto-run from `docker-compose.yml`; added `migrate-up`/`migrate-down`/`migrate-status` targets to the `Makefile` (they read `DATABASE_URL` from `.env`, gitignored, copied from `.env.example`). Verified locally: `up` → `status` → `down` → `up` all behave correctly against a real Postgres container. The `users`/`incomes`/`debts` migration itself (task 6) is still pending, now as `00002_*.sql`.
- **2026-08-06**: Task 6 completed — `00002_add_users_debts.sql` creates `users`, `incomes`, `debts`, and rebuilds `movements`/`limits` with `user_id` + `UUID` primary keys (see "Data model" section for the UUID v7 decision). Since `movements`/`limits` held no data, they were dropped and recreated rather than altered in place — the `Down` migration restores their exact original `SERIAL`-based shape from `00001_init.sql`. Verified locally: `up` → inspected all 5 tables' final structure → `down` (confirmed `movements`/`limits` revert correctly) → `up` again. Phase 1 (B2C foundations) is now complete.
