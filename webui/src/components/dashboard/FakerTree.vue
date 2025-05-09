<script setup lang="ts">
import { useFetch } from "@/composables/fetch"
import VueTree from "@ssthouse/vue3-tree-chart"
import "@ssthouse/vue3-tree-chart/dist/vue3-tree-chart.css"
import { computed, ref } from "vue"

declare interface Node{
  name: string
  custom: boolean
  children: Node[]
}

const response = useFetch('/api/faker/tree', undefined, false)

const data = computed(() => {
  if (!response.data) {
    return []
  }
  response.data.name = 'root'
  return map(response.data)
})

function map(node: Node) {
  let name = node.name
  // make words with 14 characters or longer wrappable
  if (name.length >= 14) {
    name = name.replace(/([a-z])([A-Z])/g, "$1<wbr>$2");
  }
  const data = {
    name: name,
    custom: node.custom,
    children: [] as any[]
  }
  if (!node.children) {
    return data
  }
  for (const n of node.children) {
    if (!n) {
      continue
    }
    data.children.push(map(n))
  }
  return data
}

declare interface VueTree {
  zoomIn: () => void
  zoomOut: () => void
}
const tree = ref<VueTree | null>(null)
function handleScroll(event: WheelEvent) {
  if (!tree.value) {
    return
  }
  if (event.deltaY < 0) {
    tree.value.zoomIn()
  }else if (event.deltaY > 0) {
    tree.value.zoomOut()
  }
  event.preventDefault()
}
</script>

<template>
  <section class="card" aria-labelledby="decisionTree" @wheel="handleScroll">
    <div class="card-body decisionTree">
      <div id="decisionTree" class="card-title text-center">Faker Tree</div>
      <vue-tree
        ref="tree"
        class="region"
        :dataset="data"
        :config="{ nodeWidth: 170, nodeHeight: 100, levelHeight: 150 }"
        :collapse-enabled="false"
        linkStyle="straight">
        <template v-slot:node="{ node }">
          <div class="rich-media-node">
            <span v-html="node.name"></span>
            <i v-if="node.custom" class="bi bi-person-fill-gear" title="customized node"></i>
          </div>
        </template>
      </vue-tree>
      <div class="info" style="">
        <router-link :to="{ path: '/docs/guides/get-started/test-data' }">
          <i class="bi bi-question-circle-fill"></i>
        </router-link>
      </div>
    </div>
  </section>
</template>

<style>
  .decisionTree .node-slot {
    cursor: default;
  }
</style>
<style scoped>
  .region {
    width: 100%;
    height:800px; 
  }
  .decisionTree {
    position: relative;
  }
  .decisionTree > .info {
    position: absolute;
    top: 10px;
    right: 10px;
    cursor: pointer;
    color: var(--link-color);
  }
  .decisionTree > .info:hover {
    color: var(--link-color-active);
  }
  .rich-media-node {
    width: 150px;
    padding: 8px;
    min-height: 60px;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
    text-align: center;
    color: white;
    background-color: var(--link-color);
    color: var(--color-button-text-hover);
    border-radius: 4px;
    position: relative;

  }
  .rich-media-node span {
    margin: 0 auto;
    width: 80%;
    overflow-wrap: break-word;
  }
  .rich-media-node i {
    position: absolute;
    top: 0;
    right: 5px;
  }
</style>