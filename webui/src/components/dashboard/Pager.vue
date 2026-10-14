<script setup lang="ts">
import { computed } from 'vue';
// max pages in pagination
const maxVisiblePages = 10

const props = defineProps<{
    current: number
    size: number
    total: number
}>()

const emit = defineEmits<{
  (e: "page:changed", count: number): void
}>();

const pages = computed(() => {
  return Math.ceil(props.total / props.size)
})

const range = computed(() => {
  const half = Math.floor(maxVisiblePages / 2)

  let start = props.current - half
  let end = props.current + half

  if (start < 1) {
    start = 1
    end = maxVisiblePages
  }

  if (end > pages.value) {
    end = pages.value
    start = Math.max(1, pages.value - maxVisiblePages + 1)
  }

  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})

</script>

<template>
    <nav v-if="pages > 1" aria-label="Page navigation">
        <ul class="pagination justify-content-center">
            <li class="page-item" v-if="current > 1">
                <a class="page-link" aria-label="Previous" @click="emit('page:changed', current - 1)">
                    <span aria-hidden="true">&laquo;</span>
                </a>
            </li>
            <li v-for="index in range" :key="index" class="page-item"
                :class="index === current? 'active' : ''">
                <a class="page-link" @click="emit('page:changed', index)">{{ index }}</a>
            </li>
            <li class="page-item" v-if="current < pages">
                <a class="page-link" aria-label="Next" @click="emit('page:changed', current + 1)">
                    <span aria-hidden="true">&raquo;</span>
                </a>
            </li>
        </ul>
    </nav>
</template>