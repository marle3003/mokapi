<script setup lang="ts">
import { useTemplateRef, type ComponentPublicInstance } from 'vue'
import RegexInput from '@/components/RegexInput.vue'

const url = defineModel<string>({ required: true })
const input = useTemplateRef<ComponentPublicInstance<typeof RegexInput>>('input')
</script>

<template>
    <RegexInput v-model="url" ref="input" placeholder="[regex]" aria-label="URL filter. Supports regular expressions." />
    <div class="invalid-feedback" style="display: inline;" v-if="input?.regexError">
        Invalid regular expression: {{ input.regexError }}
    </div>
</template>
