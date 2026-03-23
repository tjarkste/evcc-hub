<template>
	<div class="app">
		<ErrorBoundary section="Dashboard">
			<router-view
				v-if="showRoutes"
				:notifications="notifications"
				:offline="offline"
			></router-view>
		</ErrorBoundary>
		<ConnectionStatus />
		<SiteSwitcher
			:sites="sites"
			:selected-site-id="selectedSiteId"
			@site-changed="handleSiteChange"
		/>

		<ErrorBoundary section="Settings">
			<GlobalSettingsModal v-bind="globalSettingsProps" />
		</ErrorBoundary>
		<ErrorBoundary section="Battery">
			<BatterySettingsModal v-if="batteryModalAvailabe" v-bind="batterySettingsProps" />
		</ErrorBoundary>
		<ErrorBoundary section="Forecast">
			<ForecastModal v-bind="forecastModalProps" />
		</ErrorBoundary>
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
import ConnectionStatus from "../components/ConnectionStatus.vue";
import ErrorBoundary from "../components/ErrorBoundary.vue";
import SiteSwitcher from "../components/Top/SiteSwitcher.vue";
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
		ConnectionStatus,
		ErrorBoundary,
		SiteSwitcher,
	},
	mixins: [collector],
	props: {
		notifications: Array,
		offline: Boolean,
	},
	data: () => {
		return {
			authNotConfigured: false,
			hasCachedState: false,
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
		handleSiteChange(site: Site) {
			this.selectedSiteId = site.id;
			setSelectedSiteId(site.id);
			this.hasCachedState = false;
			store.reset();
			store.state.lastDataAt = null;
			subscribeSite(site.topicPrefix);
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
