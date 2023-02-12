<script setup lang="ts">
defineProps({
  title: { type: String, required: true },
  value: { required: true },
  additional: String
})
</script>

<template>
    <div class="card text-center">
        <div class="card-body">
            <div class="card-title">{{ title }}</div>
            <div class="card-text">{{ value }}</div>
            <div class="card-additional" v-if="additional">{{ additional }}</div>
        </div>
    </div>
</template>