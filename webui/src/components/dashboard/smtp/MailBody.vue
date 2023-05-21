<script setup lang="ts">
defineProps({
    body: { type: String, required: true },
  }
)
</script>

<template>
  <div class="card-group">
    <div class="card">
      <div class="card-body">
        <div class="row">
          <p class="col">
            <p class="label">Body</p>
            <p>{{ body }}</p>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>