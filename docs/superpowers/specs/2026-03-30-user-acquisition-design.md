# User Acquisition Design
**Datum:** 2026-03-30
**Status:** Approved

## Ziel

50–100 aktive Nutzer innerhalb von 3 Monaten nach Launch gewinnen — als unbekannter Solo-Dev, ohne Budget für Ads, durch authentische Community-Kommunikation.

---

## Rahmenbedingungen

- **Ausgangslage:** 0 Nutzer, kein Bekanntheitsgrad in der evcc-Community
- **Betreiber:** Solo-Dev, verfügbare Zeit: Abende und Wochenenden
- **Monetarisierung:** Kostenlos in der Beta-Phase, "Buy me a coffee"-Button bei ~50 Nutzern, optionales Abo-Modell langfristig. Die "kostenlos"-Botschaft gilt für die Beta-Phase; die Kommunikation eines zukünftigen Preismodells ist ein separates Thema und wird nicht zum Launch versprochen.
- **Zielgruppe:** evcc-Nutzer (deutschsprachig, technikaffin, Community-getrieben)
- **Kern-Value-Prop:** evcc von überall erreichbar — ohne VPN, ohne Portfreigabe, kostenlos (Beta)

---

## Gewählter Ansatz

**Authentic Cold Launch + Demo Artifact**

Direkter Launch-Post in der evcc-Community mit einem überzeugenden 60-Sekunden-Demo-Video. Kein Aufbau von Community-Credibility im Voraus (zu langsam). Kein Multi-Channel-Blast am ersten Tag (zu aufwändig). Stattdessen: ein starkes Demo-Artefakt, das das Selling übernimmt, kombiniert mit authentischer Kommunikation.

---

## Abschnitt 1: Positionierung & Kernbotschaft

Die eine Aussage, die alles antreibt:

> **"evcc von überall — ohne VPN, ohne Portfreigabe, kostenlos (Beta)."**

Jeder Post, jeder Kanal, jede Demo beginnt damit. Nicht mit Technologie (MQTT, TLS, Docker) — mit dem Ergebnis. Der evcc-Nutzer, der schon mal mit Tailscale oder Portfreigabe gekämpft hat, erkennt sich sofort.

Sekundäre Botschaft: **Authentizität als Asset.** Kein Unternehmen, kein SaaS-Produkt — ein Solo-Dev, der dieses Problem selbst hatte und gelöst hat. In dieser Community ist das ein Vorteil, kein Nachteil.

---

## Abschnitt 2: Das Demo-Artefakt

Das wichtigste Stück Content vor dem ersten Post.

**Voraussetzung (Blocker):** evcc-hub.de ist live und erreichbar, und die MQTT-Setup-Dokumentation ist auf der Website verfügbar — bevor das Demo aufgenommen oder gepostet wird. Ein Demo-Video, das auf eine nicht erreichbare URL zeigt, zerstört die Glaubwürdigkeit sofort.

**Was aufnehmen (~60 Sekunden):**
1. evcc läuft lokal — Dashboard sichtbar im Browser
2. `evcc.yaml` öffnen, MQTT-Config eintragen (4 Zeilen — die exakten Werte aus der Website-Dokumentation verwenden)
3. `sudo systemctl restart evcc`
4. Browser auf Mobilgröße verkleinern (oder Handy zeigen) — evcc-hub.de öffnen — Dashboard lädt live

**Format:** MP4 + animiertes GIF. GIF bettet sich inline in Forum-Posts und GitHub ein. MP4 läuft nativ auf Reddit und in den meisten Browsern.

**Was vermeiden:** Keine Musik, keine Intro-Animation, keine Untertitel. Eine saubere Bildschirmaufnahme ohne Ablenkungen ist glaubwürdiger als ein poliertes Marketing-Video.

**Aufwand:** ~1 Abend

---

## Abschnitt 3: Channel-Sequenz & Launch-Woche

Posts gestaffelt über eine Woche — nicht alles auf einmal. Den Forum-Thread erst atmen lassen und Antworten sammeln.

### Tag 1 — evcc Community Forum (community.evcc.io)
Der wichtigste Kanal. Genau die richtige Zielgruppe.

**Sprache:** Deutsch
**Ton:** Kurz, direkt, kein Marketing-Sprech

**Post-Struktur:**
1. Ein-Zeilen-Hook (der Schmerz zuerst): *"evcc von unterwegs? Ich hab's so gelöst:"*
2. Demo-GIF eingebettet
3. 3 Bullet Points: Was es ist, wie es funktioniert, dass es kostenlos (Beta) ist
4. Call to Action: *"Beta-Tester gesucht — einfach registrieren und Feedback geben"*

Nicht mit "Ich hab das gebaut" anfangen — mit der Demo führen, dann erklären.

### Tag 3 — evcc GitHub Discussions
**Sprache:** Deutsch oder Englisch (je nach aktivem Sprachgebrauch in den Discussions)
**Ton:** Technischer als Forum-Post, kürzer

**Post-Struktur:**
1. Hook: *"Remote access to evcc without VPN — I built this:"*
2. Link zum Demo-Video
3. Bullet Points: Open Source, self-hostable, free in beta
4. Call to Action: *"Looking for beta testers — feedback welcome"*
5. Verweis auf den Forum-Thread für ausführlichere Diskussion

Der genaue Post-Text wird als eigener Offene-Punkte-Eintrag vor Tag 3 ausformuliert.

