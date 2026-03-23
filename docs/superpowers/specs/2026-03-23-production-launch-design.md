# Production Launch Design
**Datum:** 2026-03-23
**Status:** Approved

## Ziel

evcc Cloud in Produktion bringen — rechtlich sauber, technisch stabil, mit einer klaren Strategie für die ersten Nutzer aus der evcc-Community.

---

## Rahmenbedingungen

- **Betreiber:** Einzelperson (kein Unternehmen)
- **Monetarisierung:** Zunächst kostenlos, später optionales Abo oder Donations
- **Launch-Typ:** Soft Launch (öffentlich erreichbar, keine große Ankündigung)
- **Infrastruktur:** Hetzner VPS (Deutschland), selbst verwaltet
- **Zielgruppe:** evcc-Nutzer (deutschsprachig, technikaffin, Community-getrieben)

---

## Launch-Voraussetzungen (Blocker — alles davon muss vor Go-Live erfüllt sein)

| # | Punkt | Hinweis |
|---|---|---|
| 1 | Domain registriert und auf Hetzner-IP zeigend | Pflicht für SSL + Impressum |
| 2 | Betreiber-Name und Anschrift für Impressum festgelegt | Pflicht § 5 TMG — auch beim Soft Launch |
| 3 | Impressum + Datenschutzerklärung final und geprüft | Muss die tatsächlichen Datenflüsse abbilden |
| 4 | Nutzungsbedingungen online | Schutz für den Betreiber |
| 5 | Datenlöschprozess definiert | DSGVO Art. 17 — mindestens per E-Mail-Anfrage |

> **Wichtig:** Ein Soft Launch bedeutet "öffentlich erreichbar ohne Ankündigung" — nicht "ohne Impressumspflicht". Sobald die App unter einer Domain erreichbar ist, gelten alle Pflichten.

---

## Abschnitt 1: Infrastruktur

### Server

| Komponente | Entscheidung |
|---|---|
| Provider | Hetzner Cloud |
| Instanz | CX22 (2 vCPU, 4 GB RAM, 40 GB SSD, ~4,15€/Monat) |
| Standort | Falkenstein oder Nürnberg (Deutschland, EU) |
| OS | Ubuntu 24.04 LTS |

### Stack

Das bestehende `docker-compose.yml` (Mosquitto + Go-Backend + Nginx) bleibt die Basis. Ergänzungen:

- **SSL:** Let's Encrypt via Certbot, automatische Verlängerung per Cron
- **Firewall:** UFW, nur Port 22 (SSH), 80 (HTTP→Redirect), 443 (HTTPS) offen
- **Backups:** Täglicher Cron-Job, SQLite-Datenbankdatei als tar.gz auf Hetzner Object Storage (EU)
  - Retention: 7 Tages-Backups, 4 Wochen-Backups
  - Restore-Test: Einmal vor Launch manuell durchführen und dokumentieren

### Authentifizierung (bestehend)

Das Backend nutzt bereits JWT-basierte Authentifizierung mit E-Mail + Passwort (golang-jwt). Im Scope des Launches:
- Passwörter werden gehasht gespeichert (bestehend)
- Session-Expiry via JWT-Ablauf (bestehend)
- Kein Self-Service Passwort-Reset zum Launch — Nutzer können per E-Mail anfragen (akzeptabel für kleine Nutzerzahlen)

### Monitoring

| Tool | Zweck | Datenschutz | Kosten |
|---|---|---|---|
| UptimeRobot EU | Uptime-Check alle 5 Min, Email-Alert | EU-Server, AV-Vertrag vorhanden | kostenlos |
| Sentry (EU-Region, Frankfurt) | Backend-Fehler + Frontend-Crashes | EU-Datenhaltung, PII-Scrubbing aktivieren | kostenlos |

> **Hinweis Sentry:** In der Sentry-Konfiguration muss PII-Scrubbing aktiviert sein, sodass keine User-IDs, E-Mail-Adressen oder Request-Pfade mit persönlichen Daten in die Fehler-Logs gelangen. Sentry bietet dies als eingebaute Option an.

> **Hinweis UptimeRobot:** Ausschließlich UptimeRobot EU (eu.uptimerobot.com) verwenden — nicht die US-Instanz. Damit entfällt das Problem des Drittlandtransfers.

---

## Abschnitt 2: Legal (DSGVO / Deutsches Recht)

Alle Dokumente müssen **vor dem öffentlichen Zugang** online und über den Footer auf jeder Seite erreichbar sein.

### Dokument 1: Impressum (§ 5 TMG)

- Vollständiger Name, Anschrift, E-Mail des Betreibers
- Erstellung via Generator (e-recht24.de) — anschließend manuell auf Vollständigkeit prüfen
- **Launch-Blocker:** Muss mit echter Adresse befüllt sein

### Dokument 2: Datenschutzerklärung (DSGVO Art. 13)

Muss die tatsächlichen Datenflüsse des Projekts abbilden — nicht nur den Generator-Standard:

