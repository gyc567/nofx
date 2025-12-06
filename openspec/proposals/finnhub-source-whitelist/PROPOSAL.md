# Proposal: Finnhub News Source Whitelist Filtering

## 1. Goal
Enhance the `FinnhubFetcher` to filter news articles based on a strict whitelist of authoritative sources. This ensures that the trading agent acts only on high-quality, verified information, reducing the risk of reacting to rumors or low-quality news.

## 2. Whitelist Sources
The following authoritative sources are allowed:
- Reuters
- Bloomberg
- Financial Times (FT)
- Wall Street Journal (WSJ)
- CNBC
- Caixin
- South China Morning Post (SCMP)
- Xinhua
- BBC
- The Economist
- IMF
- World Bank

## 3. Implementation Plan

### 3.1 Logic Update in `FetchNews`
1.  **Define Whitelist**: Create a constant or variable slice containing the allowed source names (normalized for matching).
2.  **Source Verification**: Within the article processing loop (after keyword filtering), check the `article.Source` field.
3.  **Filtering**: Discard articles where the `Source` does not contain any of the whitelisted strings (case-insensitive).

### 3.2 Configuration
The whitelist will be defined in `service/news/finnhub.go` for now, but structured to be easily externalized if needed later.

### 3.3 Testing
Add unit tests to verify:
- Articles from allowed sources are kept.
- Articles from disallowed sources are rejected.
- Keyword filtering still works in conjunction with source filtering.

## 4. Performance & Reliability
- **Performance**: String matching is efficient enough for the volume of news returned by Finnhub API (typically < 100 items per call).
- **Reliability**: Ensure matching is robust (e.g., handling "Reuters via Yahoo" if applicable, though Finnhub usually gives the primary source). We will use case-insensitive substring matching or exact matching depending on observation. Finnhub `source` field is usually clean. Substring match is safer.

## 5. Integration
This filter runs **after** the keyword filter to minimize processing. If an article is irrelevant content-wise, its source doesn't matter.

