<script setup lang="ts">
import { computed, onMounted, onRenderTriggered } from 'vue';


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
    <div>
      <div class="d-md-none">
        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" style="inset-inline-end: 16px; position:absolute"></button>
        <div>
          <h2>{{ title) }}</h2>
        </div>
      </div>
    <ul class="nav nav-pills root flex-column mb-auto pe-3" v-if="root && root.items">
        <li class="nav-item" v-for="(level1, k1) of root.items">

            <div v-if="hasChildren(level1)" class="chapter">
                <button type="button" class="btn btn-link" :class="isActive(k1) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k1)" :aria-expanded="isActive(k1) || isExpanded(level1)" :aria-controls="getId(k1)" :id="'btn'+getId(k1)">
                    <i class="bi bi-chevron-right"></i> 
                    <i class="bi bi-chevron-down"></i> 
                    {{ k1 }}
                </button>

                <section class="collapse" :class="isActive(k1) || isExpanded(level1) ? 'show' : ''" :id="getId(<string>k1)" :aria-labelledby="'btn'+getId(k1)">
                    <ul v-if="hasChildren(level1)" class="nav nav-pills flex-column mb-auto pe-3">
                        <li class="nav-item" v-for="(level2, k2) of (<DocEntry>level1).items">
                            
                            <div v-if="hasChildren(level2)" class="subchapter">
                                <button type="button" class="btn btn-link" :class="isActive(k1, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(k1, k2)" :aria-controls="getId(k2)">
                                    <i class="bi bi-chevron-right"></i> 
                                    <i class="bi bi-chevron-down"></i> 
                                    {{ k2 }}
                                </button>
                            
                                <div class="collapse" :class="isActive(k1, k2) ? 'show' : ''" :id="getId(k2)">
                                    <ul class="nav nav-pills flex-column mb-auto pe-3">
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

            
            <router-link v-if="!hasChildren(level1) && showItem(k1, level1)" class="nav-link chapter-text" :to="{ name: 'docs', params: {level2: formatParam(k1)} }">{{ k1 }}</router-link>
        </li>
    </ul>   
    </div>
</template>

<style scoped>
h2 {
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
.nav {
  font-size: 0.94rem;
  font-weight: 500;

}
.nav.root {
  padding-top: 1rem;
}

.nav-item {
  margin-left: 16px;
}

.nav .nav-link {
  padding: 0;
  padding-top: 4px;

}

@media only screen and (max-width: 600px)  {
  .nav {
    font-size: 1.7rem;
  }
}

.nav .router-link-active, .nav .router-link-exact-active {
  color: var(--color-nav-link-active);
  font-weight: 600;
}

.nav .child-active {
  color: var(--color-nav-link-active);
}

.chapter {
  margin-bottom: 0.5rem;
}

.chapter > section, .subchapter section {
  border-left: solid 1px var(--color-tabs-border);
  margin-left: 5px;
}

.nav button {
  color: var(--color-text);
  padding: 0;
  padding-top: 4px;
  padding-bottom: 0px;
  font-size: 0.94rem;
  text-decoration: none;
  border: 0;
}

.chapter > button {
  font-weight: 700;
}

.nav button:hover {
  color: var(--color-nav-link-active);
}

.nav button[aria-expanded=false] .bi-chevron-down {
  display: none;
}

.nav button[aria-expanded=true] .bi-chevron-right {
  display: none;
}
</style>