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
import { connectMqtt, disconnectMqtt } from "../services/mqtt";
import { getStoredAuth } from "../services/auth";

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
			// Login-Route immer anzeigen — kein startupCompleted nötig vor MQTT-Verbindung
			return this.state.startupCompleted || this.$route.path === '/login';
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
	mounted() {
		const auth = getStoredAuth();
		if (!auth) {
			this.$router.push('/login');
			return;
		}
		connectMqtt({
			brokerUrl: import.meta.env.VITE_MQTT_WSS_URL || 'wss://mqtt.evcc-cloud.de/mqtt',
			username: auth.mqttUsername,
			password: auth.mqttPassword,
			topicPrefix: auth.topicPrefix,
		});
	},
	unmounted() {
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
