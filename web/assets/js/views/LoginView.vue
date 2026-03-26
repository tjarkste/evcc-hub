<template>
  <div class="login-container d-flex align-items-center justify-content-center min-vh-100">
    <div class="card p-4" style="width: 100%; max-width: 400px;">

      <!-- Login-Formular -->
      <div v-if="mode === 'login'">
        <h2 class="text-center mb-4">⚡ evcc Hub</h2>
        <p class="text-center text-muted small mb-4">
          Dein evcc-Dashboard, von überall erreichbar.<br>
          <a href="https://github.com/tjarkste/evcc-hub" target="_blank" rel="noopener" class="text-primary">
            Open Source auf GitHub →
          </a>
        </p>
        <div v-if="error" class="alert alert-danger">{{ error }}</div>
        <div class="mb-3">
          <input
            v-model="email"
            type="email"
            class="form-control"
            placeholder="E-Mail"
            data-test="email"
          />
        </div>
        <div class="mb-3">
          <input
            v-model="password"
            type="password"
            class="form-control"
            placeholder="Passwort"
            data-test="password"
          />
        </div>
        <button
          @click="handleLogin"
          class="btn btn-primary w-100"
          :disabled="loading"
          data-test="login-btn"
        >
          {{ loading ? 'Wird angemeldet...' : 'Anmelden' }}
        </button>
        <div class="text-center mt-3">
          <a href="#" @click.prevent="mode = 'register'">Noch kein Konto? Registrieren</a>
        </div>
      </div>

      <!-- Registrierungs-Formular -->
      <div v-else-if="mode === 'register'">
        <h2 class="text-center mb-4">⚡ evcc Hub</h2>
        <p class="text-center text-muted small mb-4">
          Dein evcc-Dashboard, von überall erreichbar.<br>
          <a href="https://github.com/tjarkste/evcc-hub" target="_blank" rel="noopener" class="text-primary">
            Open Source auf GitHub →
          </a>
        </p>
        <p class="text-muted text-center">Kostenlosen Account erstellen</p>
        <div v-if="error" class="alert alert-danger">{{ error }}</div>
        <div class="mb-3">
          <input
            v-model="email"
            type="email"
            class="form-control"
            placeholder="E-Mail"
            data-test="email"
          />
        </div>
        <div class="mb-3">
          <input
            v-model="password"
            type="password"
            class="form-control"
            placeholder="Passwort"
            data-test="password"
          />
        </div>
        <button
          @click="handleRegister"
          class="btn btn-primary w-100"
          :disabled="loading"
          data-test="register-btn"
        >
          {{ loading ? 'Wird registriert...' : 'Registrieren' }}
        </button>
        <div class="text-center mt-3">
          <a href="#" @click.prevent="mode = 'login'">Bereits ein Konto? Anmelden</a>
        </div>
      </div>

      <!-- Onboarding nach Registrierung -->
      <div v-else-if="mode === 'onboarding'">
        <h2 class="text-center mb-1">Konto erstellt!</h2>
        <p class="text-center text-muted mb-4">Verbinde jetzt deine evcc-Instanz.</p>

        <!-- Step 1 -->
        <div class="d-flex gap-3 mb-3">
          <div
            class="rounded-circle bg-success text-white d-flex align-items-center justify-content-center flex-shrink-0"
            style="width:2rem;height:2rem;font-size:0.85rem;"
          >1</div>
          <div>
            <p class="mb-0 fw-semibold">evcc installiert?</p>
            <p class="text-muted small mb-0">
              Falls noch nicht:
              <a href="https://docs.evcc.io/docs/installation/linux" target="_blank" rel="noopener">
                evcc installieren →
              </a>
            </p>
          </div>
        </div>

        <!-- Step 2 -->
        <div class="d-flex gap-3 mb-3">
          <div
            class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center flex-shrink-0"
            style="width:2rem;height:2rem;font-size:0.85rem;"
          >2</div>
          <div class="w-100">
            <p class="mb-1 fw-semibold">MQTT-Konfiguration hinzufügen</p>
            <p class="text-muted small mb-2">
              Füge diese Zeilen in deine <code>evcc.yaml</code> ein:
            </p>
            <pre
              class="bg-dark text-light p-2 rounded mb-2"
              style="font-size: 0.8em;"
              data-test="mqtt-config"
            >{{ mqttConfig }}</pre>
            <button
              @click="copyConfig"
              class="btn btn-outline-secondary btn-sm"
              data-test="copy-config-btn"
            >
              {{ copied ? '✓ Kopiert!' : 'Kopieren' }}
            </button>
          </div>
        </div>

        <!-- Step 3 -->
        <div class="d-flex gap-3 mb-4">
          <div
            class="rounded-circle bg-secondary text-white d-flex align-items-center justify-content-center flex-shrink-0"
            style="width:2rem;height:2rem;font-size:0.85rem;"
          >3</div>
          <div>
            <p class="mb-0 fw-semibold">evcc neu starten</p>
            <p class="text-muted small mb-0">
              <code>sudo systemctl restart evcc</code>
            </p>
          </div>
        </div>

        <button
          @click="goToDashboard"
          class="btn btn-primary w-100"
          data-test="to-dashboard-btn"
        >
          Weiter zum Dashboard
        </button>
      </div>

      <!-- Footer -->
      <div class="text-center mt-4 pt-3 border-top">
        <small class="text-muted">
          <router-link to="/impressum">Impressum</router-link> ·
          <router-link to="/datenschutz">Datenschutz</router-link> ·
          <router-link to="/nutzungsbedingungen">Nutzungsbedingungen</router-link> ·
          <a href="https://github.com/tjarkste/evcc-hub" target="_blank" rel="noopener">
            <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" style="vertical-align:middle; margin-right:2px;"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
            GitHub
          </a>
        </small>
      </div>

    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { login, register } from '../services/auth'
import type { Site } from '../services/sites'

export default defineComponent({
  name: 'LoginView',
  data() {
    return {
      mode: 'login' as 'login' | 'register' | 'onboarding',
      email: '',
      password: '',
      loading: false,
      error: '',
      mqttConfig: '',
      copied: false,
      siteCredentials: null as { mqttUsername: string; mqttPassword: string; topicPrefix: string } | null,
    }
  },
  methods: {
    async handleLogin() {
      this.loading = true
      this.error = ''
      try {
        await login(this.email, this.password)
        this.$router.push('/')
      } catch {
        this.error = 'Anmeldung fehlgeschlagen. Bitte E-Mail und Passwort prüfen.'
      } finally {
        this.loading = false
      }
    },
    async handleRegister() {
      this.loading = true
      this.error = ''
      try {
        const auth = await register(this.email, this.password)
        const site = auth.defaultSite
        if (site) {
          this.mqttConfig = `mqtt:
  broker: tls://mqtt.evcc-hub.de:8883
  topic: ${site.topicPrefix}
  user: ${site.mqttUsername}
  password: "${site.mqttPassword}"`
        }
        this.mode = 'onboarding'
      } catch {
        this.error = 'Registrierung fehlgeschlagen. Bitte versuche es erneut.'
      } finally {
        this.loading = false
      }
    },
    async copyConfig() {
      await navigator.clipboard.writeText(this.mqttConfig)
      this.copied = true
      setTimeout(() => { this.copied = false }, 2000)
    },
    goToDashboard() {
      this.$router.push('/')
    },
  },
})
</script>
