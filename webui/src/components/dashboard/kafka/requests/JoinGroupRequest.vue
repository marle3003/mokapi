<script setup lang="ts">
defineProps<{ request: KafkaJoinGroupRequest }>();
</script>

<template>
  <div class="kv">
    <span class="key">Group Name</span>
    <span class="value">{{ request.groupName }}</span>
  </div>
  <div class="kv">
    <span class="key">Member Id</span>
    <span class="value">{{ request.memberId }}</span>
  </div>
  <div class="kv">
    <span class="key">Protocol Type</span>
    <span class="value">{{ request.protocolType }}</span>
  </div>
  <div class="kv">
    <span class="key">Protocols</span>
    <span class="value">{{ request.protocols.join(', ') }}</span>
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