<script setup lang="ts">
import { computed } from 'vue'

const props = // A labelled row of addresses. Cc and Bcc are missing on most mails, so the row is left out without addresses.
defineProps<{
    label: string
    addresses?: MailAddress[]
}>()

const labelId = computed(() => `${props.label.toLowerCase()}-label`)
</script>

<template>
  <div class="row" v-if="addresses">
    <div class="col">
      <p :id="labelId" class="label">{{ label }}</p>
      <ul class="list-unstyled address-list" :aria-labelledby="labelId">
        <li v-for="(addr, index) of addresses" :key="index">
          <span v-if="index>0">, </span>
          <strong v-if="addr.name">{{ addr.name }}</strong>
          <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
          <span v-else>{{ addr.address }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.address-list li {
  display: inline;
}
</style>
