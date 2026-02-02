<script setup lang="ts">
import { useGuid } from '@/composables/guid';
import { useMarkdown } from '@/composables/markdown';
import { computed, inject, ref, watch } from 'vue';
import { Modal } from 'bootstrap';
import { useLocalStorage } from '@/composables/local-storage';
import { useDashboard } from '@/composables/dashboard';
import type { AppInfoResponse } from '@/types/dashboard';
import { useRoute } from '@/router';

const route = useRoute();
const files = inject<Record<string, string>>('files')!
const { createGuid } = useGuid();
const id = createGuid()
const { dashboard } = useDashboard()
let appInfo = ref<AppInfoResponse | null>(null)

const data = files['/src/assets/docs/release.md'];
const { content } = useMarkdown(data);
const dismissedVersion = useLocalStorage<string>(`release-notes-version`, '')
const dismissed = useLocalStorage<boolean>(`release-notes-dismissed`, false)
const displayed = ref(false)

const version = computed(() => {
  if (!appInfo.value?.data) {
    return undefined;
  }
  return appInfo.value.data.version
})

watch(appInfo, () => {
  if (!appInfo.value || displayed.value) {
    return;
  }
  if (appInfo.value.isLoading) {
    return
  }

  if (!version.value) {
    return
  }
  if (dismissedVersion.value === version.value && dismissed.value) {
    return
  }

  if (getItemWithExpiry("release-notes-hide")) {
    return
  }
  if (import.meta.env.DEV) {
    if (!Object.keys(route.query).find(x => x === 'show-release-notes')) {
      return
    }
  }
  if (!content) {
    return
  }

  const modalEl = document.getElementById(id)
  if (modalEl) {
    const modal = new Modal(modalEl)
    modal.show()
    displayed.value = true
  }
}, { deep: true })

watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getAppInfo();
    appInfo.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);

function dismiss() {
  dismissed.value = true;
  dismissedVersion.value = version.value!
}

function close() {
  setItemWithExpiry("release-notes-hide", "true", 12);
}

function setItemWithExpiry(key: string, value: string, hours: number) {
  const now = new Date();
  const item = {
    value,
    expiry: now.getTime() + hours * 60 * 60 * 1000, // expiration timestamp
  };
  localStorage.setItem(key, JSON.stringify(item));
}

function getItemWithExpiry(key: string): string | null {
  const itemStr = localStorage.getItem(key);
  if (!itemStr) return null;

  const item = JSON.parse(itemStr);
  const now = new Date();

  if (now.getTime() > item.expiry) {
    // Item has expired
    localStorage.removeItem(key);
    return null;
  }
  return item.value;
}
</script>

<template>
  <div class="modal fade" :id="id" tabindex="-1" aria-hidden="true" :aria-labelledby="id + 'title'">
    <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content release">
        <div class="modal-header visually-hidden">
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body" v-html="content"></div>
        <div class="modal-footer">
          <button type="button" class="btn btn-sm btn-outline-secondary" data-bs-dismiss="modal" @click="dismiss">Dismiss for this
            release</button>
          <button type="button" class="btn btn-sm btn-primary" data-bs-dismiss="modal" @click="close">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.release h1 {
  margin-top: 0.5rem;
}
</style>