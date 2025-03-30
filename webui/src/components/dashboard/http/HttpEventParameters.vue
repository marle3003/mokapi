<script setup lang="ts">
import { computed, type PropType } from 'vue';

const props = defineProps({
    parameters: { type: Object as PropType<HttpEventParameter[]>, required: true },
})
const sorted = computed(() => {
    return props.parameters.sort((p1, p2) => {
        if (p1.value && p2.value) {
            return p1.name.localeCompare(p2.name)
        }
        if (p1.value) {
            return -1
        }
        return 1
    })
})
</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left w-25">Name</th>
                <th scope="col" class="text-left w-10">Type</th>
                <th scope="col" class="text-center w-10">OpenAPI</th>
                <th scope="col" class="text-left">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="p in sorted">
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