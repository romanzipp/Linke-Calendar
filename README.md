# Linke Calendar

Ein Go Webserver, der Termine von Zetkin scraped und als `<iframe>` einbettbare Seite bereitstellt.

## Funktionen

- Anbindung an Zetkin inklusive Filterung nach KV/LV ("Organisation")
- Einbettung des Kalenders via `<iframe>` in die Website
- Stellt iCal-Link zur Verfügung, womit der Kalender in mobilen Kalender-Apps abonniert werden kann

![Vorschau Screenshot](/art/preview.png)

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
- `GET /cal/{siteID}/ical` - iCal endpoint for subscribing with mobile device
- `GET /event/{eventID}` - Event detail modal
- `GET /static/*` - Static files (CSS, JS, fonts)

## Embedding

To embed the calendar in your website:

```html
<iframe src="http://your-server:8080/cal/fulda" width="100%" height="800" frameborder="0"></iframe>
```

## License

MIT
