# Linke Calendar

Ein Go Webserver, der Termine eines Typo3 Frontend scraped und als `<iframe>` einbettbare Seite bereitstellt.

![](/art/preview.png)

## Configuration

Create a `config.yaml` file based on `config.yaml.example`:

```yaml
sites:
  - id: "fulda"
    name: "Die Linke Fulda"
    url: "https://www.die-linke-fulda.de/termine/seite/{page}/"
    zetkin_enabled: true
    zetkin_cookie: "Fe....645"
    zetkin_organization: "KV Fulda"

scraper:
  interval: "6h"
  max_pages: 10
  timeout: "30s"

server:
  port: "8080"
  host: "0.0.0.0"
```

Zetkin: If enabled, provide the Zetkin Session ID (Cookie string after `zsid=`) and an organization name for filtering.

## API Endpoints

- `GET /health` - Health check endpoint
- `GET /cal/{siteID}` - Calendar view for a specific site
  - Query params: `year`, `month` (optional)
- `GET /event/{eventID}` - Event detail modal
- `GET /static/*` - Static files (CSS, JS, fonts)

## Embedding

To embed the calendar in your website:

```html
<iframe src="http://your-server:8080/cal/fulda" width="100%" height="800" frameborder="0"></iframe>
```

## License

MIT
