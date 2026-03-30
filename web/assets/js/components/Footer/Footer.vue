<template>
	<footer class="footer" data-testid="footer">
		<div class="container py-2">
			<div class="d-flex justify-content-between align-items-center gap-2">
				<div class="d-flex gap-3">
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
				<Savings v-bind="savings" />
			</div>
		</div>
	</footer>
</template>

<script lang="ts">
import Savings from "../Savings/Savings.vue";
import { defineComponent } from "vue";

export default defineComponent({
	name: "Footer",
	components: { Savings },
	props: {
		savings: Object,
	},
});
</script>
