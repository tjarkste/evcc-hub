<template>
	<div class="container py-5" style="max-width: 600px;">
		<h2 class="mb-4">Profil & Einstellungen</h2>

		<!-- Account Info -->
		<div class="card mb-4">
			<div class="card-header">
				<h5 class="mb-0">Konto</h5>
			</div>
			<div class="card-body">
				<div v-if="profileLoading" class="text-muted">Lädt...</div>
				<div v-else-if="profileError" class="alert alert-danger">{{ profileError }}</div>
				<template v-else>
					<div class="mb-3">
						<label class="form-label fw-semibold">E-Mail-Adresse</label>
						<input type="email" class="form-control" :value="profile.email" readonly />
					</div>
					<div class="mb-0">
						<label class="form-label fw-semibold">Registriert seit</label>
						<input type="text" class="form-control" :value="formattedCreatedAt" readonly />
					</div>
				</template>
			</div>
		</div>

		<!-- Change Password -->
		<div class="card mb-4">
			<div class="card-header">
				<h5 class="mb-0">Passwort ändern</h5>
			</div>
			<div class="card-body">
				<div v-if="passwordSuccess" class="alert alert-success">
					Passwort erfolgreich geändert. Du wirst in Kürze abgemeldet…
				</div>
				<div v-if="passwordError" class="alert alert-danger">{{ passwordError }}</div>
				<form @submit.prevent="changePassword">
					<div class="mb-3">
						<label for="currentPassword" class="form-label">Aktuelles Passwort</label>
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
						<label for="newPassword" class="form-label">Neues Passwort</label>
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
						<label for="confirmPassword" class="form-label">Neues Passwort bestätigen</label>
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
						{{ passwordSaving ? 'Speichern…' : 'Passwort speichern' }}
					</button>
				</form>
			</div>
		</div>

		<!-- Logout -->
		<div class="text-end">
			<button type="button" class="btn btn-outline-danger" @click="handleLogout">
				Abmelden
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
			return new Date(this.profile.createdAt).toLocaleDateString('de-DE', {
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
				if (!resp.ok) throw new Error('Profil konnte nicht geladen werden.')
				const data = await resp.json()
				this.profile = data
			} catch (err: any) {
				this.profileError = err.message || 'Profil konnte nicht geladen werden.'
			} finally {
				this.profileLoading = false
			}
		},
		async changePassword() {
			this.passwordError = ''
			this.passwordSuccess = false

			if (this.newPassword !== this.confirmPassword) {
				this.passwordError = 'Die neuen Passwörter stimmen nicht überein.'
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
					throw new Error(data.error || 'Passwort konnte nicht geändert werden.')
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
				this.passwordError = err.message || 'Passwort konnte nicht geändert werden.'
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
