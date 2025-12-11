# OpenSpec: Trading Decision Credit Consumption

## 1. Background
To monetize the AI trading service and manage resource usage, we need to implement a credit consumption mechanism. Specifically, when a "TopTrader" (or any configured trader) makes a trading decision using the AI model, it should consume credits from the user's account.

## 2. Requirements
1.  **Trigger:** Credit consumption occurs when the AI executes a decision cycle (regardless of whether a trade is placed, as the AI inference cost is incurred).
2.  **Configuration:** The cost per decision should be configurable in the database (`system_config` table), key: `trading_decision_points_cost`, default: `1`.
3.  **Target:** Initially targets "TopTrader", but should be applicable to any trader based on configuration (for now, we'll enforce it for "TopTrader").
4.  **Action:**
    *   Deduct `X` credits from `Available Credits`.
    *   Increase `Used Credits` by `X`.
    *   If credits are insufficient, stop/skip the decision cycle and log a warning.
5.  **User Interface:** The user profile page (`/profile`) should reflect these changes in real-time (already handled by existing API if backend updates DB).

## 3. Implementation Design

### 3.1 Database & Configuration
*   **File:** `config/database.go`
*   **Change:** Add `trading_decision_points_cost` to `systemConfigs` map in `initDefaultData`.

### 3.2 Service Layer
*   **File:** `trader/auto_trader.go`
*   **Struct:** `AutoTrader`
*   **Changes:**
    *   Add `creditService` (of type `credits.Service`) to `AutoTrader` struct.
    *   Add `userID` to `AutoTrader` (it seems `trader` doesn't explicitly store `userID` in `AutoTrader` struct, but `TraderManager` knows it. We need `userID` to deduct credits).
    *   *Correction:* `AutoTraderConfig` usually has `ID`, but `UserID` might be implicit or part of `ID`. Looking at `manager/trader_manager.go`, `traderCfg.UserID` exists. We need to pass `UserID` to `AutoTrader`.
    *   Inject `config.Database` into `NewAutoTrader` to initialize `CreditService`.

### 3.3 Execution Logic
*   **Method:** `AutoTrader.runCycle()`
*   **Logic:**
    1.  At the start of `runCycle`, check if credit consumption is enabled for this trader (e.g., name is "TopTrader").
    2.  Fetch `trading_decision_points_cost` from system config.
    3.  Call `creditService.DeductCredits(userID, cost, "decision", ...)`
    4.  If error (insufficient funds), log error and return immediately (skip AI call).

### 3.4 Manager Layer
*   **File:** `manager/trader_manager.go`
*   **Changes:**
    *   Pass `database` instance and `UserID` to `AutoTraderConfig` when creating new traders.

## 4. Testing Plan
*   **Unit Test:** Create a test that initializes an `AutoTrader` with a mock `CreditService` or a real DB connection and verifies `runCycle` fails when credits are 0.
*   **Integration:** Run the system and observe "TopTrader" consuming credits in the logs/DB.

## 5. Constraints
*   **KISS:** Reuse existing `CreditsService`.
*   **Safety:** Ensure `TopTrader` stops if credits are empty to prevent free usage.
