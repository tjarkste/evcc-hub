<template>
  <div class="container py-4" style="max-width: 600px;">
    <h2 class="mb-4">Meine Standorte</h2>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <!-- Site List -->
    <div v-for="site in sites" :key="site.id" class="card mb-3">
      <div class="card-body d-flex justify-content-between align-items-center">
        <div>
          <h5 class="card-title mb-1">{{ site.name }}</h5>
          <small class="text-muted">{{ site.topicPrefix }}</small>
        </div>
        <div>
          <button
            v-if="site.id !== selectedSiteId"
            class="btn btn-sm btn-outline-primary me-2"
            @click="selectSite(site)"
            data-test="select-site-btn"
          >
            Auswählen
          </button>
          <span v-else class="badge bg-primary me-2">Aktiv</span>
          <button
            class="btn btn-sm btn-outline-danger"
            @click="handleDelete(site.id)"
            :disabled="sites.length <= 1"
            data-test="delete-site-btn"
          >
            Löschen
          </button>
        </div>
      </div>
    </div>

    <!-- Add Site -->
    <div class="card">
      <div class="card-body">
        <h5 class="card-title">Neuen Standort hinzufügen</h5>
        <div class="input-group">
          <input
            v-model="newSiteName"
            type="text"
            class="form-control"
            placeholder="Name (z.B. Ferienhaus)"
            data-test="new-site-name"
          />
          <button
            class="btn btn-primary"
            @click="handleCreate"
            :disabled="!newSiteName.trim() || creating"
            data-test="create-site-btn"
          >
            {{ creating ? '...' : 'Hinzufügen' }}
          </button>
        </div>
        <!-- Show credentials after creation -->
        <div v-if="createdSite" class="mt-3">
          <div class="alert alert-success">
            <strong>{{ createdSite.name }}</strong> wurde erstellt.
            <p class="mt-2 mb-1">Füge diese Zeilen in deine <code>evcc.yaml</code> ein:</p>
            <pre class="bg-dark text-light p-2 rounded" style="font-size: 0.85em;">{{ createdSiteConfig }}</pre>
            <button class="btn btn-outline-secondary btn-sm" @click="copyConfig">
              {{ copied ? 'Kopiert!' : 'Kopieren' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="mt-3">
      <router-link to="/" class="btn btn-outline-secondary">Zurück zum Dashboard</router-link>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { fetchSites, createSite, deleteSite, getSelectedSiteId, setSelectedSiteId } from '../services/sites'
import type { Site } from '../services/sites'
import { subscribeSite } from '../services/mqtt'

export default defineComponent({
  name: 'SiteManager',
  data() {
    return {
      sites: [] as Site[],
      selectedSiteId: getSelectedSiteId(),
      newSiteName: '',
      creating: false,
      error: '',
      createdSite: null as Site | null,
      createdSiteConfig: '',
      copied: false,
    }
  },
  async mounted() {
    try {
      this.sites = await fetchSites()
    } catch {
      this.error = 'Standorte konnten nicht geladen werden.'
    }
  },
  methods: {
    selectSite(site: Site) {
      this.selectedSiteId = site.id
      setSelectedSiteId(site.id)
      subscribeSite(site.topicPrefix)
    },
    async handleCreate() {
      this.creating = true
      this.error = ''
      this.createdSite = null
      try {
        const site = await createSite(this.newSiteName.trim())
        this.createdSite = site
        this.createdSiteConfig = `mqtt:\n  broker: tls://mqtt.evcc-hub.de:8883\n  topic: ${site.topicPrefix}\n  user: ${site.mqttUsername}\n  password: "${site.mqttPassword}"`
        this.newSiteName = ''
        this.sites = await fetchSites()
      } catch {
        this.error = 'Standort konnte nicht erstellt werden.'
      } finally {
        this.creating = false
      }
    },
    async handleDelete(siteId: string) {
      this.error = ''
      try {
        await deleteSite(siteId)
        this.sites = await fetchSites()
        if (this.selectedSiteId === siteId && this.sites.length > 0) {
          this.selectSite(this.sites[0])
        }
      } catch {
        this.error = 'Standort konnte nicht gelöscht werden.'
      }
    },
    async copyConfig() {
      await navigator.clipboard.writeText(this.createdSiteConfig)
      this.copied = true
      setTimeout(() => { this.copied = false }, 2000)
    },
  },
})
</script>