### Tag 5 — Reddit (r/selfhosted + r/homeautomation)
**Sprache:** Englisch
**Ton:** Direkt, mit Disclosure-Zeile ("I built this")

**Vor dem Post (Pflicht):** Subreddit-Regeln für r/selfhosted und r/homeautomation prüfen — insbesondere Regeln zu Self-Promotion, required Flair und Posting-Frequenz. Ohne diese Prüfung riskiert ein Post die Entfernung durch Moderatoren.

**Post-Struktur:**
1. Titel: *"I built a free cloud dashboard for evcc (EV charger controller) — no VPN, no port forwarding needed"*
2. Demo-Video eingebettet oder verlinkt
3. Kurze Erklärung: Was evcc ist, welches Problem gelöst wird, wie es funktioniert
4. Disclosure: *"I'm the developer — happy to answer questions"*
5. Link auf evcc-hub.de

Bei Entfernung durch Moderatoren: kurz Bescheid geben und alternative Subreddits prüfen (z.B. r/evcharging, r/electricvehicles).

---

## Abschnitt 4: Von ersten Nutzern zu weiteren Nutzern

### Persönliche E-Mail an jeden Early Signup
Für die ersten 30–40 Nutzer: kurze persönliche E-Mail nach der Registrierung. Keine Vorlage — ein Satz: *"Hey, danke fürs Ausprobieren — funktioniert alles?"*

**Trigger:** Täglich manuell neue Signups im Backend prüfen (oder automatische E-Mail-Benachrichtigung bei neuem Nutzer-Account einrichten). Ziel: E-Mail innerhalb von 24 Stunden nach Signup verschicken.

Konvertiert stille Nutzer in Feedback-Geber und organische Fürsprecher im Forum.

### Forum-Thread aktiv halten
Alle 2–3 Wochen ein Update im ursprünglichen Thread: *"Update: X Nutzer aktiv, diese Woche folgende Verbesserungen gemacht."*

Hält den Thread sichtbar und signalisiert, dass das Projekt lebt. Aktive Pflege ist in dieser Community das stärkste Vertrauenssignal.

---

## Abschnitt 5: Timeline & Erfolgsmessung

### 3-Monats-Roadmap

| Zeitraum | Aktion |
|----------|--------|
| Woche 1, Tag 1 | Demo aufnehmen, Forum-Post schreiben, auf community.evcc.io launchen |
| Woche 1, Tag 3 | Post in evcc GitHub Discussions |
| Woche 1, Tag 5 | Posts auf Reddit (r/selfhosted + r/homeautomation) |
| Woche 1–3 | Jeden Forum-Kommentar beantworten, persönliche E-Mail an jeden Signup |
| Woche 3 | Erstes Forum-Update ("X Nutzer, neue Features") |
| Woche 5 | Zweites Forum-Update |
| Woche 6–8 | Monitoring, Bugs fixen, Feedback einarbeiten |
| Woche 10–12 | Bewertung: Bei 50+ Nutzern → "Buy me a coffee"-Link hinzufügen |

### Erfolgsmessung

| Zeitpunkt | Ziel | Anmerkung |
|-----------|------|-----------|
| Ende Woche 1 | 10 Signups | Grobe Annahme: Forum hat mehrere hundert aktive Mitglieder; 1–2% Conversion bei passendem Problem ist realistisch, aber nicht garantiert |
| Ende Monat 1 | 25 aktive Nutzer | Aktiv = mind. 1 verbundener Standort |
| Ende Monat 3 | 50–100 aktive Nutzer, 3–5 Nutzer mit direktem Feedback | |

**Definition "aktiv":** Nutzer hat mind. einen Standort verbunden und das Dashboard in den letzten 30 Tagen aufgerufen.

**Tracking-Voraussetzung:** Das Backend muss Dashboard-Aufrufe pro Nutzer und verbundene Standorte auswertbar erfassen. Vor Launch prüfen, ob diese Daten abfragbar sind — andernfalls als Offener Punkt ergänzen.

---

## Was nicht gemacht wird (YAGNI)

- Keine bezahlten Ads
- Keine Kaltakquise
- Kein wochenlanger Community-Credibility-Aufbau vor dem Launch
- Kein gleichzeitiger Multi-Channel-Blast am Tag 1
- Kein In-Produkt-Sharing-Prompt (zu früh, zu viel Aufwand)
- Keine unrealistischen Versprechen (Roadmap-Garantien, SLA, "kostenlos für immer")
- Kein "Buy me a coffee"-Button zum Launch

---

## Offene Punkte

| Punkt | Wann erledigen |
|-------|---------------|
| Website evcc-hub.de live-Prüfung — erreichbar und MQTT-Doku vorhanden | Vor Demo-Aufnahme |
| Demo-Video aufnehmen (MP4 + GIF) | Vor Tag-1-Post |
| Forum-Post auf Deutsch ausformulieren | Vor Tag-1-Post |
| Subreddit-Regeln für r/selfhosted + r/homeautomation prüfen | Vor Tag-5-Post |
| GitHub-Discussions-Post ausformulieren | Vor Tag-3-Post |
| Persönliche E-Mail-Benachrichtigung bei neuem Signup einrichten (oder manuellen Prozess festlegen) | Vor Woche 2 |
| Aktivitäts-Tracking verifizieren: Sind Dashboard-Aufrufe und verbundene Standorte pro Nutzer abfragbar? | Vor Launch |
