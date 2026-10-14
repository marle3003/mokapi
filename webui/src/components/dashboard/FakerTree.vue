<script setup lang="ts">
import { useFetch } from "@/composables/fetch"
import VueTree from "@ssthouse/vue3-tree-chart"
import "@ssthouse/vue3-tree-chart/dist/vue3-tree-chart.css"
import { computed, ref, useTemplateRef } from "vue"
import { Modal } from 'bootstrap';
import hljs from '@/plugins/highlight'

declare interface Node {
  name: string
  path: string
  custom: boolean
  children: Node[]
  dependsOn: string[]
  weight: number
  attributes: string[]
  hasFakeFunc: boolean
}

declare interface Data {
  name: string,
  path: string,
  custom: boolean,
  children: Data[],
  weight: number
  attributes?: { name: string }[]
  dependsOn?: { name: string }[]
  hasFakeFunc: boolean
}

const response = useFetch('/api/faker/tree', undefined, false)
const selected = ref<Data | null>(null);
const example = ref<any>(undefined)
const dialogRef = useTemplateRef('dialogRef')
const dialog = ref<Modal | undefined>()
const data = computed(() => {
  if (!response.data) {
    return []
  }
  response.data.name = 'root'
  return map(response.data)
  //return response.data
})

function map(node: Node) {
  let name = node.name
  // make words with 14 characters or longer wrappable
  if (name.length >= 14) {
    name = name.replace(/([a-z])([A-Z])/g, "$1<wbr>$2");
  }
  const data: Data = {
    name: name,
    path: node.path,
    custom: node.custom,
    weight: node.weight,
    children: [] as any[],
    attributes: node.attributes?.map(x => { return { name: x } }),
    dependsOn: node.dependsOn?.map(x => { return { name: x } }),
    hasFakeFunc: node.hasFakeFunc
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
  } else if (event.deltaY > 0) {
    tree.value.zoomOut()
  }
  event.preventDefault()
}
function openDialog(node: Data, event: MouseEvent) {
  selected.value = node;
  example.value = undefined
  getExample()
  if (!dialog.value) {
    dialog.value = new Modal(dialogRef.value!)
  }
  console.log('node', node)
  dialog.value.show()
}
async function getExample() {
  if (!selected.value || !selected.value.hasFakeFunc) {
    return
  }
  example.value = useFetch(`/api/faker/node/${selected.value.name}/fake`, undefined, false, false)
}
</script>

<template>
  <section class="card" aria-labelledby="decisionTree" @wheel="handleScroll">
    <div class="card-body decisionTree">
      <div id="decisionTree" class="card-title text-center">Faker Tree</div>
      <vue-tree ref="tree" class="region" :dataset="data"
        :config="{ nodeWidth: 170, nodeHeight: 100, levelHeight: 150 }" :collapse-enabled="false" linkStyle="straight">
        <template v-slot:node="{ node }">
          <div class="rich-media-node" @click="openDialog(node, $event)">
            <span class="name" v-html="node.name"></span>
            <span v-if="node.custom" class="bi bi-person-fill-gear" title="customized node"></span>
          </div>
        </template>
      </vue-tree>
      <div class="info" style="">
        <router-link :to="{ path: '/docs/get-started/test-data' }">
          <span class="bi bi-question-circle-fill"></span>
        </router-link>
      </div>
    </div>
  </section>

  <div class="modal fade faker-dialog" tabindex="-1" ref="dialogRef" aria-hidden="true" aria-labelledby="dialog-title">
    <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <h6 id="dialog-title" class="modal-title">Node Information</h6>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body" v-if="selected">
          <div class="card-group">
            <div class="card">
              <div class="card-body">
                <div class="row mb-2">
                  <div class="col">
                    <p class="label">Name</p>
                    <p>{{ selected.name }}</p>
                  </div>
                  <div class="col">
                    <p class="label">Path</p>
                    <p>{{ selected.path }}</p>
                  </div>
                </div>
                <div class="row mb-2">
                  <div class="col">
                    <p class="label">Customized</p>
                    <p>{{ selected.custom ?? 'false' }}</p>
                  </div>
                  <div class="col" v-if="selected.weight">
                    <p class="label">Weight</p>
                    <p>{{ selected.weight }}</p>
                  </div>
                </div>
                <div class="row mb-2">
                  <div class="col">
                    <p class="label">Attributes</p>
                    <p>{{ selected.attributes?.map(x => x.name).join(', ') ?? '-' }}</p>
                  </div>
                  <div class="col">
                    <p class="label">Depends on</p>
                    <p>{{ selected.dependsOn?.map(x => x.name).join(', ') ?? '-' }}</p>
                  </div>
                </div>
                <div class="row mb-2" v-if="example">
                  <div class="col">
                    <p class="label">Example</p>
                    <pre class="hljs" v-if="!example.isLoading">
                      <code v-html="hljs.highlight(JSON.stringify(example.data), { language: 'javascript' }).value"></code>
                    </pre>
                    <span v-else>Loading example data...</span>
                    <button v-if="selected?.hasFakeFunc" type="button" class="btn btn-primary btn-sm mt-2" @click="getExample()">Next Example</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

</template>

<style>
.decisionTree .node-slot {
  cursor: default;
}
</style>
<style scoped>
.region {
  width: 100%;
  height: 800px;
}

.decisionTree {
  position: relative;
}

.decisionTree>.info {
  position: absolute;
  top: 10px;
  right: 10px;
  cursor: pointer;
  color: var(--link-color);
}

.decisionTree>.info:hover {
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
  cursor: pointer;
}

.rich-media-node span.name {
  margin: 0 auto;
  width: 80%;
  overflow-wrap: break-word;
}

.rich-media-node span.bi {
  position: absolute;
  top: 0;
  right: 5px;
}

.faker-dialog pre {
  display: block;
  max-width: 760px;
  margin: 0 auto auto;
  white-space: pre-wrap;
  word-break: break-all;
  box-shadow: none;
  line-height: 1.0;
  font-family: Menlo,Monaco,Consolas,"Courier New",monospace !important;
  font-size: 0.85rem;
  padding: 4px;
  padding-left: 12px;
}
.faker-dialog pre.hljs code {
    display: inline-block;
    width: 100%;
}
</style>