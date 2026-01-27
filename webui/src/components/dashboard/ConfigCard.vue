<script setup lang="ts">
import { computed } from 'vue'
import Configs from './Configs.vue';

const props = defineProps<{
    configs?: Config[] | ConfigRef[] | null,
    title?: string,
    hideTitle?: boolean
}>()

const title = computed(() => props.title ? props.title : "Configs")
</script>

<template>
  <section class="card" aria-labelledby="configs">
      <div class="card-body">
          <h2 v-if="!hideTitle" id="configs" class="card-title text-center">{{ title }}</h2>
          <configs :configs="configs" />
      </div>
  </section>
</template>