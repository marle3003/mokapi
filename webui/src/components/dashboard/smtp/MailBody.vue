<script setup lang="ts">
import { ref } from 'vue'

defineProps({
    body: { type: String, required: true },
    contentType: { type: String, required: true },
  }
)

const showBodyLabel = ref<boolean>(false)
</script>

<template>
  <div class="card-group" @mouseover="showBodyLabel = true" @mouseleave="showBodyLabel = false">
    <div class="card">
      <div class="card-body">
        <div class="card-title text-center bodyLabel" :style="[showBodyLabel ? '' : 'display: none']">Body</div>
        <div class="row">
          <p class="col">
            <p v-if="contentType == 'text/html'" v-html="body"></p>
            <p v-else>{{ body }}</p>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .bodyLabel {
    position: absolute;
    left: 50%;
  }
</style>