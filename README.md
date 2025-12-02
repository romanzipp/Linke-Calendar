# Linke Calendar

Ein Go Webserver, der Termine von Zetkin scraped und als `<iframe>` einbettbare Seite bereitstellt.

## Funktionen

- Anbindung an Zetkin inklusive Filterung nach KV/LV ("Organisation")
- Einbettung des Kalenders via `<iframe>` in die Website
- Stellt iCal-Link zur Verfügung, womit der Kalender in mobilen Kalender-Apps abonniert werden kann

![Vorschau](art/preview.jpg)

Eine Live-Vorschau kann auf der Seite des [KV Fuldas](https://www.die-linke-fulda.de/termine/) eingesehen werden.

## 1. Eure URL zusammenstellen

Damit die korrekten Daten für euren KV/LV angezeit werden, müsste ihr eure eigene URL zusammenstellen.

### Zetkin

Dafür musst ihr zuerst eure **Zetkin Organisations-ID** rausfinden.

1. In Zetkin einloggen und das [Dashboard](https://www.zetkin.die-linke.de/dashboard/organizations) aufrufen.
2. Den gewünschten KV/LV anklicken. (hier z.B. "KV Fulda")
3. Die aufgerufene Browser-Adresse hat folgendes Schema: `https://www.zetkin.die-linke.de/o/192/`
    Hier endet die Adresse mit der Zahl `192` - also ist eure Zetkin Organisations-ID die `192`

### Die URL

Ersetzt nun den Platzhalter `<ORG>` mit der vorher ausgelesenen Organisations-ID:

```
https://linke-calendar.romanzipp.com/org/<ORG>/calendar
https://linke-calendar.romanzipp.com/org/<ORG>/list
```

## 2. Einbetten in CMS

Die verfügbaren Komponenten (Kalender & Termin-Liste) werden per `iframe` eingebettet. Selbst werden sie auf einem anderen Server bereitgestellt und nur im CMS "referenziert".

### Kalender

Zum Einbetten des Kalenders muss ein neuer Seitinhalt "Reines HTML" eingefügt werden.

![Anleitung](art/cms-01.png?v=2)

In das "HTML-Code" Feld wird nun der folgende Code eingefügt. Stelle sicher, dass die URL in `src="..."` der korrekt erstellte Link für euren KV ist.

![Anleitung](art/cms-03.png?v=2)

Die URL für den Kalender hat das folgende Format. Wie oben - ersetzt `<ORG>` mit eurer Organisations-ID.

```
https://linke-calendar.romanzipp.com/org/<ORG>/calendar
```

```html
<iframe id="calendar-embed"
        src="https://linke-calendar.romanzipp.com/org/<ORG>/calendar" 
        width="100%" 
        height="700" 
        scrolling="no"></iframe>

<style>
  #calendar-embed { 
    display: none; 
    margin-bottom: 2rem;
  }
  
  @media (min-width: 576px) {
    #calendar-embed { display: block; }
  }
</style>
```

---

### Termin-Liste

Zum Einfügen der Termin-Liste muss ein neuer Seitinhalt "Reines HTML" eingefügt werden.

![Anleitung](art/cms-01.png?v=2)

In das "HTML-Code" Feld wird nun der folgende Code eingefügt. Stelle sicher, dass die URL in `src="..."` der korrekt erstellte Link für euren KV ist.

![Anleitung](art/cms-02.png?v=2)

Die URL für den Kalender hat das folgende Format. Wie oben - ersetzt `<ORG>` mit eurer Organisations-ID.

```
https://linke-calendar.romanzipp.com/org/<ORG>/list?color=white
```

Es können zusätzliche URL Query-Parameter angehangen werden:

- `color`: Farbe des Texts (verfügbar: `black`, `white`)

```html
<header class="card-header">
  <div class="card-title">
    <h1 itemprop="headline">
      Termine
    </h1>
  </div>
</header>

<iframe src="https://linke-calendar.romanzipp.com/org/<ORG>/list?color=white"
        width="100%" 
        height="630"></iframe>

<div>
  <a href="/termine/" class="btn btn-primary mt-2">
    Kalender anzeigen
  </a>
</div>
```

## API Endpoints

- `GET /health` - Health check endpoint
- `GET /site/{siteID}/calendar` - Calendar view for a specific site
  - Query params: `year`, `month` (optional)
- `GET /site/{siteID}/list` - List view showing all upcoming events in chronological order
- `GET /site/{siteID}/ical` - iCal endpoint for subscribing with mobile device
- `GET /event/{eventID}` - Event detail modal
- `GET /static/*` - Static files (CSS, JS, fonts)

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

## License

MIT
