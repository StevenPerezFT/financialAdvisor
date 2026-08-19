# Project Log

This file tracks decisions and progress. Updated at each work session.

## Product goal

- **Phase 1 (current) — B2C**: help individuals with multiple income sources and multiple debts understand their financial situation and tax obligations.
- **Phase 2 (future) — B2B**: package the above as a tool for financial advisory companies (multi-client, payroll/employees). Not implemented yet, but Phase 1 decisions must leave room for it.

## Architecture decisions

- **No multi-tenancy yet**, but a `User` entity is introduced now, and everything (`incomes`, `debts`, `movimientos`, `limites`) is related through `user_id`. This way, adding a `company_id` later is an additive migration, not a rewrite of the model.
- **Incomes and debts are collections of their own entities**, not accumulated `float64` fields. The business needs a breakdown per income source and per creditor (to compute payment capacity, prioritize debts, and report each income separately for tax purposes).
- The existing layered architecture in `services/core` (`domain` → `usecase` → `delivery` → `infra`) is kept as-is; it's being completed, not changed.
- **API transport is gRPC**, following the scaffold's existing `delivery/rpc` folders and `finance.proto` (empty until now). Contracts are defined with [buf](https://buf.build) (already installed locally), not raw `protoc`, and generated Go code lands in `services/core/internal/pb` — kept inside the single existing Go module rather than as a separate module/workspace, since there's only one Go service today (YAGNI on cross-module wiring until a second service actually needs the generated code).
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

**Primary keys are `UUID`, not `SERIAL`** (decided for all tables, including `movements`/`limits`, for schema consistency). Reasoning: non-guessable IDs (no row-count leakage through URLs/APIs) and no collision risk if data from separate environments/companies is ever merged or exported — both matter given the B2B growth goal. Prefer `UUID v7` (time-ordered) over `v4` (random) for better index locality.

Originally the plan was app-side generation only (Postgres 16, pinned at the time, has no native `uuidv7()`). Revisited: relying solely on the application to supply `id` was flagged as a real risk (any direct insert bypassing app code — manual fix, admin tool, future service — would violate `NOT NULL` on `id`). **Upgraded Postgres 16 → 18 specifically to get the native `uuidv7()` function**, and added `DEFAULT uuidv7()` on every `id` column (migration `00003_uuid_defaults.sql`) as a DB-level safety net. The app can still generate its own `id` client-side when it wants to (e.g. to return it before the insert completes), but nothing depends on it doing so anymore. `docker-compose.yml` also needed its Postgres volume mount updated from `/var/lib/postgresql/data` to `/var/lib/postgresql` — Postgres 18's image changed the expected mount point (versioned subdirectories now live under it, to support in-place `pg_upgrade`).

## Phase 1 — Tasks

Status as of 2026-08-06:

1. [x] Model `Debt` in `domain` (creditor, balance, installment, due date)
2. [x] Change `Advisor.Income float64` to `Advisor.Incomes []Income` (the `Income` type already exists in `asesor.go`, just needs to be wired in)
3. [x] Add `Advisor.Debts []Debt` plus aggregation methods: total income, total debt, payment capacity (income - fixed expenses - debt installments)
4. [x] Update `AdvisorUseCase` (`RegisterIncome`, `CalculateTax`) to operate on the collections instead of the accumulator — done as part of task 2's work
5. [x] Unit tests (`calculator_test.go`, plus a new one for `usecase`)
6. [x] Migration (goose): `users`, `incomes`, `debts` tables, plus `user_id` added to `movements` and `limits`

## Phase 1 — Wiring (gRPC + Postgres)

Goal: turn the domain/tests/schema built so far into an app that actually runs against real Postgres, end to end. `main.go` currently just prints a line; nothing persists or serves requests yet.

Status as of 2026-08-06:

1. [x] Postgres connection: driver dependency + a pool opened from `DATABASE_URL`, usable from `main.go`
2. [x] gRPC toolchain: install `protoc-gen-go`/`protoc-gen-go-grpc`, set up `buf.yaml`/`buf.gen.yaml` in `packages/proto`, generate into `services/core/internal/pb`
3. [x] Proto contracts: `advisor.proto` (RegisterIncomes, RegisterDebts, CalculateTax) and `finance.proto` (RegisterMovements) — done. `RegisterExpense` (advisor) has no proto contract yet, not tracked as part of this task's scope; add later if needed
4. [x] `pgrepo`: Postgres-backed persistence for the aggregates needed by the above — `users`, `incomes`, `debts`, `limits`, `movements`
5. [x] `delivery/rpc` handlers: implement the generated gRPC service interfaces, wiring usecase + repo — `CalculateTax`, `RegisterIncomes`, `RegisterDebts`, `RegisterMovement`
6. [ ] `main.go`: load env, open the DB pool, build the repo → usecase → handler graph, start the gRPC server, graceful shutdown
7. [ ] End-to-end smoke test against real Postgres (e.g. `grpcurl`)

