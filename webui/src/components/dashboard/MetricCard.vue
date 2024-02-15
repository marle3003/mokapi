<script setup lang="ts">

const props = defineProps({
  title: { type: String, required: true },
  value: { required: true },
  additional: String
})

const labelId = 'metric-'+props.title.replaceAll(' ', '-')
</script>

<template>
    <div class="card text-center" role="marquee" aria-live="off" :aria-labelledby="labelId">
        <div class="card-body">
            <div class="card-title" :id="labelId">{{ title }}</div>
            <div class="card-text">{{ value }}</div>
            <div class="card-additional" v-if="additional">{{ additional }}</div>
        </div>
    </div>
</template>