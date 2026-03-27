<template>
	<div class="container py-5" style="max-width: 600px;">
		<h2 class="mb-4">{{ $t('hub.settings.title') }}</h2>

		<!-- Account Info -->
		<div class="card mb-4">
			<div class="card-header">
				<h5 class="mb-0">{{ $t('hub.settings.accountSection') }}</h5>
			</div>
			<div class="card-body">
				<div v-if="profileLoading" class="text-muted">{{ $t('hub.settings.loading') }}</div>
				<div v-else-if="profileError" class="alert alert-danger">{{ profileError }}</div>
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
		</div>

		<!-- Change Password -->
		<div class="card mb-4">
			<div class="card-header">
				<h5 class="mb-0">{{ $t('hub.settings.changePassword') }}</h5>
			</div>
			<div class="card-body">
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
					<button
						type="submit"
						class="btn btn-primary"
						:disabled="passwordSaving"
					>
						{{ passwordSaving ? $t('hub.settings.saving') : $t('hub.settings.savePassword') }}
					</button>
				</form>
			</div>
		</div>

		<!-- Logout -->
		<div class="text-end">
			<button type="button" class="btn btn-outline-danger" @click="handleLogout">
				{{ $t('hub.settings.logout') }}
			</button>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { getAuthToken, logout } from '../services/auth'

interface Profile {
	email: string
	createdAt: string
}

export default defineComponent({
	name: 'SettingsView',
	data() {
		return {
			profile: { email: '', createdAt: '' } as Profile,
			profileLoading: true,
			profileError: '',
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
		await this.loadProfile()
	},
	methods: {
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
				const data = await resp.json()
				this.profile = data
			} catch (err: any) {
				this.profileError = err.message || this.$t('hub.settings.profileLoadError')
			} finally {
				this.profileLoading = false
			}
		},
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
		async handleLogout() {
			await logout()
			this.$router.push('/login')
		},
	},
})
</script>
