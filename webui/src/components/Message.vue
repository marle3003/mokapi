<script setup lang="ts">
defineProps({
    message: { type: String, required: true },
})
</script>

<template>
    <div class="position-absolute top-50 start-50">
        <div class="card-group">
            <div class="card">
                <div class="card-body">{{ message }}</div>
            </div>
        </div>
    </div>
</template>