<script setup lang="ts">
import { usePrettyText } from '@/composables/usePrettyText';
import { computed } from 'vue';

const props = defineProps<{
    item: SearchItem
}>()

const { adaptiveTruncate } = usePrettyText()

const title = computed(() => {
    let title = props.item.title
    return adaptiveTruncate(title)
})
const route = computed(() => {
    if (props.item.params.mailbox) {
        return { name: 'smtpMailbox', params: { ...{ name: props.item.params.mailbox }, ...props.item.params } }
      }
      return { name: 'mailService', params: props.item.params }
})
</script>

<template>
    <RouterLink :to="route" class="card-body">

        <div class="d-flex justify-content-between">
            <h6 class="mb-1">
                <span class="badge bg-server me-1">Mail</span>
                {{ title }}
            </h6>
        </div>

        <div class="text-muted small mb-1">
            <span v-if="item.domain"> {{ item.domain }} &middot; </span>
            <span>{{ item.type }}</span>
        </div>

        <p class="small mb-0" v-html="item.fragments?.join(' ... ')"></p>

    </RouterLink>
</template>

<style scoped>
.bg-server {
    background-color: #343a40;
}
</style>