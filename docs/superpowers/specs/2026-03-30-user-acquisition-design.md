# User Acquisition Design
**Datum:** 2026-03-30
**Status:** Approved

## Ziel

50–100 aktive Nutzer innerhalb von 3 Monaten nach Launch gewinnen — als unbekannter Solo-Dev, ohne Budget für Ads, durch authentische Community-Kommunikation.

---

## Rahmenbedingungen

- **Ausgangslage:** 0 Nutzer, kein Bekanntheitsgrad in der evcc-Community
- **Betreiber:** Solo-Dev, verfügbare Zeit: Abende und Wochenenden
- **Monetarisierung:** Kostenlos zum Launch, "Buy me a coffee"-Button bei ~50 Nutzern, Abo-Modell langfristig
- **Zielgruppe:** evcc-Nutzer (deutschsprachig, technikaffin, Community-getrieben)
- **Kern-Value-Prop:** evcc von überall erreichbar — ohne VPN, ohne Portfreigabe, kostenlos

---

## Gewählter Ansatz

**Authentic Cold Launch + Demo Artifact**

Direkter Launch-Post in der evcc-Community mit einem überzeugenden 60-Sekunden-Demo-Video. Kein Aufbau von Community-Credibility im Voraus (zu langsam). Kein Multi-Channel-Blast am ersten Tag (zu aufwändig). Stattdessen: ein starkes Demo-Artefakt, das das Selling übernimmt, kombiniert mit authentischer Kommunikation.

---

## Abschnitt 1: Positionierung & Kernbotschaft

Die eine Aussage, die alles antreibt:

> **"evcc von überall — ohne VPN, ohne Portfreigabe, kostenlos."**

Jeder Post, jeder Kanal, jede Demo beginnt damit. Nicht mit Technologie (MQTT, TLS, Docker) — mit dem Ergebnis. Der evcc-Nutzer, der schon mal mit Tailscale oder Portfreigabe gekämpft hat, erkennt sich sofort.

Sekundäre Botschaft: **Authentizität als Asset.** Kein Unternehmen, kein SaaS-Produkt — ein Solo-Dev, der dieses Problem selbst hatte und gelöst hat. In dieser Community ist das ein Vorteil, kein Nachteil.

---

## Abschnitt 2: Das Demo-Artefakt

Das wichtigste Stück Content vor dem ersten Post.

**Was aufnehmen (~60 Sekunden):**
1. evcc läuft lokal — Dashboard sichtbar im Browser
2. `evcc.yaml` öffnen, MQTT-Config eintragen (4 Zeilen)
3. `sudo systemctl restart evcc`
4. Browser auf Mobilgröße verkleinern (oder Handy zeigen) — evcc-hub.de öffnen — Dashboard lädt live

**Format:** MP4 + animiertes GIF. GIF bettet sich inline in Forum-Posts und GitHub ein. MP4 läuft nativ auf Reddit und in den meisten Browsern.

**Was vermeiden:** Keine Musik, keine Intro-Animation, keine Untertitel. Eine saubere Bildschirmaufnahme ohne Ablenkungen ist glaubwürdiger als ein poliertes Marketing-Video.

**Aufwand:** ~1 Abend

---

## Abschnitt 3: Channel-Sequenz & Launch-Woche

Posts gestaffelt über 3–5 Tage — nicht alles auf einmal. Den Forum-Thread erst atmen lassen und Antworten sammeln.

### Tag 1 — evcc Community Forum (community.evcc.io)
Der wichtigste Kanal. Genau die richtige Zielgruppe.

**Sprache:** Deutsch
**Ton:** Kurz, direkt, kein Marketing-Sprech

**Post-Struktur:**
1. Ein-Zeilen-Hook (der Schmerz zuerst): *"evcc von unterwegs? Ich hab's so gelöst:"*
2. Demo-GIF eingebettet
3. 3 Bullet Points: Was es ist, wie es funktioniert, dass es kostenlos ist
4. Call to Action: *"Beta-Tester gesucht — einfach registrieren und Feedback geben"*

Nicht mit "Ich hab das gebaut" anfangen — mit der Demo führen, dann erklären.

### Tag 3 — evcc GitHub Discussions
Kürzerer Post, Verweis auf den Forum-Thread. GitHub-Publikum ist etwas technischer und selbst-hoster-affiner — Open-Source-Aspekt und Self-Hosting-Option hier explizit erwähnen.

### Tag 5 — Reddit (r/selfhosted + r/homeautomation)
**Sprache:** Englisch, gleiche Struktur. r/selfhosted hat eine große Zielgruppe, die "kein VPN, keine Portfreigabe" schätzt. Link auf die Website, nicht auf den Forum-Thread.

---

## Abschnitt 4: Von ersten Nutzern zu weiteren Nutzern

### Persönliche E-Mail an jeden Early Signup
Für die ersten 30–40 Nutzer: kurze persönliche E-Mail nach der Registrierung. Keine Vorlage — ein Satz: *"Hey, danke fürs Ausprobieren — funktioniert alles?"*

Konvertiert stille Nutzer in Feedback-Geber und organische Fürsprecher im Forum.

### Forum-Thread aktiv halten
Alle 2–3 Wochen ein Update im ursprünglichen Thread: *"Update: X Nutzer aktiv, diese Woche folgende Verbesserungen gemacht."*

Hält den Thread sichtbar und signalisiert, dass das Projekt lebt. Aktive Pflege ist in dieser Community das stärkste Vertrauenssignal.

---

## Abschnitt 5: Timeline & Erfolgsmessung

### 3-Monats-Roadmap

| Woche | Aktion |
|-------|--------|
| 1 | Demo aufnehmen, Forum-Post schreiben, auf community.evcc.io launchen |
| 2 | Post in GitHub Discussions, persönliche E-Mail an jeden Signup |
| 3 | Posts auf Reddit (r/selfhosted + r/homeautomation) |
| 4–6 | Monitoring, Bugs fixen, jeden Forum-Kommentar beantworten |
| 6 | Erstes Forum-Update ("X Nutzer, neue Features") |
| 8 | Zweites Forum-Update |
| 10–12 | Bewertung: Bei 50+ Nutzern → "Buy me a coffee"-Link hinzufügen |

### Erfolgsmessung

| Zeitpunkt | Ziel |
|-----------|------|
| Ende Woche 1 | 10 Signups aus dem Forum-Post |
| Ende Monat 1 | 25 aktive Nutzer (mind. 1 verbundener Standort) |
| Ende Monat 3 | 50–100 aktive Nutzer, 3–5 Nutzer mit direktem Feedback |

**Definition "aktiv":** Nutzer hat mind. einen Standort verbunden und das Dashboard in den letzten 30 Tagen aufgerufen.

---

## Was nicht gemacht wird (YAGNI)

- Keine bezahlten Ads
- Keine Kaltakquise
- Kein wochenlanger Community-Credibility-Aufbau vor dem Launch
- Kein gleichzeitiger Multi-Channel-Blast am Tag 1
- Kein In-Produkt-Sharing-Prompt (zu früh, zu viel Aufwand)
- Keine unrealistischen Versprechen (Roadmap-Garantien, SLA)
- Kein "Buy me a coffee"-Button zum Launch

---

## Offene Punkte

| Punkt | Status |
|-------|--------|
| Demo-Video aufnehmen | Vor Tag-1-Post erledigen |
| Forum-Post-Text auf Deutsch ausformulieren | Vor Tag-1-Post erledigen |
| Persönliche E-Mail-Vorlage als Ausgangspunkt erstellen | Vor Woche-2 erledigen |
