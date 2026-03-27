<!-- web/assets/js/components/ErrorBoundary.vue -->
<template>
  <slot v-if="!hasError" />
  <div v-else class="alert alert-warning m-3" role="alert" data-testid="error-boundary">
    <strong>{{ section }} {{ $t('hub.error.failedToRender') }}</strong>
    <p class="mb-1 small">{{ errorMessage }}</p>
    <button class="btn btn-sm btn-outline-secondary" @click="reset">{{ $t('hub.error.retry') }}</button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'ErrorBoundary',
  props: {
    section: { type: String, default: 'This section' },
  },
  data() {
    return {
      hasError: false,
      errorMessage: '',
    }
  },
  errorCaptured(err: Error) {
    this.hasError = true
    this.errorMessage = err.message || this.$t('hub.error.unknown')
    console.error(`[ErrorBoundary:${this.section}]`, err)
    return false
  },
  methods: {
    reset() {
      this.hasError = false
      this.errorMessage = ''
    },
  },
})
</script>
