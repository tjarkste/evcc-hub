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
						{{ $t('hub.sites.credentials.title', { siteName }) }}
					</h5>
					<button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
				</div>
				<div class="modal-body">
					<p class="text-muted small mb-3">
						{{ $t('hub.sites.credentials.description') }}
					</p>

					<div v-if="loading" class="text-center py-4">
						<div class="spinner-border text-primary" role="status">
							<span class="visually-hidden">{{ $t('hub.sites.credentials.loading') }}</span>
						</div>
					</div>

					<div v-else-if="error" class="alert alert-danger">{{ error }}</div>

					<template v-else-if="credentials">
						<div class="mb-3">
							<label class="form-label fw-semibold">{{ $t('hub.sites.credentials.broker') }}</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="brokerWithPort" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(brokerWithPort, 'broker')">
									{{ copied === 'broker' ? $t('hub.sites.credentials.copied') : $t('hub.sites.credentials.copy') }}
								</button>
							</div>
						</div>

						<div class="mb-3">
							<label class="form-label fw-semibold">{{ $t('hub.sites.credentials.topicPrefix') }}</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="credentials.topicPrefix" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.topicPrefix, 'topicPrefix')">
									{{ copied === 'topicPrefix' ? $t('hub.sites.credentials.copied') : $t('hub.sites.credentials.copy') }}
								</button>
							</div>
						</div>

						<div class="mb-3">
							<label class="form-label fw-semibold">{{ $t('hub.sites.credentials.mqttUsername') }}</label>
							<div class="input-group">
								<input type="text" class="form-control" :value="credentials.mqttUsername" readonly />
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.mqttUsername, 'mqttUsername')">
									{{ copied === 'mqttUsername' ? $t('hub.sites.credentials.copied') : $t('hub.sites.credentials.copy') }}
								</button>
							</div>
						</div>

						<div class="mb-3">
							<label class="form-label fw-semibold">{{ $t('hub.sites.credentials.mqttPassword') }}</label>
							<div class="input-group">
								<input
									class="form-control"
									:type="showPassword ? 'text' : 'password'"
									:value="credentials.mqttPassword"
									readonly
								/>
								<button class="btn btn-outline-secondary" type="button" @click="showPassword = !showPassword">
									{{ showPassword ? $t('hub.sites.credentials.hide') : $t('hub.sites.credentials.show') }}
								</button>
								<button class="btn btn-outline-secondary" type="button" @click="copy(credentials.mqttPassword, 'mqttPassword')">
									{{ copied === 'mqttPassword' ? $t('hub.sites.credentials.copied') : $t('hub.sites.credentials.copy') }}
								</button>
							</div>
						</div>

						<p class="text-muted small mb-0">
							<strong>{{ $t('hub.sites.credentials.selfSignedNoteLabel') }}</strong>
							{{ $t('hub.sites.credentials.selfSignedNote') }}
						</p>
					</template>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ $t('hub.sites.credentials.close') }}</button>
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
	computed: {
		brokerWithPort(): string {
			if (!this.credentials) return ''
			const url = this.credentials.brokerUrl || 'tls://evcc-hub.de'
			const port = this.credentials.brokerPort || 8883
			// If the URL already contains a port, use it as-is
			if (url.match(/:\d+$/)) return url
			return `${url}:${port}`
		},
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
				if (!resp.ok) throw new Error(this.$t('hub.sites.credentials.loadError'))
				const data = await resp.json()
				this.credentials = data
			} catch (err: any) {
				this.error = err.message || this.$t('hub.sites.credentials.loadError')
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
