<template>
	<div
		class="waiting-overlay d-flex flex-column align-items-center justify-content-center"
		data-testid="waiting-for-data"
	>
		<div
			class="spinner-border text-primary mb-3"
			role="status"
			style="width: 3rem; height: 3rem;"
		>
			<span class="visually-hidden">Verbinde...</span>
		</div>
		<p class="text-muted mb-1">Verbinde mit evcc...</p>
		<p class="text-muted small">{{ statusText }}</p>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import store from '../store'
import { ConnectionState } from '../types/evcc'

export default defineComponent({
	name: 'WaitingForData',
	computed: {
		statusText(): string {
			switch (store.state.connectionState) {
				case ConnectionState.CONNECTED:
					return 'Warte auf erste Daten...'
				case ConnectionState.RECONNECTING:
					return 'Verbindung wird wiederhergestellt...'
				case ConnectionState.OFFLINE:
					return 'Keine Verbindung zum Broker.'
				default:
					return ''
			}
		},
	},
})
</script>

<style scoped>
.waiting-overlay {
	min-height: 100vh;
	min-height: 100dvh;
}
</style>
