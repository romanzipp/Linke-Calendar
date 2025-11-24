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
    url: "https://www.die-linke-fulda.de/"
    zetkin_organization: 192

scraper:
  interval: "6h"
  timeout: "30s"

server:
  port: "8080"
  host: "0.0.0.0"
```

## API Endpoints

- `GET /health` - Health check endpoint
- `GET /site/{siteID}/calendar` - Calendar view for a specific site
  - Query params: `year`, `month` (optional)
- `GET /site/{siteID}/list` - List view showing all upcoming events in chronological order
- `GET /site/{siteID}/ical` - iCal endpoint for subscribing with mobile device
- `GET /event/{eventID}` - Event detail modal
- `GET /static/*` - Static files (CSS, JS, fonts)

## Embedding

To embed the calendar view in your website:

```html
<iframe src="http://your-server:8080/site/fulda/calendar" width="100%" height="800" frameborder="0"></iframe>
```

To embed the list view of upcoming events:

```html
<iframe src="http://your-server:8080/site/fulda/list" width="100%" height="600" frameborder="0"></iframe>
```

## License

MIT
