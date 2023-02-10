<script setup lang="ts">
import type { HttpService } from '@/composables/services'
import type { PropType } from 'vue';

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Endpoints</div>
            <div class="card-body">
                <table class="table dataTable selectable">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left">Path</th>
                            <th scope="col" class="text-left">Operations</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="path in service.paths" key="service.name">
                            <td>{{ path.path }}</td>
                            <td>
                                <span v-for="operation in path.operations" key="operation.method" class="badge operation" :class="operation.method">{{ operation.method }}</span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>