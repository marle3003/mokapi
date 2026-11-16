<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    service: { type: Object as PropType<LdapService>, required: true },
})
</script>

<template>
    <div class="table-responsive-sm">
        <table class="table dataTable" aria-label="Servers">
            <thead>
                <tr>
                    <th scope="col" class="text-left">Address</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>{{ service.server }}</td>
                </tr>
            </tbody>
        </table>
    </div>
</template>
