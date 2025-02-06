# RSS Alchemy

## About

RSS Alchemy is a website-to-rss converter, like RSSHub, RSS-bridge or Rss.app. Here is the difference:

- Convert arbitrary website to RSS feed using CSS selectors
- Dynamic websites are supported using headless chrome (playwright)
- Cookies[^1] (supports scraping private feeds, eg youtube subscriptions)
- Proxy
- Results caching
- Screenshots (primarily for debugging)
- Stateless[^2] (all task parameters are encoded into url, no database needed)
- Distruibuted by design (deploy as many workers as you need)
- Self-hosted; easy to deploy; docker-compose provided
- Relatively small codebase, written in go + typescript

[^1]: Cookies require support from your RSS reader/aggregator. Miniflux works, others are not checked yet.
[^2]: Nats KV is used to store cookies permanently, it's required for sites that update cookies on every request, like youtube 

| program/feature      | RSS Alchemy               | RSS Hub                      | RSS-Bridge              | RSS.app       |
|----------------------|---------------------------|------------------------------|-------------------------|---------------|
| Custom websites      | ✅ (using CSS selectors)   | ❌ (only hardcoded site list) | ✅ (using CSS selectors) | ✅             |
| Render dynamic sites | ✅ (using headless chrome) | ❌                            | ❌                       | ✅             |
| Hosting              | Self-hosting              | Self-hosting                 | Self-hosting            | Only cloud    |
| Price                | Free and open-source      | Free and open-source         | Free and open-source    | Paid ($8/mon) |

## Project status
Program is still under development. The code architecture is not final, tests are missing, no CI, no demo page, no docs, etc...
