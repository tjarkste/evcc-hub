<template>
  <div class="container py-4">
    <p v-if="loading" class="text-muted">{{ $t('hub.legal.loading') }}</p>
    <div v-else v-html="content" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "Impressum",
  data() {
    return {
      loading: true,
      content: "",
    };
  },
  async mounted() {
    try {
      const res = await fetch("/legal/impressum.html");
      if (res.ok) this.content = await res.text();
    } catch {
      // content remains empty on error
    }
    this.loading = false;
  },
});
</script>