Order matters: 1 and 2 are independent and foundational; 3 needs 2; 4 needs 1; 5 needs 3 and 4; 6 needs 5; 7 needs 6.

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
- **2026-08-06**: Wiring task 1 done — `internal/db/connection.go` (`Dbc`) opens a `pgxpool.Pool` from `DATABASE_URL` and returns it (error handling left to the caller, no `os.Exit` inside the package, so it's reusable/testable). Verified with a throwaway smoke test against the real Postgres container (connect + ping succeeded), removed after confirming.
- **2026-08-06**: Wiring task 2 done — installed `protoc-gen-go`/`protoc-gen-go-grpc`; added `packages/proto/buf.yaml` (module) and `buf.gen.yaml` (generates into `services/core/internal/pb`, `paths=source_relative`). Gave `finance.proto` a minimal valid placeholder (`syntax`/`package`/`go_package`) just to prove the pipeline — no messages/services yet, that's task 3. Added `google.golang.org/protobuf` and `google.golang.org/grpc` as Go module deps (needed to compile the generated code, not just to generate it). Verified: `buf lint`, `buf generate`, and `go build ./...` all pass.
- **2026-08-06**: Renamed the proto package from `aprexa` to `financialadvisor` — `aprexa` turned out to be a leftover codename from the original (pre-session) scaffold commit, unrelated to the product name. Moved `packages/proto/aprexa/` → `packages/proto/financialadvisor/`, updated `package financialadvisor.finance.v1;` in `finance.proto`, regenerated (old output at `internal/pb/aprexa/` removed, new one at `internal/pb/financialadvisor/finance/v1/`). Verified: `buf lint`, `buf generate`, `go build`/`vet`/`test` all pass.
- **2026-08-06**: `advisor.proto` completed — `CalculateTax`, `RegisterIncomes` (`Income` message mirrors `domain.Income`), `RegisterDebts` (`Debt` message mirrors `domain.Debt`). Added `finance/usecase/finance_usecase.go` (`MovementUseCase.RegisterMovement`, first code in that module's `usecase` layer) and `finance/domain/movement.go` (`Movement` struct mirroring the `movements` table). `finance.proto` written from that usecase's real signature: `RegisterMovements` returns the registered `Movement` (with its generated id) rather than an invented total, since nothing in the usecase computes one and the caller needs the id back to reference the record later. Verified: `buf lint`, `buf generate`, `go build`/`vet`/`test` all pass. Task 3 complete except `RegisterExpense`, left out of scope for now.
- **2026-08-06**: Upgraded Postgres 16 → 18 (see "Data model" for why) and added `00003_uuid_defaults.sql` (`DEFAULT uuidv7()` on every `id` column). Also fixed `docker-compose.yml`'s volume mount for the new image (`/var/lib/postgresql`, not `/var/lib/postgresql/data`) and unified the `advisor/infra/pgrepo` package name to `pgrepo` (was inconsistently `infra`, didn't match its own folder name or `finance/infra/pgrepo`). Verified end to end: recreated the Postgres volume, ran all 3 migrations from scratch, inserted a row without supplying `id` and confirmed Postgres generated a real v7 UUID (version nibble `7`), tested `down`/`up` on the new migration, `go build`/`vet`/`test` all pass.
- **2026-08-06**: `advisor/infra/pgrepo` repos written — `IncomeRepo`, `DebtRepo`, `LimitRepo`, `UserRepo`, all `Create(ctx, ...) (domain.X, error)` using `RETURNING id` (Postgres generates the id via the `00003` default, not the app). Added `internal/const/advisor/dateformat` (`Layout = "2006-01-02"`) to replace a broken earlier attempt at the same idea (`c_usecase`: wrong import name, unexported constant) — both `IncomeRepo`/`DebtRepo` parse their date/due-date strings through it. Fixed `domain.User` field naming to match the rest of the domain (`UserID`→`Id`, `Tax_Regime`→`TaxRegime`, dropped an unused `Created_At`). Every repo verified with a throwaway smoke test against real Postgres (insert + confirm a v7 id came back), removed after confirming. `MovementRepo` (finance module) is still an empty stub — not written yet, only sketched as an example in conversation.
- **2026-08-06**: `MovementRepo` written — first attempt landed in the wrong bounded context (`advisor/infra/pgrepo/movement_repo.go`, importing `advisor/domain`, plus a duplicate `Movement` struct created in `advisor/domain/movement.go` just to make that import compile). Moved to `finance/infra/pgrepo/movement_repo.go` importing `finance/domain` (where `Movement`, `MovementUseCase`, and `finance.proto` already live), duplicate deleted. Verified against real Postgres. Task 4 (pgrepo) complete — all 5 repos (`users`, `incomes`, `debts`, `limits`, `movements`) written and confirmed against a live database.
- **2026-08-06**: Built `AdvisorHandler.CalculateTax` (`advisor_handler.go`) as the reference example for task 5's remaining handlers. Needed two new read methods not covered by task 4 (which only asked for `Create`): `UserRepo.FindByID` and `IncomeRepo.ListByUserID` (the latter converts the DB `date` column back to the `dateformat.Layout` string domain.Income expects). Handler loads user + incomes from Postgres, builds a `*domain.Advisor`, calls the existing `AdvisorUseCase.CalculateTax` — no new business logic, purely wiring. Verified end to end against real Postgres: inserted a user (Small Contributor) and a 1000 income, called the handler, got `tax = 50` back correctly. `RegisterIncomes`, `RegisterDebts` (advisor) and `RegisterMovements` (finance) still need their handlers, following this same shape.
- **2026-08-06**: Implemented `AdvisorHandler.RegisterIncomes`. The user got stuck twice on the same broken draft (`req.Incomes`, a `[]*advisorv1.Income`, passed directly into `IncomeRepo.Create`, which takes one `domain.Income`), so I wrote it directly: loop over `req.Incomes`, map each proto `Income` to `domain.Income` (field names happen to match 1:1), `Create` one at a time, then re-read with `ListByUserID` to compute `TotalIncome()` for the response (needs the full post-insert set, not just what was just added). Verified against real Postgres: registered two incomes (1000 + 250.50), got `total_income = 1250.50` back.
- **2026-08-06**: Reworked the `Register*` response contract after the user correctly questioned why a "create" operation needed to read anything back at all. `RegisterIncomesResponse`/`RegisterDebtsResponse` no longer return a computed total (`total_income`/`total_debt`) — they now return the created `Income`/`Debt` records themselves (with generated `id`), matching the pattern `finance.proto`'s `RegisterMovementsResponse` already used. This removes the `ListByUserID` call from both handlers entirely — a `Register*` handler now only loops, maps proto→domain, and calls `Create`; no read logic left inside a write operation. (Computing "total income across everything" stays `CalculateTax`'s job, which already reads by design.) Also fixed `RegisterDebts`, which was left mid-edit (missing `Create` call, no `return`, and the constructor didn't wire `debtsRepo` at all — would've been a nil-pointer panic). Verified end to end against real Postgres: registered one income and one debt, both came back with generated ids.
- **2026-08-06**: `finance.proto`'s RPC renamed `RegisterMovements` → `RegisterMovement` (singular) after the user asked why the request wasn't a list — it's singular by design, mirroring `MovementUseCase.RegisterMovement` (registers one at a time, same as `advisor.proto`'s `RegisterIncomes`/`RegisterDebts` do per call, just not batched). The plural RPC name next to a singular payload was the actual source of the confusion; renaming it to match resolves that. While fixing this, found `RegisterMovementRequest` had no `user_id`/`advisor_id` field at all — unlike every `advisor.proto` request — so the handler had no way to know whose movement it was, and `movements.user_id` is `NOT NULL`. Added `string user_id = 1;` to the request. Rewrote `FinanceHandler`: renamed `Create`→`RegisterMovement` to satisfy the generated interface (name and `(*Response, error)` signature both required), swapped `UnsafeFinanceServiceServer`→`UnimplementedFinanceServiceServer` (the generated comment explicitly says Unsafe isn't recommended), fixed the constructor (had an accidental method receiver). Verified against real Postgres: registered a movement, got it back with a generated id. Task 5 (delivery/rpc handlers) complete.
