<template>
	<div class="hub-mode-notice alert alert-secondary d-flex align-items-start gap-2 my-2">
		<span class="text-muted small">{{ message }}</span>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
	name: 'HubModeNotice',
	props: {
		message: {
			type: String,
			required: true,
		},
	},
})
</script>

<style scoped>
.hub-mode-notice {
	border-radius: 0.5rem;
}
</style>
