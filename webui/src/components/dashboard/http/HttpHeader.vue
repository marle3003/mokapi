<script setup lang="ts">
import type { HttpHeader } from '@/composables/events'
import type { PropType } from 'vue';

const props = defineProps({
    headers: { type: Object as PropType<HttpHeader>, required: true },
})

</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left w-25">Name</th>
                <th scope="col" class="text-left">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(k, v) of headers">
                <td>{{ k }}</td>
                <td>{{ v }}</td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
.w-10{
    width: 10%;
}
</style>