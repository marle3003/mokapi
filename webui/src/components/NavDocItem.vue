<script setup lang="ts">
import { computed } from 'vue'
import { useDocNav } from '@/composables/doc-nav'

defineOptions({
  name: 'NavDocItem'
})

const props = defineProps<{
  current: DocEntry
  actives: DocEntry[] | undefined
}>()

const { kind: kindOf, isActive, isCurrentPage, isExpanded, getId } = useDocNav(() => props.actives)
const kind = computed(() => kindOf(props.current))
const isHeadline = computed(() => kind.value === 'headline-link' || kind.value === 'headline-label')
const hasItems = computed(() => isHeadline.value || kind.value === 'group-link' || kind.value === 'group-button')
</script>

<template>
  <div v-if="hasItems" :class="isHeadline ? 'headline' : 'chapter'">
    <router-link v-if="kind === 'headline-link'" class="nav-link" :class="{ active: isCurrentPage(current) }"
      :to="{ path: current.path }" :id="'section-' + getId(current)">
      {{ current.label }}
    </router-link>
    <div v-else-if="kind === 'headline-label'" :id="'section-' + getId(current)">{{ current.label }}</div>
    <div v-else class="d-flex align-items-center justify-content-between">
      <router-link v-if="kind === 'group-link'" class="nav-link" :class="{ active: isCurrentPage(current) }"
        :to="{ path: current.path }" :id="'section-' + getId(current)">
        {{ current.label }}
      </router-link>
      <button type="button" v-else class="btn btn-link w-100 text-start" :class="{ 'child-active': isActive(current) }"
        data-bs-toggle="collapse" :data-bs-target="'#' + getId(current)"
        :aria-expanded="isExpanded(current)" :aria-controls="getId(current)"
        :id="'section-' + getId(current)">
        {{ current.label }}
      </button>
      <button type="button" class="btn btn-link" :class="{ 'child-active': isActive(current) }"
        :aria-label="'Toggle ' + current.label"
        data-bs-toggle="collapse" :data-bs-target="'#' + getId(current)"
        :aria-expanded="isExpanded(current)" :aria-controls="getId(current)">
        <span class="bi bi-chevron-right"></span>
        <span class="bi bi-chevron-down"></span>
      </button>
    </div>

    <section class="collapse" :class="{ show: isExpanded(current), chapter: !isHeadline }"
      :id="getId(current)" :aria-labelledby="'section-' + getId(current)">
      <ul class="nav nav-pills flex-column mb-auto">
        <li class="nav-item" v-for="item of current.items" :key="item.label">
          <nav-doc-item :current="item" :actives="actives" />
        </li>
      </ul>
    </section>
  </div>

  <div v-else>
    <router-link v-if="kind === 'leaf'" :class="{ active: isCurrentPage(current) }"
      class="nav-link chapter-text" :to="{ path: current.path }">
      {{ current.label }}
    </router-link>
  </div>
</template>
