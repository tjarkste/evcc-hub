<template>
  <div class="container py-4" v-html="content" />
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from "vue";
export default defineComponent({
  name: "Nutzungsbedingungen",
  setup() {
    const content = ref('<p class="text-muted">Lädt...</p>');
    onMounted(async () => {
      try {
        const res = await fetch("/legal/nutzungsbedingungen.html");
        if (res.ok) content.value = await res.text();
      } catch {
        // content bleibt als Fallback
      }
    });
    return { content };
  },
});
</script>
