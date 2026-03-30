<template>
	<div class="root safe-area-inset">
		<div class="container px-4">
			<TopHeader :title="$t('hub.account.title')" />

			<div class="wrapper pb-5">

				<!-- ── Sites Section ──────────────────────────────────── -->
				<h2 class="my-4 mt-5">{{ $t('hub.sites.title') }}</h2>

				<div v-if="sitesError" class="alert alert-danger">{{ sitesError }}</div>

				<div class="p-0 config-list">
					<div
						v-for="site in sites"
						:key="site.id"
						class="round-box p-3 mb-2 d-flex justify-content-between align-items-center gap-3"
					>
						<div class="flex-grow-1 min-width-0">
							<div v-if="editingSiteId === site.id" class="d-flex gap-2">
								<input
									v-model="editingSiteName"
									type="text"
									class="form-control form-control-sm"
									@keyup.enter="saveRename(site.id)"
									@keyup.escape="cancelRename"
									ref="renameInput"
								/>
								<button class="btn btn-sm btn-primary text-nowrap" @click="saveRename(site.id)">
									{{ $t('hub.account.save') }}
								</button>
								<button class="btn btn-sm btn-outline-secondary" @click="cancelRename">
									{{ $t('hub.account.cancel') }}
								</button>
							</div>
							<div v-else class="d-flex align-items-center gap-2">
								<span class="fw-semibold text-truncate">{{ site.name }}</span>
								<span v-if="site.id === selectedSiteId" class="badge bg-primary flex-shrink-0">
									{{ $t('hub.sites.active') }}
								</span>
								<button
									class="btn btn-link btn-sm p-0 text-muted flex-shrink-0"
									:aria-label="$t('hub.account.renameSite')"
									@click="startRename(site)"
								>
									<shopicon-regular-edit size="s"></shopicon-regular-edit>
								</button>
							</div>
							<div class="text-muted small text-truncate">{{ site.topicPrefix }}</div>
						</div>
						<div class="d-flex gap-2 flex-shrink-0">
							<button
								v-if="site.id !== selectedSiteId"
								class="btn btn-sm btn-outline-primary"
								@click="selectSite(site)"
							>
								{{ $t('hub.sites.select') }}
							</button>
							<button
								class="btn btn-sm btn-outline-secondary"
								:aria-label="$t('hub.sites.credentials.ariaLabel', { name: site.name })"
								@click="openCredentials(site)"
							>
								MQTT
							</button>
							<button
								class="btn btn-sm btn-outline-danger"
								:disabled="sites.length <= 1"
								@click="deleteSite(site.id)"
							>
								{{ $t('hub.sites.delete') }}
							</button>
						</div>
					</div>
				</div>

				<!-- Add Site -->
				<div class="round-box p-3 mt-2">
					<h5 class="mb-3">{{ $t('hub.sites.addSite') }}</h5>
					<div class="input-group">
						<input
							v-model="newSiteName"
							type="text"
							class="form-control"
							:placeholder="$t('hub.sites.namePlaceholder')"
						/>
						<button
							class="btn btn-primary"
							:disabled="!newSiteName.trim() || creating"
							@click="handleCreateSite"
						>
							{{ creating ? $t('hub.sites.adding') : $t('hub.sites.add') }}
						</button>
					</div>
					<div v-if="createdSite" class="mt-3">
						<div class="alert alert-success mb-0">
							<strong>{{ createdSite.name }}</strong> {{ $t('hub.sites.siteCreatedSuffix') }}
							<p class="mt-2 mb-1">{{ $t('hub.sites.mqttConfigHint') }}</p>
							<pre class="bg-dark text-light p-2 rounded mb-2" style="font-size: 0.85em;">{{ createdSiteConfig }}</pre>
							<p class="text-muted small mb-2">
								<strong>{{ $t('hub.sites.credentials.selfSignedNoteLabel') }}</strong>
								{{ $t('hub.sites.credentials.selfSignedNote') }}
							</p>
							<button class="btn btn-outline-secondary btn-sm" @click="copyCreatedConfig">
								{{ configCopied ? $t('hub.sites.copied') : $t('hub.sites.copy') }}
							</button>
						</div>
					</div>
				</div>

				<!-- ── Account Section ────────────────────────────────── -->
				<h2 class="my-4 mt-5">{{ $t('hub.settings.accountSection') }}</h2>
				<div class="round-box p-4">
					<div v-if="profileLoading" class="text-muted">{{ $t('hub.settings.loading') }}</div>
					<div v-else-if="profileError" class="alert alert-danger mb-0">{{ profileError }}</div>
					<template v-else>
						<div class="mb-3">
							<label class="form-label fw-semibold">{{ $t('hub.settings.emailLabel') }}</label>
							<input type="email" class="form-control" :value="profile.email" readonly />
						</div>
						<div class="mb-0">
							<label class="form-label fw-semibold">{{ $t('hub.settings.registeredSince') }}</label>
							<input type="text" class="form-control" :value="formattedCreatedAt" readonly />
						</div>
					</template>
				</div>

				<!-- ── Change Password ────────────────────────────────── -->
				<h2 class="my-4 mt-5">{{ $t('hub.settings.changePassword') }}</h2>
				<div class="round-box p-4">
					<div v-if="passwordSuccess" class="alert alert-success">
						{{ $t('hub.settings.passwordSuccess') }}
					</div>
					<div v-if="passwordError" class="alert alert-danger">{{ passwordError }}</div>
					<form @submit.prevent="changePassword">
						<div class="mb-3">
							<label for="currentPassword" class="form-label">{{ $t('hub.settings.currentPassword') }}</label>
							<input
								id="currentPassword"
								v-model="currentPassword"
								type="password"
								class="form-control"
								autocomplete="current-password"
								required
							/>
						</div>
						<div class="mb-3">
							<label for="newPassword" class="form-label">{{ $t('hub.settings.newPassword') }}</label>
							<input
								id="newPassword"
								v-model="newPassword"
								type="password"
								class="form-control"
								autocomplete="new-password"
								required
							/>
						</div>
						<div class="mb-3">
							<label for="confirmPassword" class="form-label">{{ $t('hub.settings.confirmPassword') }}</label>
							<input
								id="confirmPassword"
								v-model="confirmPassword"
								type="password"
								class="form-control"
								autocomplete="new-password"
								required
							/>
						</div>
						<button type="submit" class="btn btn-primary" :disabled="passwordSaving">
							{{ passwordSaving ? $t('hub.settings.saving') : $t('hub.settings.savePassword') }}
						</button>
					</form>
				</div>

				<!-- ── Logout ─────────────────────────────────────────── -->
				<div class="mt-4">
					<button type="button" class="btn btn-outline-danger" @click="handleLogout">
						{{ $t('hub.settings.logout') }}
					</button>
				</div>

			</div>
		</div>

		<SiteCredentialsModal
			:siteId="credentialsSiteId"
			:siteName="credentialsSiteName"
		/>
	</div>
