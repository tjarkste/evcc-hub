<template>
  <div class="login-container d-flex align-items-center justify-content-center min-vh-100">
    <div class="card p-4" style="width: 100%; max-width: 400px;">

      <!-- Login-Formular -->
      <div v-if="mode === 'login'">
        <h2 class="text-center mb-4">☀ evcc Cloud Connect</h2>
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
        <h2 class="text-center mb-4">☀ evcc Cloud Connect</h2>
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
