<template>
	<div class="container py-5" style="max-width: 700px;">
		<h2 class="mb-4">Meine Standorte</h2>

		<div class="row g-3">
			<div
				v-for="site in sites"
				:key="site.id"
				class="col-sm-6"
			>
				<div
					class="card h-100"
					:class="{ 'border-primary': site.id === selectedSiteId }"
					data-testid="site-card"
				>
					<div class="card-body d-flex flex-column">
						<h5 class="card-title" :data-testid="`site-card-name-${site.id}`">
							{{ site.name }}
						</h5>
						<p class="card-text text-muted small flex-grow-1">
							{{ site.topicPrefix }}
						</p>
						<div class="d-flex gap-2 mt-auto">
							<button
								class="btn btn-primary flex-grow-1"
								@click="$emit('select-site', site)"
								:data-testid="`view-site-${site.id}`"
								:aria-label="`${site.name} anzeigen`"
							>
								Anzeigen
							</button>
							<button
								class="btn btn-outline-secondary"
								:aria-label="`MQTT-Zugangsdaten für ${site.name}`"
								@click="openCredentials(site)"
							>
								MQTT
							</button>
						</div>
					</div>
					<div v-if="site.id === selectedSiteId" class="card-footer text-muted small">
						Aktiv
					</div>
				</div>
			</div>
		</div>

		<div class="mt-4">
			<router-link to="/sites" class="btn btn-outline-secondary btn-sm">
				Standorte verwalten
			</router-link>
		</div>

		<SiteCredentialsModal
			:siteId="credentialsSiteId"
			:siteName="credentialsSiteName"
		/>
	</div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import Modal from 'bootstrap/js/dist/modal'
import type { Site } from '../services/sites'
import SiteCredentialsModal from '../components/SiteCredentialsModal.vue'

export default defineComponent({
	name: 'SiteOverview',
	components: {
		SiteCredentialsModal,
	},
	props: {
		sites: {
			type: Array as PropType<Site[]>,
			required: true,
		},
		selectedSiteId: {
			type: String as PropType<string | null>,
			default: null,
		},
	},
	emits: ['select-site'],
	data() {
		return {
			credentialsSiteId: null as string | null,
			credentialsSiteName: '',
		}
	},
	methods: {
		openCredentials(site: Site) {
			this.credentialsSiteId = site.id
			this.credentialsSiteName = site.name
			this.$nextTick(() => {
				const el = document.getElementById('siteCredentialsModal')
				if (el) {
					Modal.getOrCreateInstance(el).show()
				}
			})
		},
	},
})
</script>
