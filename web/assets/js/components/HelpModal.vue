<template>
	<Teleport to="body">
		<div
			id="helpModal"
			class="modal fade text-dark"
			tabindex="-1"
			role="dialog"
			aria-hidden="true"
		>
			<div class="modal-dialog modal-dialog-centered" role="document">
				<div class="modal-content">
					<div class="modal-header">
						<h5 class="modal-title">{{ $t("help.modalTitle") }}</h5>
						<button
							type="button"
							class="btn-close"
							data-bs-dismiss="modal"
							aria-label="Close"
						></button>
					</div>
					<div class="modal-body">
						<p>
							{{ $t("help.primaryActions") }}
						</p>
						<div
							class="d-block d-sm-flex justify-content-between align-items-stretch mb-4"
						>
							<a
								:href="docsUrl"
								target="_blank"
								class="btn btn-outline-primary w-100 w-sm-auto flex-grow-1 mb-3 mb-sm-0 me-sm-3"
								type="button"
							>
								{{ $t("help.documentationButton") }}
							</a>
							<a
								href="https://github.com/evcc-io/evcc/discussions"
								target="_blank"
								class="btn btn-outline-primary w-100 w-sm-auto flex-grow-1 mb-3 mb-sm-0 me-sm-3"
								type="button"
							>
								{{ $t("help.discussionsButton") }}
							</a>
							<a
								href="https://github.com/tjarkste/evcc-hub"
								target="_blank"
								class="btn btn-outline-secondary w-100 w-sm-auto flex-grow-1"
								type="button"
							>
								{{ $t("hub.help.hubGithub") }}
							</a>
						</div>
					</div>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<script lang="ts">
import { docsPrefix } from "../i18n";
import { defineComponent } from "vue";

export default defineComponent({
	name: "HelpModal",
	computed: {
		docsUrl() {
			return `${docsPrefix()}/`;
		},
	},
});
</script>
