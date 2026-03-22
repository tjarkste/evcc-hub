<template>
	<div class="app">
		<router-view
			v-if="showRoutes"
			:notifications="notifications"
			:offline="offline"
		></router-view>

		<GlobalSettingsModal v-bind="globalSettingsProps" />
		<BatterySettingsModal v-if="batteryModalAvailabe" v-bind="batterySettingsProps" />
		<ForecastModal v-bind="forecastModalProps" />
		<HelpModal />
		<PasswordModal />
		<LoginModal v-bind="loginModalProps" />
		<OfflineIndicator v-if="$route.path !== '/login'" v-bind="offlineIndicatorProps" />
	</div>
</template>

<script lang="ts">
import store from "../store";
import GlobalSettingsModal from "../components/GlobalSettings/GlobalSettingsModal.vue";
import BatterySettingsModal from "../components/Battery/BatterySettingsModal.vue";
import ForecastModal from "../components/Forecast/ForecastModal.vue";
import OfflineIndicator from "../components/Footer/OfflineIndicator.vue";
import PasswordModal from "../components/Auth/PasswordModal.vue";
import LoginModal from "../components/Auth/LoginModal.vue";
import HelpModal from "../components/HelpModal.vue";
import collector from "../mixins/collector";
import { defineComponent } from "vue";
import { connectMqtt, disconnectMqtt, subscribeSite, getCachedTopicPrefix } from "../services/mqtt";
import { loadStateCache } from "../services/stateCache";
import { getStoredAuth, scheduleTokenRefresh, stopTokenRefresh } from "../services/auth";
import { fetchSites, getSelectedSiteId, setSelectedSiteId } from "../services/sites";
import type { Site } from "../services/sites";

export default defineComponent({
	name: "App",
	components: {
		GlobalSettingsModal,
		HelpModal,
		BatterySettingsModal,
		ForecastModal,
		PasswordModal,
		LoginModal,
		OfflineIndicator,
	},
	mixins: [collector],
	props: {
		notifications: Array,
		offline: Boolean,
	},
	data: () => {
		return {
			authNotConfigured: false,
			sites: [] as Site[],
			selectedSiteId: null as string | null,
		};
	},
	head() {
		return { title: "...", titleTemplate: "%s | evcc" };
	},
	computed: {
		version() {
			return store.state.version;
		},
		batteryModalAvailabe() {
			return store.state.battery?.devices?.length;
		},
		showRoutes() {
			// Cloud: immer anzeigen — kein startupCompleted nötig (kein WebSocket-Handshake)
			return true;
		},
		state() {
			const { state, uiLoadpoints } = store;
			return { ...state, uiLoadpoints: uiLoadpoints.value };
		},
		globalSettingsProps() {
			return this.collectProps(GlobalSettingsModal, this.state);
		},
		batterySettingsProps() {
			return this.collectProps(BatterySettingsModal, this.state);
		},
		offlineIndicatorProps() {
			return this.collectProps(OfflineIndicator, this.state);
		},
		forecastModalProps() {
			return this.collectProps(ForecastModal, this.state);
		},
		loginModalProps() {
			return this.collectProps(LoginModal, this.state);
		},
	},
	watch: {
		version(now, prev) {
			if (!!prev && !!now) {
				console.log("new version detected. reloading browser", { now, prev });
				this.reload();
			}
		},
	},
	async mounted() {
		const auth = getStoredAuth();
		if (!auth) {
			this.$router.push('/login');
			return;
		}

		// Restore cached state while waiting for live MQTT data
		const cached = loadStateCache();
		if (cached) {
			store.update(cached);
		}

		scheduleTokenRefresh();

		connectMqtt({
			brokerUrl: import.meta.env.VITE_MQTT_WSS_URL || 'wss://mqtt.evcc-cloud.de/mqtt',
			username: auth.mqttUsername,
			password: auth.mqttPassword,
		});

		// Fetch sites and subscribe; fall back to cached topic if backend is unreachable
		try {
			this.sites = await fetchSites();
			if (this.sites.length > 0) {
				const savedId = getSelectedSiteId();
				const site = this.sites.find(s => s.id === savedId) || this.sites[0];
				this.selectedSiteId = site.id;
				setSelectedSiteId(site.id);
				subscribeSite(site.topicPrefix);
			}
		} catch (e) {
			console.error('Failed to fetch sites:', e);
			const cachedPrefix = getCachedTopicPrefix();
			if (cachedPrefix) {
				console.log('Using cached topic prefix:', cachedPrefix);
				subscribeSite(cachedPrefix);
			}
		}
	},
	unmounted() {
		stopTokenRefresh();
		disconnectMqtt();
	},
	methods: {
		reload() {
			window.location.reload();
		},
	},
});
</script>
<style scoped>
.app {
	min-height: 100vh;
	min-height: 100dvh;
}
</style>
