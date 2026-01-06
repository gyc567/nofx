# Bug Proposal: Deleted AI Models and Exchanges Persistence

## Problem Description
In the `AITradersPage`, when a user deletes (disables) an AI Model or Exchange configuration, the item still remains in the "AI Models" or "Exchanges" configuration cards at the top of the page. This is because the frontend displays all retrieved configurations from the backend, including those where `enabled` is false. Users expect deleted configurations to disappear from these lists.

Additionally, if a trader is deleted while the user is viewing its details (dashboard), the application stays on the dashboard with stale data or a loading skeleton instead of navigating back to the traders list.

## Proposed Changes

### Frontend (web/src/components/AITradersPage.tsx)
1.  Update the filtering logic for `configuredModels` and `configuredExchanges`:
    - Only include items where `enabled` is true.
2.  Add an `onTraderDelete` callback to `AITradersPageProps`.
3.  In `handleDeleteTrader`, call `onTraderDelete(traderId)` after successful deletion.

### Frontend (web/src/App.tsx)
1.  Implement the `onTraderDelete` callback for `AITradersPage`.
2.  If the deleted trader matches `selectedTraderId`, clear `selectedTraderId` and navigate the user back to the `/traders` page.

## Implementation Plan
1.  Modify `web/src/components/AITradersPage.tsx` to filter the displayed models and exchanges.
2.  Update `AITradersPage` component and its usage in `App.tsx` to handle trader deletion navigation.
3.  Verify the fix by simulating the deletion process.
