<template>
	<footer class="footer" data-testid="footer">
		<div class="container py-2">
			<div class="d-flex justify-content-between gap-2">
				<Version v-bind="version" />
				<Savings v-bind="savings" />
			</div>
			<div class="d-flex justify-content-center gap-3 mt-1">
				<router-link to="/impressum" class="text-muted small text-decoration-none">
					Impressum
				</router-link>
				<router-link to="/datenschutz" class="text-muted small text-decoration-none">
					Datenschutz
				</router-link>
				<router-link to="/nutzungsbedingungen" class="text-muted small text-decoration-none">
					Nutzungsbedingungen
				</router-link>
			</div>
		</div>
	</footer>
</template>

<script lang="ts">
import "@h2d2/shopicons/es/filled/testtube";

import Version from "./Version.vue";
import Savings from "../Savings/Savings.vue";
import { defineComponent } from "vue";

export default defineComponent({
	name: "Footer",
	components: { Version, Savings },
	props: {
		version: Object,
		savings: Object,
	},
});
</script>
