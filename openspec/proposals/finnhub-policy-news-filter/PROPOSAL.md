# Proposal: Finnhub News Filtering for Major Financial Policies

## 1. Goal
Optimize the `FinnhubFetcher` in `service/news/finnhub.go` to filter news articles, focusing only on core financial policies from the US, China, and Europe. This reduces noise and highlights critical market-moving information (e.g., interest rate decisions, central bank statements).

## 2. Filtering Logic
Based on the provided Python reference, the filtering logic will be:

### 2.1 Keywords
Only articles containing at least one of the following keywords (case-insensitive) in their `headline` or `summary` will be retained:
- **US**: "Fed", "FOMC", "Powell", "interest rate decision", "rate cut", "rate hike"
- **China**: "PBOC", "People's Bank of China", "LPR", "reserve requirement ratio", "China stimulus"
- **Europe**: "ECB", "European Central Bank", "Lagarde", "Eurozone", "EU regulation", "monetary policy"

### 2.2 Region Tagging (Optional Enhancement)
While the `Article` struct doesn't strictly require a region tag, we can prepending the region to the `Headline` or strictly using the keywords to filter. The primary goal is to filter the list returned by `FetchNews`.

### 2.3 Time Range
The Python script fetches the last 48 hours. The current `FetchNews` implementation just calls the API. We should ensure we are processing relevant news. The Finnhub API `news` endpoint returns the latest news. We can implement client-side filtering on the returned results.

## 3. Implementation Details

### 3.1 Modify `FinnhubFetcher`
- Add a `filterPolicyNews` method or integrate logic into `FetchNews`.
- Since `FetchNews` takes a `category`, we can apply this filter specifically when `category` is `general`, or add a configuration to enable it. However, given the specific request, we might enforce this for the `general` category or update the fetcher to always prioritize these topics if that's the intended role of this fetcher in the system.
- **Assumption**: The user wants this specific filtering applied to the `FinnhubFetcher`. We will apply it to the `FetchNews` method.

### 3.2 Code Changes in `service/news/finnhub.go`
1.  Define the `policyKeywords` slice.
2.  In `FetchNews`, after decoding the JSON response:
    - Iterate through the articles.
    - Check if `Headline` + `Summary` contains any keyword.
    - If yes, keep the article.
    - If no, discard.
3.  (Optional) We can also infer the region and prepend it to the title if that helps downstream processing, similar to the python script printing `[Region]`. Let's append `[Region]` to the start of the headline for clarity.

## 4. Verification
- Write a test case or run the code to verify it filters out irrelevant news.

