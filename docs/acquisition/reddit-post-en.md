**Published:**
- r/selfhosted: [add URL after posting]
- r/homeautomation: [add URL after posting]

---

**Title:** I built a free cloud dashboard for evcc (EV charger controller) — no VPN, no port forwarding needed

**Body:**

Hey r/selfhosted — I built this and wanted to share.

[DEMO VIDEO/GIF]

**The problem:** evcc is a great open-source EV charging controller, but accessing it
remotely meant setting up VPN or opening ports — annoying for a home server setup.

**What I built:** evcc hub — a free cloud dashboard that syncs with your local evcc
instance via MQTT over TLS. Your evcc connects outbound, so no inbound ports needed.

**Features:**
- Remote access to your evcc dashboard from anywhere
- Multi-site support (multiple locations in one account)
- Real-time data via MQTT with TLS encryption
- Open source and self-hostable (MIT)
- Free in beta

**Setup:** Register → get 4 lines of MQTT config → paste into evcc.yaml → restart evcc. Done.

→ **evcc-hub.de** | GitHub: **https://github.com/tjarksteenblock/evcc_hub**

I'm the developer — happy to answer questions. Looking for beta testers and feedback.
