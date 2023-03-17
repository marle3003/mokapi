<script setup lang="ts">
import Message from '@/components/Message.vue'
</script>

<template>
    <main>
        <message :message="'Page not found'" :showBackToHome="true"></message>
    </main>
</template>