<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    headers: { type: Object as PropType<HttpHeader>, required: true },
})

</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left col-3">Name</th>
                <th scope="col" class="text-left col">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(value, name) of headers">
                <td>{{ name }}</td>
                <td>{{ value }}</td>
            </tr>
        </tbody>
    </table>
</template>