<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Endpoints</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Path</th>
                        <th scope="col" class="text-left">Operations</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="path in service.paths" :key="path.path">
                        <td>{{ path.path }}</td>
                        <td>
                            <span v-for="operation in path.operations" key="operation.method" class="badge operation" :class="operation.method">{{ operation.method }}</span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>