| Datenkategorie | Zweck | Rechtsgrundlage (Art. 6 DSGVO) |
|---|---|---|
| E-Mail-Adresse | Kontoanmeldung, Kommunikation | Art. 6 (1) b — Vertragserfüllung |
| Passwort (gehasht) | Authentifizierung | Art. 6 (1) b — Vertragserfüllung |
| evcc-Gerätedaten (Ladedaten, Energiedaten) | Kernfunktion des Dienstes | Art. 6 (1) b — Vertragserfüllung |
| Fehler-Logs via Sentry | Betrieb und Fehlerbehebung | Art. 6 (1) f — berechtigtes Interesse |
| Uptime-Monitoring via UptimeRobot EU | Betrieb | Art. 6 (1) f — berechtigtes Interesse |

- Speicherort: Hetzner Deutschland (EU — kein Drittlandtransfer)
- Drittdienste als Auftragsverarbeiter nennen: Sentry (EU/Frankfurt), UptimeRobot EU
- Erstellung via Generator (e-recht24.de), danach mit obiger Tabelle abgleichen

### Dokument 3: Nutzungsbedingungen

Schutzklauseln für den Betreiber:
- Dienst wird kostenlos und ohne Garantie bereitgestellt
- Kein SLA, kein Support-Versprechen
- Betreiber kann Dienst jederzeit einstellen oder ändern
- Nutzer ist selbst für seine Daten verantwortlich

### Datenlöschung und Datenportabilität (DSGVO Art. 17 + 20)

Minimaler Prozess — ausreichend für kleine Nutzerzahlen:
- Nutzer können per E-Mail die Löschung ihres Kontos und aller Daten beantragen
- Betreiber führt Löschung manuell in SQLite durch und bestätigt per E-Mail
- E-Mail-Adresse hierfür im Impressum und in der Datenschutzerklärung angeben
- Datenexport auf Anfrage als JSON — kein Self-Service zum Launch erforderlich

### Cookie / Consent

- Sentry (JavaScript SDK) und UptimeRobot setzen ggf. keine First-Party-Cookies, aber der Sentry-SDK macht Netzwerkanfragen
- Im Datenschutz-Footer-Link kurze Erwähnung ausreichend, kein Cookie-Banner erforderlich wenn keine Tracking-Cookies gesetzt werden
- Prüfen ob Sentry-SDK Cookies setzt — falls ja, einfacher Hinweis-Banner ohne Consent-Gate ausreichend (kein Tracking-Zweck)

### Donations / Ko-fi

Ko-fi oder GitHub Sponsors werden **erst nach dem Community-Launch** und **nur als optionaler Hinweis** eingebaut — nicht zum Soft Launch. Solange kein wiederkehrender Zahlungsfluss entsteht, gelten keine Fernabsatz-Pflichten.

### Was nicht nötig ist

- Kein formeller AGB-Komplex im e-commerce-Sinne
- Kein Steuerberater (solange keine Einnahmen)
- Keine GmbH oder andere Rechtsform

**Geschätzter Zeitaufwand:** 3–4 Stunden (Generator + manuelle Prüfung gegen Datentabelle oben).

---

## Abschnitt 3: Nutzer-Akquise

### Phase 1 — Soft Launch (Woche 1–4)

Keine aktive Kommunikation. System stabilisieren, erste organische Nutzer beobachten, Bugs beheben. Alle rechtlichen Dokumente sind bereits online.

### Phase 2 — Community Launch (Woche 5+, Trigger: System stabil, keine kritischen Bugs)

Ein einziger, authentischer Post im **evcc Community Forum** (community.evcc.io):
- Was die App ist und warum sie gebaut wurde
- Screenshot oder kurze Demo
- Link + Hinweis "kostenlos, solange ich es betreibe"
- Kein Marketing-Sprech — evcc-Nutzer reagieren auf Ehrlichkeit

### Weitere Kanäle (nach erstem Community-Feedback)

| Kanal | Trigger | Aufwand |
|---|---|---|
| evcc GitHub Discussions | Nach stabilem Community-Feedback | gering |
| Reddit (r/evcc, r/homeautomation) | Nach Community-Post | gering |
| YouTube Demo-Video (~3 Min) | Nach ~50 aktiven Nutzern | mittel |

### Was nicht gemacht wird

- Keine bezahlten Ads
- Keine Kaltakquise
- Keine unrealistischen Versprechen (Roadmap-Garantien, SLA)

### Monetarisierung (nach Community-Launch, nicht zum Soft Launch)

1. "Support this project"-Link mit **Ko-fi** oder **GitHub Sponsors** im Footer — freiwillig
2. Nach ~6 Monaten und ~50+ aktiven Nutzern: Community-Umfrage zu optionalem Abo (2–3€/Monat)

---

## Offene Punkte (vor Launch zu klären)

| Punkt | Status |
|---|---|
| Domain-Name | Offen — Ideen: `evcc-hub.de`, `solar-cloud.app`, `evccdash.de` |
| Domain-Registrar | Offen — empfohlen: INWX (DE) oder Hetzner |
| Betreiber-Name + Anschrift für Impressum | Zu ergänzen |
| Restore-Test SQLite-Backup dokumentieren | Vor Launch durchführen |

---

## Nicht im Scope (bewusste YAGNI-Entscheidungen)

- Kubernetes / automatisches Scaling
- Vollständiges AGB-Werk mit Anwalt
- Stripe-Integration / Billing-System
- Analytics-Platform
- Mehrsprachige Legal-Dokumente
- Self-Service Passwort-Reset (zum Launch)
- Formelles Auftragsverarbeitungsverzeichnis (erst ab 250 Mitarbeiter verpflichtend, empfohlen aber nicht Blocker)
