# Proposal: Finnhub News Keyword Expansion & Negative Filtering

## 1. Goal
Enhance the relevance of news filtered by `FinnhubFetcher` by expanding the keyword list to include broader macroeconomic indicators and geopolitical terms, while introducing negative filtering to exclude irrelevant topics (e.g., sports, entertainment).

## 2. Keyword Strategy

### 2.1 Expansion (Inclusion)
We will expand the existing `policyKeywords` to cover:
- **US Indicators**: CPI, PCE inflation, non-farm payrolls, Treasury yields, fiscal deficit, trade tariffs.
- **China Indicators**: GDP growth, export data, Belt and Road, real estate policy, tech regulation.
- **Europe Indicators**: CPI, PMI, Brexit, EU fiscal rules, energy policy.
- **Global/Macro**: Monetary easing, fiscal stimulus, trade war, sanctions, geopolitical tensions, IMF forecast, WTO ruling, supply chain disruption, climate policy.

### 2.2 Exclusion (Negative Filtering)
To ensure high signal-to-noise ratio, we will explicitly exclude articles containing terms related to non-financial topics, even if they come from authoritative sources.
- **Excluded Terms**: "sports", "entertainment", "celebrity", "movie", "music", "football", "basketball", "soccer", "cricket".

## 3. Implementation Plan

### 3.1 Code Changes in `service/news/finnhub.go`
1.  **Update `policyKeywords`**: Merge the new high-value keywords into the existing list.
2.  **Define `excludedKeywords`**: Create a new constant/variable list for negative terms.
3.  **Update `FetchNews` Logic**:
    - **Step 1**: Source Whitelist Check (Existing).
    - **Step 2**: Exclusion Check. Iterate through `excludedKeywords`. If the title/summary contains any, discard the article.
    - **Step 3**: Inclusion Check. Iterate through the expanded `policyKeywords`. Keep if matched.
    - **Step 4**: Region Tagging (Existing).

### 3.2 Testing
- Verify that articles with new keywords (e.g., "CPI", "supply chain") are captured.
- Verify that articles with excluded keywords (e.g., "World Cup", "movie star") are rejected, even if they mention "China" or come from "Reuters".

## 4. Design Principles
- **KISS**: Simple string matching lists. No complex regex unless necessary.
- **High Cohesion**: All filtering logic remains encapsulated within `FinnhubFetcher`.
- **Performance**: String checks are lightweight. Source filtering first reduces the dataset significantly.

