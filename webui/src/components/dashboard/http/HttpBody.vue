<script setup lang="ts">
import SourceView from '../SourceView.vue';

defineProps({
    contentType: { type: String, required: true },
    body: {type: String, required: true}
})
</script>

<template>
  <source-view v-if="body" :source="body" :content-type="contentType" style="max-height: 500px;"></source-view>
  <div v-else>No body returned for response</div>
</template>

<style scoped>
.codeBlock {
  position: relative;
}
.card-body .codeBlock .label {
    position: absolute;
    right: 4px;
    top: 4px;
    font-size: 0.8rem;
  }
</style>