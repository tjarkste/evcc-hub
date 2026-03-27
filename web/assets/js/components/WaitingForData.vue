<template>
	<div class="waiting-overlay d-flex flex-column align-items-center justify-content-center" data-testid="waiting-for-data">
		<!-- Stage 3: No data after timeout -->
		<template v-if="noDataTimeout">
			<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" fill="currentColor"
				class="text-warning mb-3" viewBox="0 0 16 16">
				<path d="M8.982 1.566a1.13 1.13 0 0 0-1.96 0L.165 13.233c-.457.778.091 1.767.98 1.767h13.713c.889 0 1.438-.99.98-1.767L8.982 1.566zM8 5c.535 0 .954.462.9.995l-.35 3.507a.552.552 0 0 1-1.1 0L7.1 5.995A.905.905 0 0 1 8 5zm.002 6a1 1 0 1 1 0 2 1 1 0 0 1 0-2z"/>
			</svg>
			<p class="text-muted mb-1">{{ $t('hub.waiting.noData') }}</p>
			<p class="text-muted small">{{ $t('hub.waiting.noDataHint') }}</p>
		</template>
		<!-- Stages 1 & 2: Connecting / Waiting -->
		<template v-else>
			<div class="spinner-border text-primary mb-3" role="status" style="width: 3rem; height: 3rem;">
				<span class="visually-hidden">{{ $t('hub.waiting.loading') }}</span>
			</div>
			<p class="text-muted mb-1">{{ statusTitle }}</p>
		</template>
	</div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import store from "../store";
import { ConnectionState } from "../types/evcc";

export default defineComponent({
	name: "WaitingForData",
	data() {
		return {
			connectedElapsed: 0,
			timer: null as ReturnType<typeof setInterval> | null,
			connectedSince: null as number | null,
		};
	},
	computed: {
		isConnected(): boolean {
			return store.state.connectionState === ConnectionState.CONNECTED;
		},
		noDataTimeout(): boolean {
			return this.isConnected && store.state.lastDataAt === null && this.connectedElapsed >= 30;
		},
		statusTitle(): string {
			if (store.state.connectionState === ConnectionState.RECONNECTING) {
				return this.$t('hub.waiting.connecting');
			}
			if (this.isConnected) {
				return this.$t('hub.waiting.waitingForData');
			}
			return this.$t('hub.waiting.connecting');
		},
	},
	watch: {
		isConnected(connected) {
			if (connected) {
				this.startTimer();
			} else {
				this.stopTimer();
			}
		},
	},
	mounted() {
		if (this.isConnected) {
			this.startTimer();
		}
	},
	unmounted() {
		this.stopTimer();
	},
	methods: {
		startTimer() {
			this.connectedSince = Date.now();
			this.connectedElapsed = 0;
			this.stopTimer();
			this.timer = setInterval(() => {
				if (this.connectedSince) {
					this.connectedElapsed = (Date.now() - this.connectedSince) / 1000;
				}
			}, 1000);
		},
		stopTimer() {
			if (this.timer) {
				clearInterval(this.timer);
				this.timer = null;
			}
		},
	},
});
</script>

<style scoped>
.waiting-overlay {
	flex: 1;
}
</style>
