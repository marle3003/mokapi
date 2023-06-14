<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    parameters: { type: Object as PropType<HttpEventParameter[]>, required: true },
})
</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left w-25">Name</th>
                <th scope="col" class="text-left w-10">Type</th>
                <th scope="col" class="text-center w-15">OpenAPI Specification</th>
                <th scope="col" class="text-left">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="p in parameters">
                <td>{{ p.name }}</td>
                <td>{{ p.type }}</td>
                <td class="text-center">{{ p.value ? 'yes' : 'no' }}</td>
                <td>{{ p.value ? p.value : p.raw }}</td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
.w-10{
    width: 10%;
}
</style>