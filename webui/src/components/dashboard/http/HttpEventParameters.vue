<script setup lang="ts">
import { Modal } from 'bootstrap';
import { computed, ref, useTemplateRef, type PropType } from 'vue';
import SourceView from '../SourceView.vue';

const props = defineProps({
    parameters: { type: Object as PropType<HttpEventParameter[]>, required: true },
})
const sorted = computed(() => {
    const result = props.parameters.sort((p1, p2) => {
        if (p1.value && p2.value) {
            return p1.name.localeCompare(p2.name)
        }
        if (p1.value) {
            return -1
        }
        return 1
    });
    return result.map(p => ({
        ...p,
        rendered: renderJsonValue(p.value)
    }))
})
const showRaw = ref<{[name: string]: boolean}>({})
const dialogShowRaw = ref(false)
const selected = ref<HttpEventParameter | undefined>()
const dialogRef = useTemplateRef('dialogRef')
const dialog = ref<Modal | undefined>()
function renderJsonValue(value: any) {
    try {
        const parsed = typeof value === 'string' ? JSON.parse(value) : value;
        if (typeof parsed === 'string') {
            return parsed;
        }
        return JSON.stringify(parsed, null, 2);
    } catch {
        return value
    }
}
const useValueSwitcher = computed(() => {
    for (const p of sorted.value) {
        if (!p.value) {
            continue;
        }
        if (p.rendered != p.raw) {
            return true
        }
    }
    return false;
})
function openDialog(p: HttpEventParameter) {
  selected.value = p
  dialogShowRaw.value = false
  if (!dialog.value) {
    dialog.value = new Modal(dialogRef.value!)
  }
  dialog.value.show()
}
</script>

<template>
    <table class="table table.sm dataTable">
        <thead>
            <tr>
                <th scope="col" style="width:40px" v-if="useValueSwitcher"></th>
                <th scope="col" class="text-left w-20">Name</th>
                <th scope="col" class="text-left" style="width:100px;">Type</th>
                <th scope="col" class="text-center" style="width: 130px;">OpenAPI</th>
                <th scope="col" class="text-left" style="width:70%">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="p in sorted"  @click="openDialog(p)" class="table-row-clickable">
                <td v-if="useValueSwitcher">
                    <button v-if="p.value && p.rendered !== p.raw" class="btn btn-sm btn-outline-secondary" style="--bs-btn-padding-y: .1rem; --bs-btn-padding-x: .25rem; --bs-btn-font-size: .75rem;"
                        @click.stop="showRaw[p.name] = !showRaw[p.name]">
                        <i v-if="showRaw[p.name]" class="bi bi-layout-text-sidebar" title="Show parsed value"></i>
                        <i v-else class="bi bi-code" title="Show raw value"></i>
                    </button>
                </td>
                <td class="align-middle">{{ p.name }}</td>
                <td class="align-middle">{{ p.type }}</td>
                <td class="text-center align-middle">{{ p.value ? 'yes' : 'no' }}</td>
                <td class="align-middle text-truncate">{{ p.value ? (showRaw[p.name] ? p.raw : p.rendered) : p.raw }}</td>
            </tr>
        </tbody>
    </table>
    <!-- Modal -->
    <div class="modal fade" ref="dialogRef" tabindex="-1">
      <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ selected?.name }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row mb-3">
                <div class="col">
                    <p class="label">Type</p>
                    <p>{{ selected?.type }}</p>
                </div>
                <div class="col">
                    <p class="label">OpenAPI</p>
                    <p>{{ selected?.value ? 'yes' : 'no' }}</p>
                </div>
            </div>
            <div class="row" v-if="!selected?.value">
                <div class="col">
                    <p class="label">Value</p>
                    <p>{{ selected?.raw }}</p>
                </div>
            </div>
            <div class="row" v-if="selected?.value">
                <div class="col">
                    <section>
                        <div class="header">
                            <div data-v-d7e2f089="" class="view controls">
                                <button type="button" class="btn btn-link" :class="{ active: !dialogShowRaw }" @click="dialogShowRaw = false">Value</button>
                                <button type="button" class="btn btn-link" :class="{ active: dialogShowRaw }" @click="dialogShowRaw = true">Raw</button>
                            </div>
                        </div>
                        <div class="body">
                            {{ dialogShowRaw || !selected?.value ? selected?.raw : renderJsonValue(selected?.value) }}
                        </div>
                    </section>
                </div>
            </div>
          </div>
        </div>
      </div>
    </div>
</template>

<style scoped>
.w-10{
    width: 10%;
}
.w-20{
    width: 20%;
}
.table-row-clickable {
    cursor: pointer;
}
.header {
    border: 1px solid var(--source-border);
    border-radius: 6px 6px 0 0;
    color: var(--color-text-light);
    display: flex;
    padding: 6px;
}
.header .controls {
    border: 1px solid var(--source-border);
    border-radius: 6px;
}
.header .controls > button {
    font-size: 0.9rem;
    vertical-align: middle;
    color: var(--source-header-color);
    display: inline-grid;
    place-content: center;
    border-right: 1px solid var(--source-border);
}
.header .controls > button {
    height: 28px;
    line-height: 18px;
    position: relative;
    text-decoration: none;
    border-top-right-radius: 0 !important;
    border-bottom-right-radius: 0 !important;
}
.header .controls > button.active {
    /* background-color: var(--color-button-link-active);
    color: var(--color-button-text-hover); */
    background-color: var(--color-button-link-active);
    color: var(--color-button-text-hover);
    outline: 1px solid var(--source-border);
    border-radius: 6px;
}
.header .controls > button + button {
    border-top-left-radius: 0 !important;
    border-bottom-left-radius: 0 !important;
}
.header .controls > *:last-child {
    border-right: 0;
    border-top-right-radius: 6px !important;
    border-bottom-right-radius: 6px !important;
}
.body {
    border: 1px solid var(--source-border);
    border-top: 0;
    border-radius: 0 0 6px 6px;
    padding: 8px;
}
</style>