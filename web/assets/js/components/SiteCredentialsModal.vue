<template>
	<div
		id="siteCredentialsModal"
		class="modal fade"
		tabindex="-1"
		aria-labelledby="siteCredentialsModalLabel"
		aria-hidden="true"
	>
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 id="siteCredentialsModalLabel" class="modal-title">
						MQTT-Zugangsdaten — {{ siteName }}
					</h5>
					<button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Schließen"></button>
				</div>
				<div class="modal-body">
					<p class="text-muted small mb-3">
						Diese Daten brauchst du für die evcc-Konfiguration dieser Site.
					</p>

					<div v-if="loading" class="text-center py-4">
						<div class="spinner-border text-primary" role="status">
							<span class="visually-hidden">Lädt…</span>
						</div>
					</div>

					<div v-else-if="error" class="alert alert-danger">{{ error }}</div>

					<template v-else-if="credentials">
						<div class="mb-3">
							<label class="form-label fw-semibold">Broker URL</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="credentials.brokerUrl" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.brokerUrl, 'brokerUrl')">
									{{ copied === 'brokerUrl' ? 'Kopiert!' : 'Kopieren' }}
								</button>
							</div>
						</div>

						<div class="mb-3">
							<label class="form-label fw-semibold">Broker Port</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="credentials.brokerPort" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(String(credentials.brokerPort), 'brokerPort')">
									{{ copied === 'brokerPort' ? 'Kopiert!' : 'Kopieren' }}
								</button>
							</div>
						</div>

						<div class="mb-3">
							<label class="form-label fw-semibold">MQTT-Benutzername</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="credentials.mqttUsername" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.mqttUsername, 'mqttUsername')">
									{{ copied === 'mqttUsername' ? 'Kopiert!' : 'Kopieren' }}
								</button>
							</div>
						</div>

						<div class="mb-3">
							<label class="form-label fw-semibold">MQTT-Passwort</label>
							<div class="input-group">
								<input
									class="form-control"
									:type="showPassword ? 'text' : 'password'"
									:value="credentials.mqttPassword"
									readonly
								/>
								<button class="btn btn-outline-secondary" type="button" @click="showPassword = !showPassword">
									{{ showPassword ? 'Verbergen' : 'Anzeigen' }}
								</button>
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.mqttPassword, 'mqttPassword')">
									{{ copied === 'mqttPassword' ? 'Kopiert!' : 'Kopieren' }}
								</button>
							</div>
						</div>

						<div class="mb-0">
							<label class="form-label fw-semibold">Topic Prefix</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="credentials.topicPrefix" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.topicPrefix, 'topicPrefix')">
									{{ copied === 'topicPrefix' ? 'Kopiert!' : 'Kopieren' }}
								</button>
							</div>
						</div>
					</template>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Schließen</button>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { getAuthToken } from '../services/auth'

interface SiteCredentials {
	brokerUrl: string
	brokerPort: number
	mqttUsername: string
	mqttPassword: string
	topicPrefix: string
}

export default defineComponent({
	name: 'SiteCredentialsModal',
	props: {
		siteId: {
			type: String,
			default: null,
		},
		siteName: {
			type: String,
			default: '',
		},
	},
	data() {
		return {
			credentials: null as SiteCredentials | null,
			loading: false,
			error: '',
			showPassword: false,
			copied: '' as string,
		}
	},
	watch: {
		siteId(newId: string | null) {
			if (newId) {
				this.fetchCredentials(newId)
			} else {
				this.credentials = null
				this.error = ''
				this.showPassword = false
			}
		},
	},
	methods: {
		async fetchCredentials(siteId: string) {
			this.loading = true
			this.error = ''
			this.credentials = null
			this.showPassword = false
			try {
				const token = getAuthToken()
				const resp = await fetch(`/api/sites/${siteId}/credentials`, {
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`,
					},
				})
				if (!resp.ok) throw new Error('Zugangsdaten konnten nicht geladen werden.')
				const data = await resp.json()
				this.credentials = data
			} catch (err: any) {
				this.error = err.message || 'Zugangsdaten konnten nicht geladen werden.'
			} finally {
				this.loading = false
			}
		},
		async copy(value: string, field: string) {
			try {
				await navigator.clipboard.writeText(value)
				this.copied = field
				setTimeout(() => {
					if (this.copied === field) this.copied = ''
				}, 2000)
			} catch {
				// clipboard not available
			}
		},
	},
})
</script>
