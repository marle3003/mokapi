<script setup lang="ts">
import Pager from './Pager.vue';

const props = defineProps<{
    count: number
    total: number
    options?: number[]
}>()

const defaultOptions = [10, 25, 50, 100]

const size = defineModel<number>('size', { required: true })
const current = defineModel<number>('current', { required: true })

</script>

<template>
    <div class="row" v-if="total > 0">
        <div class="d-flex justify-content-between align-items-center mb-3">

            <div class="small text-muted">
                Showing {{ count }} of {{ total }} requests
            </div>

            <Pager :current="current" :size="size" :total="total" @page:changed="(index: number) => current = index" />

            <select class="form-select form-select-sm" style="width:auto" v-model.number="size">
                <option v-for="size in options ?? defaultOptions" :key="size" :value="size">
                    {{ size }}
                </option>
            </select>

        </div>
    </div>
</template>