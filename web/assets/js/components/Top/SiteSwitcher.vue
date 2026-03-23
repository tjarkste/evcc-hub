<template>
  <div
    v-if="sites.length > 1"
    class="dropdown site-switcher"
    data-testid="site-switcher"
  >
    <button
      class="btn btn-sm btn-outline-secondary dropdown-toggle"
      type="button"
      data-bs-toggle="dropdown"
      data-testid="site-switcher-toggle"
    >
      {{ currentSiteName }}
    </button>
    <ul class="dropdown-menu dropdown-menu-end">
      <li v-for="site in sites" :key="site.id">
        <button
          class="dropdown-item d-flex justify-content-between align-items-center"
          :class="{ active: site.id === selectedSiteId }"
          @click="$emit('site-changed', site)"
          :data-testid="`site-option-${site.id}`"
        >
          {{ site.name }}
          <span v-if="site.id === selectedSiteId" class="ms-2">✓</span>
        </button>
      </li>
    </ul>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { Site } from '../../services/sites'

export default defineComponent({
  name: 'SiteSwitcher',
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
  emits: ['site-changed'],
  computed: {
    currentSiteName(): string {
      const site = this.sites.find(s => s.id === this.selectedSiteId)
      return site?.name ?? '—'
    },
  },
})
</script>

<style scoped>
.site-switcher {
  position: fixed;
  top: 0.75rem;
  right: 0.75rem;
  z-index: 1050;
}
</style>
