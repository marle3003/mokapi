<script setup lang="ts">
import { computed, ref, type PropType } from 'vue';

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
const showRaw = ref<{[name: string]: boolean}>({})
function renderJsonValue(value: any) {
    try {
        const parsed = typeof value === 'string' ? JSON.parse(value) : value;
        if (typeof parsed === 'string') {
            return parsed;
        }
        return JSON.stringify(parsed, null, 2);
    } catch {
        return value
    }
}
</script>

<template>
    <table class="table table.sm dataTable">
        <thead>
            <tr>
                <th scope="col" style="width:40px"></th>
                <th scope="col" class="text-left w-20">Name</th>
                <th scope="col" class="text-left" style="width:100px;">Type</th>
                <th scope="col" class="text-center" style="width: 130px;">OpenAPI</th>
                <th scope="col" class="text-left" style="width:70%">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="p in sorted">
                <td>
                    <button class="btn btn-sm btn-outline-secondary" style="--bs-btn-padding-y: .1rem; --bs-btn-padding-x: .25rem; --bs-btn-font-size: .75rem;"
                        @click="showRaw[p.name] = !showRaw[p.name]">
                        <i v-if="showRaw[p.name]" class="bi bi-layout-text-sidebar" title="Show parsed value"></i>
                        <i v-else class="bi bi-code" title="Show raw value"></i>
                    </button>
                </td>
                <td class="align-middle">{{ p.name }}</td>
                <td class="align-middle">{{ p.type }}</td>
                <td class="text-center align-middle">{{ p.value ? 'yes' : 'no' }}</td>
                <td class="align-middle">{{ p.value ? (showRaw[p.name] ? p.raw : renderJsonValue(p.value)) : p.raw }}</td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
.w-10{
    width: 10%;
}
.w-20{
    width: 20%;
}
</style>