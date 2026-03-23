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
- **Backups:** Täglicher Cron-Job sichert die SQLite-Datenbankdatei (tar.gz + rsync auf Hetzner Object Storage oder zweiten Ort)

### Monitoring

| Tool | Zweck | Kosten |
|---|---|---|
| UptimeRobot (Free) | Uptime-Check alle 5 Min, Email-Alert bei Ausfall | kostenlos |
| Sentry (Free Tier) | Backend-Fehler + Frontend-Crashes | kostenlos |

---

## Abschnitt 2: Legal (DSGVO / Deutsches Recht)

Alle drei Dokumente müssen **vor dem öffentlichen Zugang** online und über den Footer erreichbar sein.

### Dokument 1: Impressum (§ 5 TMG)

- Vollständiger Name, Anschrift, E-Mail des Betreibers
- Erstellung via Generator (e-recht24.de)
- Kein Anwalt erforderlich

### Dokument 2: Datenschutzerklärung (DSGVO Art. 13)

Mindestinhalt:
- Welche Daten gespeichert werden: E-Mail, Passwort (gehashed), evcc-Gerätedaten
- Speicherort: Hetzner Deutschland (EU — kein Drittlandtransfer)
- Keine Weitergabe an Dritte
- Drittdienste mit Auftragsverarbeitung explizit nennen: **Sentry** (Fehler-Logging), **UptimeRobot** (Monitoring)
- Erstellung via Generator (e-recht24.de oder datenschutz.org)

### Dokument 3: Nutzungsbedingungen

Schutzklauseln für den Betreiber:
- Dienst wird kostenlos und ohne Garantie bereitgestellt
- Kein SLA, kein Support-Versprechen
- Betreiber kann Dienst jederzeit einstellen oder ändern
- Nutzer ist selbst für seine Daten verantwortlich

### Was nicht nötig ist (solange kein Geld fließt)

- Kein Widerrufsrecht / AGB im e-commerce-Sinne
- Kein Steuerberater
- Keine GmbH oder andere Rechtsform

**Geschätzter Zeitaufwand:** 2–3 Stunden mit Generatoren.

---

## Abschnitt 3: Nutzer-Akquise

### Phase 1 — Soft Launch (Woche 1–4)

Keine aktive Kommunikation. System stabilisieren, erste organische Nutzer beobachten, Bugs beheben. Wer via GitHub oder direktem Link draufstößt, ist willkommen.

### Phase 2 — Community Launch (Woche 5+)

Ein einziger, authentischer Post im **evcc Community Forum** (community.evcc.io):
- Was die App ist und warum sie gebaut wurde
- Screenshot oder kurze Demo
- Link + Hinweis "kostenlos, solange ich es betreibe"
- Kein Marketing-Sprech — evcc-Nutzer reagieren auf Ehrlichkeit

### Weitere Kanäle (nach erstem Community-Feedback)

| Kanal | Zeitpunkt | Aufwand |
|---|---|---|
| evcc GitHub Discussions | Nach erstem stabilem Feedback | gering |
| Reddit (r/evcc, r/homeautomation) | Nach Community-Post | gering |
| YouTube Demo-Video (~3 Min) | Nach ~50 aktiven Nutzern | mittel |

### Was nicht gemacht wird

- Keine bezahlten Ads
- Keine Kaltakquise
- Keine unrealistischen Versprechen (Roadmap-Garantien, SLA)

### Monetarisierung vorbereiten (noch nicht aktivieren)

- "Support this project"-Link mit **Ko-fi** oder **GitHub Sponsors** im Footer einbauen — freiwillig, kein Druck
- Nach ~6 Monaten und ~50+ aktiven Nutzern: Community-Umfrage ob optionales Abo (2–3€/Monat für erweiterte Features) gewünscht

---

## Offene Punkte

| Punkt | Status |
|---|---|
| Domain-Name | Offen — Ideen: `evcc-hub.de`, `solar-cloud.app`, `evccdash.de` |
| Domain-Registrar | Offen — empfohlen: Hetzner, INWX, oder Namecheap |
| Betreiber-Adresse für Impressum | Zu ergänzen |

---

## Nicht im Scope (bewusste YAGNI-Entscheidungen)

- Kubernetes / automatisches Scaling
- Vollständiges AGB-Werk mit Anwalt
- Stripe-Integration / Billing-System
- Analytics-Platform
- Mehrsprachige Legal-Dokumente
