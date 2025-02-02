<script setup lang="ts">
import { computed } from 'vue';


const props = defineProps<{
    levels: string[],
    config: DocConfig,
    title: string
}>()

const root = computed(() => <DocEntry>props.config[props.levels[0]])

function matchLevel(label: any, level: number) {
    if (level > props.levels.length){
        return false
    }
    return label.toString().toLowerCase() == props.levels[level - 1].toLowerCase()
}

function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
function hasChildren(item: DocEntry | string) {
    if (typeof item === 'string') {
        return false
    }
    const entry = <DocEntry>item
    if ('file' in entry || 'component' in entry) {
        return false
    }
    return true
}
function showItem(name: string | number, item: DocConfig | DocEntry | string){
    if (typeof item == 'string' && props.levels[0] != name) {
        return true
    }
    const entry = <DocEntry>item
    if ('hideInNavigation' in entry) {
        return !entry.hideInNavigation
    }
    return true
}
function getId(name: any) {
    return name.toString().replaceAll("/", "-").replaceAll(" ", "-")
}
function isActive(...levels: any[]) {
    for (let i = 0; i < levels.length; i++) {
        if (!matchLevel(levels[i], i+2)) {
            return false
        }
    }
    return true
}
function isExpanded(item: DocEntry | string) {
    if (typeof item === 'string') {
        return false
    }
    return item.expanded || false
}

</script>

<template>
    <nav class="p-4 ps-2">
    <span v-if="root && root.items" class="px-3">{{ title }}</span>
    <hr class="m-2" />
    <ul class="nav nav-pills root flex-column mb-auto px-3" v-if="root && root.items">
        <li class="nav-item" v-for="(level1, k1) of root.items">

            <div v-if="hasChildren(level1)" class="chapter">
              <div class="d-flex align-items-center justify-content-between">
                <router-link v-if="(<DocEntry>level1).index" class="nav-link" :class="levels[1] == k1 && levels.length == 2 ? 'active' : ''" :to="{ name: 'docs', params: {level2: formatParam(k1)} }" :id="'btn'+getId(k1)">{{ k1 }}</router-link>
                <button type="button" v-else class="btn btn-link w-100 text-start" :class="isActive(k1) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k1)" :aria-expanded="isActive(k1) || isExpanded(level1)" :aria-controls="getId(k1)" :id="'btn'+getId(k1)">
                  {{ k1 }}
                </button>
                <button type="button" class="btn btn-link" :class="isActive(k1) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k1)" :aria-expanded="isActive(k1) || isExpanded(level1)" :aria-controls="getId(k1)">
                  <i class="bi bi-caret-up-fill"></i> 
                  <i class="bi bi-caret-down-fill"></i> 
                </button>
              </div>

                <section class="collapse" :class="isActive(k1) || isExpanded(level1) ? 'show' : ''" :id="getId(<string>k1)" :aria-labelledby="'btn'+getId(k1)">
                    <ul v-if="hasChildren(level1)" class="nav nav-pills flex-column mb-auto">
                        <li class="nav-item ps-3" v-for="(level2, k2) of (<DocEntry>level1).items">
                            
                            <div v-if="hasChildren(level2)" class="subchapter">
                              <div class="d-flex align-items-center justify-content-between">
                                <button type="button" class="btn btn-link w-100 text-start" :class="isActive(k1, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(k1, k2)" :aria-controls="getId(k2)">
                                    {{ k2 }}
                                </button>
                                <button type="button" class="btn btn-link" :class="isActive(k1, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(k1, k2)" :aria-controls="getId(k2)">
                                    <i class="bi bi-caret-up-fill"></i> 
                                    <i class="bi bi-caret-down-fill"></i> 
                                </button>
                              </div>
                            
                                <div class="collapse" :class="isActive(k1, k2) ? 'show' : ''" :id="getId(k2)">
                                    <ul class="nav nav-pills flex-column mb-auto">
                                        <li class="nav-item" v-for="(_, k3) of (<DocEntry>level2).items">
                                            <router-link v-if="k2 != k3" class="nav-link" :class="isActive(k1, k2, k3) ? 'active' : ''" :to="{ name: 'docs', params: {level2: formatParam(k1), level3: formatParam(k2), level4: formatParam(k3)} }">{{ k3 }}</router-link>
                                        </li>
                                    </ul>
                                </div>
                            </div>
                            
                            <router-link v-if="k1 != k2 && !hasChildren(level2) && showItem(k2, level2)" class="nav-link" :class="isActive(k1, k2) ? 'active' : ''" :to="{ name: 'docs', params: {level2: formatParam(k1), level3: formatParam(k2)} }">{{ k2 }}</router-link>
                        </li>
                    </ul>
                </section>
            </div>

            
            <router-link v-if="!hasChildren(level1) && showItem(k1, level1)" :class="isActive(k1) ? 'active' : ''"  class="nav-link chapter-text" :to="{ name: 'docs', params: {level2: formatParam(k1)} }">{{ k1 }}</router-link>
        </li>
    </ul>   
  </nav>
</template>

<style scoped>
.page-title {
  font-size: 1.15rem;
  font-weight: 700;
  padding-left: 16px;
  padding-right: 16px;
  margin-top: 0;
  margin-bottom: 0;
  padding-bottom: 1.3rem;
  border-bottom-color: var(--color-background-light);
  border-bottom-style: solid;
  border-bottom-width: 1px;
}
nav {
  font-size: 1rem;
}
.nav, .nav .btn {
  font-size: 0.95rem;
}

.nav-item a, .nav-item .chapter > div {
  padding-top: 7px;
  padding-bottom: 7px;
}

.nav .nav-link {
  padding-left: 0;
}

@media only screen and (max-width: 768px)  {
  .nav {
    font-size: 1.7rem;
  }
}

.nav .active {
  color: var(--color-nav-link-active);
}

.nav button {
  color: var(--color-text);
  padding: 0;
  text-decoration: none;
  border: 0;
}

.nav button:hover {
  color: var(--color-nav-link-active);
}

.nav button[aria-expanded=false] .bi-caret-up-fill {
  display: none;
}

.nav button[aria-expanded=true] .bi-caret-down-fill {
  display: none;
}
</style>