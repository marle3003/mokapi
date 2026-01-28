<script setup lang="ts">
defineProps<{ response: KafkaJoinGroupResponse }>();
</script>

<template>
  <div class="kv">
    <span class="key">Generation Id</span>
    <span class="value">{{ response.generationId }}</span>
  </div>
  <div class="kv">
    <span class="key">Member Id</span>
    <span class="value">{{ response.memberId }}</span>
  </div>
  <div class="kv">
    <span class="key">Protocol Name</span>
    <span class="value">{{ response.protocolName }}</span>
  </div>
  <div class="kv">
    <span class="key">Leader Id</span>
    <span class="value">{{ response.leaderId }}</span>
  </div>
  <div class="kv">
    <span class="key">Members</span>
    <span class="value">{{ response.members?.length }}</span>
  </div>
</template>

<style scoped>
.kv {
  display: flex;
  gap: 0.5rem;
}

.key {
  color: #888;
  min-width: 90px;
}

.value {
  font-family: monospace;
}
</style>