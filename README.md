
<p align="center">
<img src="frontend/wizard-vue/src/assets/logo.png" alt="logo" width="300"/>
</p>

---

**RSSAlchemy** is a website-to-rss converter, like RSSHub, RSS-bridge or Rss.app. Here are main features:

- Convert arbitrary website to RSS feed using CSS selectors
- Dynamic websites are supported using headless chrome (playwright)
- Cookies[^1] (supports scraping private feeds, eg youtube subscriptions)
- Proxy
- Results caching
- Screenshots (primarily for debugging)
- [Presets](presets) for sharing configurations
- Stateless[^2] (all task parameters are encoded into url, no database needed)
- Distruibuted by design (deploy as many workers as you need)
- Self-hosted; easy to deploy; docker-compose provided
- Relatively small codebase, written in go + typescript
- Security and reliability:
  - Rate-limit by source client IP
  - Rate-limit by target domain (to prevent 429 if many tasks target the same site)
  - Block service workers
  - Prevent WebRTC leak if using proxy
  - Block localhost and private IPs (including proxy server's internal services)
  - Chrome is sandboxed; container is UNprivileged

[^1]: Cookies require support from your RSS reader/aggregator. Miniflux works, others are not checked yet.
[^2]: Nats KV is used to store cookies permanently, it's required for sites that update cookies on every request, like
youtube

| feature/program      | RSS Alchemy               | RSS Hub                      | RSS-Bridge              | RSS.app       |
|----------------------|---------------------------|------------------------------|-------------------------|---------------|
| Custom websites      | ✅ (using CSS selectors)   | ❌ (only hardcoded site list) | ✅ (using CSS selectors) | ✅             |
| Render dynamic sites | ✅ (using headless chrome) | ❌                            | ❌                       | ✅             |
| Hosting              | Self-hosting              | Self-hosting                 | Self-hosting            | Only cloud    |
| Price                | Free and open-source      | Free and open-source         | Free and open-source    | Paid ($8/mon) |

## Project status

Program is still under development. The code architecture is not final, no CI, no demo page, etc...

## Deployment

```bash
git clone https://github.com/egor3f/rssalchemy
cd rssalchemy/deploy
docker-compose up -d
```

Then open your browser targeting to port 8080.

For SSL, authentication, domains, etc. - use Caddy or Nginx (no specific configuration required). Personally I recommend Caddy, if you haven't used it before - give it a try :)

### Configuration

Configuration is done using environment variables

You can see all available options in [config.go file](internal/config/config.go) (struct Config)

Docker-compose deployment uses [deploy/.env file](deploy/.env)

### Scaling

Each worker can process 1 page at a time, so to scale you should run multiple worker instances. This is done using replicas parameter in worker section in [docker-compose.yml file](deploy/docker-compose.yml)

## Development

You need 
- Go 1.23 (most of application)
- Node.js 20 (frontend)
- Nats (with jetstream)
- Redis

Instaling dependencies example for MacOS:

```bash
brew install go@1.23
brew install node@20
brew install redis
brew install nats-server  # Don't use brew services to manage nats because it lacks config support
go mod download
cd frontend/wizard-vue && npm install
nats -js
```

Also this repository contains some useful git hooks. To enable them, use:
```bash
git config --local core.hooksPath .githooks/
```