</template>

<script lang="ts">
import "@h2d2/shopicons/es/regular/edit";
import { defineComponent } from 'vue'
import Modal from 'bootstrap/js/dist/modal'
import Header from '../components/Top/Header.vue'
import SiteCredentialsModal from '../components/SiteCredentialsModal.vue'
import { fetchSites, createSite, deleteSite as deleteSiteApi, updateSite, getSelectedSiteId, setSelectedSiteId } from '../services/sites'
import { subscribeSite } from '../services/mqtt'
import { getAuthToken, logout } from '../services/auth'
import type { Site } from '../services/sites'

interface Profile {
	email: string
	createdAt: string
}

export default defineComponent({
	name: 'AccountView',
	components: {
		TopHeader: Header,
		SiteCredentialsModal,
	},
	data() {
		return {
			// Sites
			sites: [] as Site[],
			selectedSiteId: getSelectedSiteId(),
			sitesError: '',
			newSiteName: '',
			creating: false,
			createdSite: null as Site | null,
			createdSiteConfig: '',
			configCopied: false,
			editingSiteId: null as string | null,
			editingSiteName: '',
			renameError: '',
			credentialsSiteId: null as string | null,
			credentialsSiteName: '',
			// Profile
			profile: { email: '', createdAt: '' } as Profile,
			profileLoading: true,
			profileError: '',
			// Password
			currentPassword: '',
			newPassword: '',
			confirmPassword: '',
			passwordSaving: false,
			passwordSuccess: false,
			passwordError: '',
		}
	},
	computed: {
		formattedCreatedAt(): string {
			if (!this.profile.createdAt) return ''
			return new Date(this.profile.createdAt).toLocaleDateString(this.$i18n.locale, {
				year: 'numeric',
				month: 'long',
				day: 'numeric',
			})
		},
	},
	async mounted() {
		await Promise.all([this.loadSites(), this.loadProfile()])
	},
	methods: {
		// ── Sites ──────────────────────────────────────────────────────────
		async loadSites() {
			this.sitesError = ''
			try {
				this.sites = await fetchSites()
			} catch {
				this.sitesError = this.$t('hub.sites.loadError')
			}
		},
		selectSite(site: Site) {
			this.selectedSiteId = site.id
			setSelectedSiteId(site.id)
			subscribeSite(site.topicPrefix)
		},
		async handleCreateSite() {
			this.creating = true
			this.sitesError = ''
			this.createdSite = null
			try {
				const site = await createSite(this.newSiteName.trim())
				this.createdSite = site
				this.createdSiteConfig = `mqtt:\n  broker: tls://evcc-hub.de:8883\n  topic: ${site.topicPrefix}\n  user: ${site.mqttUsername}\n  password: "${site.mqttPassword}"`
				this.newSiteName = ''
				this.sites = await fetchSites()
			} catch {
				this.sitesError = this.$t('hub.sites.createError')
			} finally {
				this.creating = false
			}
		},
		async deleteSite(siteId: string) {
			this.sitesError = ''
			try {
				await deleteSiteApi(siteId)
				this.sites = await fetchSites()
				if (this.selectedSiteId === siteId && this.sites.length > 0) {
					this.selectSite(this.sites[0])
				}
			} catch {
				this.sitesError = this.$t('hub.sites.deleteError')
			}
		},
		startRename(site: Site) {
			this.editingSiteId = site.id
			this.editingSiteName = site.name
			this.$nextTick(() => {
				const input = this.$refs['renameInput'] as HTMLInputElement | HTMLInputElement[]
				const el = Array.isArray(input) ? input[0] : input
				el?.focus()
			})
		},
		cancelRename() {
			this.editingSiteId = null
			this.editingSiteName = ''
		},
		async saveRename(siteId: string) {
			const name = this.editingSiteName.trim()
			if (!name) return
			try {
				await updateSite(siteId, name)
				this.sites = await fetchSites()
			} catch {
				// silently ignore — site list will refresh on next load
			} finally {
				this.cancelRename()
			}
		},
		async copyCreatedConfig() {
			await navigator.clipboard.writeText(this.createdSiteConfig)
			this.configCopied = true
			setTimeout(() => { this.configCopied = false }, 2000)
		},
		openCredentials(site: Site) {
			this.credentialsSiteId = site.id
			this.credentialsSiteName = site.name
			this.$nextTick(() => {
				const el = document.getElementById('siteCredentialsModal')
				if (el) Modal.getOrCreateInstance(el).show()
			})
		},
		// ── Profile ────────────────────────────────────────────────────────
		async loadProfile() {
			this.profileLoading = true
			this.profileError = ''
			try {
				const token = getAuthToken()
				const resp = await fetch('/api/auth/profile', {
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`,
					},
				})
				if (!resp.ok) throw new Error(this.$t('hub.settings.profileLoadError'))
				this.profile = await resp.json()
			} catch (err: any) {
				this.profileError = err.message || this.$t('hub.settings.profileLoadError')
			} finally {
				this.profileLoading = false
			}
		},
		// ── Password ───────────────────────────────────────────────────────
		async changePassword() {
			this.passwordError = ''
			this.passwordSuccess = false
			if (this.newPassword !== this.confirmPassword) {
				this.passwordError = this.$t('hub.settings.passwordMismatch')
				return
			}
			this.passwordSaving = true
			try {
				const token = getAuthToken()
				const resp = await fetch('/api/auth/password', {
					method: 'PUT',
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`,
					},
					body: JSON.stringify({
						currentPassword: this.currentPassword,
						newPassword: this.newPassword,
					}),
				})
				if (!resp.ok) {
					const data = await resp.json().catch(() => ({}))
					throw new Error(data.error || this.$t('hub.settings.passwordChangeError'))
				}
				this.passwordSuccess = true
				this.currentPassword = ''
				this.newPassword = ''
				this.confirmPassword = ''
				setTimeout(async () => {
					await logout()
					this.$router.push('/login')
				}, 2000)
			} catch (err: any) {
				this.passwordError = err.message || this.$t('hub.settings.passwordChangeError')
			} finally {
				this.passwordSaving = false
			}
		},
		// ── Logout ─────────────────────────────────────────────────────────
		async handleLogout() {
			await logout()
			this.$router.push('/login')
		},
	},
})
</script>

<style scoped>
.root {
	min-height: 100vh;
	min-height: 100dvh;
}
.config-list {
	display: flex;
	flex-direction: column;
}
</style>
