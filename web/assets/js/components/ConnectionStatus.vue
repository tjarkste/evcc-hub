<!-- web/assets/js/components/ConnectionStatus.vue -->
<template>
  <div
    v-if="visible"
    class="connection-status d-flex align-items-center gap-1 small"
    :class="statusClass"
    role="status"
    data-testid="connection-status"
  >
    <span
      v-if="isReconnecting"
      class="spinner-border spinner-border-sm"
      role="status"
      aria-hidden="true"
    ></span>
    <span>{{ statusText }}</span>
    <span v-if="staleText" class="text-muted ms-1">{{ staleText }}</span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import store from '../store'
import { ConnectionState } from '../types/evcc'

export default defineComponent({
  name: 'ConnectionStatus',
  data() {
    return {
      staleTimer: null as ReturnType<typeof setInterval> | null,
    }
  },
  computed: {
    connectionState(): ConnectionState {
      return store.state.connectionState
    },
    lastDataAt(): number | null {
      return store.state.lastDataAt
    },
    isReconnecting(): boolean {
      return this.connectionState === ConnectionState.RECONNECTING
    },
    visible(): boolean {
      return this.connectionState !== ConnectionState.CONNECTED || this.isStale
    },
    isStale(): boolean {
      if (!this.lastDataAt) return false
      return Date.now() - this.lastDataAt > 60_000
    },
    statusText(): string {
      switch (this.connectionState) {
        case ConnectionState.RECONNECTING:
          return 'Reconnecting...'
        case ConnectionState.OFFLINE:
          return 'Offline'
        default:
          return ''
      }
    },
    staleText(): string {
      if (!this.isStale || !this.lastDataAt) return ''
      const seconds = Math.round((Date.now() - this.lastDataAt) / 1000)
      if (seconds < 120) return `(last update ${seconds}s ago)`
      const minutes = Math.round(seconds / 60)
      return `(last update ${minutes}m ago)`
    },
    statusClass(): string {
      if (this.connectionState === ConnectionState.OFFLINE) return 'text-danger'
      if (this.isReconnecting) return 'text-warning'
      if (this.isStale) return 'text-muted'
      return ''
    },
  },
  mounted() {
    this.staleTimer = setInterval(() => this.$forceUpdate(), 10_000)
  },
  unmounted() {
    if (this.staleTimer) clearInterval(this.staleTimer)
  },
})
</script>

<style scoped>
.connection-status {
  padding: 0.25rem 0.5rem;
  font-size: 0.8rem;
}
</style>
