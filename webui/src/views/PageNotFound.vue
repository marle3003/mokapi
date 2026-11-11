<script setup lang="ts">
import Message from '@/components/Message.vue'
import { setTitle } from '@/composables/meta'

setTitle('Page not found | Mokapi')
</script>

<template>
    <main>
        <message :message="'Page not found'" :showBackToHome="true"></message>
    </main>
</template>