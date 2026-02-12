<script setup lang="ts">
defineOptions({
  name: 'NavDocItem'
})

const props = defineProps<{
  current: DocEntry
  actives: DocEntry[] | undefined
}>()

function isActive(entry: DocEntry) {
  if (!props.actives) {
    return false
  }
  return props.actives.find(x => x === entry) !== undefined;
}
function getId(entry: DocEntry) {
  let s = entry.path
  if (!s) {
    s = entry.label
  }
  return s.replaceAll("/", "-").replaceAll(" ", "-")
}
function hasChildren(entry: DocEntry) {
  if (entry.items) {
    return entry.items.length > 0
  }
  return false
}
function showItem(entry: DocEntry) {
  return !entry.hideInNavigation
}
function isExpanded(item: DocEntry | string) {
  if (typeof item === 'string') {
    return false
  }
  return item.expanded || false
}
</script>

<template>
  <div v-if="hasChildren(current)" :class="{ headline: current.type === 'headline', chapter: !current.type }">
    <div v-if="current.type !== 'headline'" class="d-flex align-items-center justify-content-between">
      <router-link v-if="current.index" class="nav-link" :class="{ active: isActive(current) }"
        :to="{ path: current.path }" :id="'btn' + getId(current)">
        {{ current.label }}
      </router-link>
      <button type="button" v-else class="btn btn-link w-100 text-start" :class="{ 'child-active': isActive(current) }"
        data-bs-toggle="collapse" :data-bs-target="'#' + getId(current)"
        :aria-expanded="isActive(current) || isExpanded(current)" :aria-controls="getId(current)"
        :id="'section-' + getId(current)">
        {{ current.label }}
      </button>
      <button type="button" class="btn btn-link" :class="{ 'child-active': isActive(current) }"
        data-bs-toggle="collapse" :data-bs-target="'#' + getId(current)"
        :aria-expanded="isActive(current) || isExpanded(current)" :aria-controls="getId(current)">
        <span class="bi bi-chevron-right"></span>
        <span class="bi bi-chevron-down"></span>
      </button>
    </div>
    <div v-else :id="'section-' + getId(current)">{{ current.label }}</div>

    <section class="collapse"
      :class="{ show: isActive(current) || isExpanded(current) || current.type === 'headline', chapter: current.type !== 'headline' }"
      :id="getId(current)" :aria-labelledby="'section-' + getId(current)">
      <ul v-if="hasChildren(current)" class="nav nav-pills flex-column mb-auto">
        <li class="nav-item" v-for="item of current.items">
          <nav-doc-item :current="item" :actives="actives" />
        </li>
      </ul>
    </section>
  </div>

  <div v-else>
    <router-link v-if="!hasChildren(current) && showItem(current)" :class="{ active: isActive(current) }"
      class="nav-link chapter-text" :to="{ path: current.path }">
      {{ current.label }}
    </router-link>
  </div>
</template>