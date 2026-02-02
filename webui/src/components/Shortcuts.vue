<script setup lang="ts">
import { inject, onMounted, onUnmounted } from 'vue';
import { Modal } from 'bootstrap';
import { useMarkdown } from '@/composables/markdown';

const files = inject<Record<string, string>>('files')!
const data = files['/src/assets/docs/release.md'];
const { content } = useMarkdown(data);

const shortcutHandler = (e: KeyboardEvent) => {
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || e.isComposing) {
      return
    }
    if (e.key === '?') {
      openDialog();
    }
}

onMounted(() => {
  window.addEventListener('keydown', shortcutHandler)
})
onUnmounted(() => {
  window.removeEventListener('keydown', shortcutHandler)
})

function openDialog() {
  const modalEl = document.getElementById('shortcuts')
  if (modalEl) {
    const modal = new Modal(modalEl)
    modal.show()
  }
}
</script>

<template>
  <div class="modal fade dialog-shortcuts" id="shortcuts" tabindex="-1" aria-hidden="true" aria-labelledby="shortcuts">
    <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content release">
        <div class="modal-header">
          <h6 id="shortcuts" class="modal-title">Help & Updates</h6>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <div class="nav card-tabs" role="tablist">
            <button class="active" id="tab-shortcuts" data-bs-toggle="tab" data-bs-target="#shortcuts-pane"
              type="button" role="tab" aria-controls="shortcuts-pane" :aria-selected="true">
              Shortcuts
            </button>
            <button id="tab-release" type="button" data-bs-toggle="tab" data-bs-target="#release-pane" role="tab"
              aria-controls="release-pane" :aria-selected="false">
              Release Notes
            </button>
          </div>
          <div class="tab-content">
            <div class="tab-pane fade show active" id="shortcuts-pane" role="tabpanel" aria-labelledby="shortcuts-tab">
              <div class="row">
                <div class="col-sm-6 mb-3 mb-sm-0">
                  <div class="card" aria-labelledby="search">
                    <div class="card-body">
                      <div class="card-header">Site-wide shortcuts</div>
                      <ul class="shortcuts">
                        <li>
                          <div>Bring up this help dialog</div>
                          <div>
                            <kbd>?</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Open search bar</div>
                          <div>
                            <kbd>/</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>d</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to HTTP dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>h</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to Kafka dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>k</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to LDAP dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>l</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to Mail dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>m</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to Jobs dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>j</kbd>
                          </div>
                        </li>
                        <li>
                          <div>Go to Configs dashboard</div>
                          <div>
                            <kbd>g</kbd>
                            <kbd>c</kbd>
                          </div>
                        </li>
                      </ul>
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="card" aria-labelledby="search">
                    <div class="card-header">Dashboard shortcuts</div>
                    <div class="card-body">
                      <ul class="shortcuts">
                        <li>
                          <div>Go to search</div>
                          <div>
                            <kbd>/</kbd>
                          </div>
                        </li>
                      </ul>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="tab-pane fade" id="release-pane" role="tabpanel" aria-labelledby="release-tab">
              <div v-html="content"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.dialog-shortcuts .card-header {
  font-size: 0.9rem;
  font-weight: 600;
}
.dialog-shortcuts .card-body {
  padding: 0;
}
.dialog-shortcuts .col-sm-6 {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}
.dialog-shortcuts .col-sm-6:first-child {
  padding-left: 0;
}
.dialog-shortcuts .col-sm-6:last-child {
  padding-right: 0;
}
.release h1 {
  margin-top: 0.5rem;
}

.shortcuts {
  padding-left: 0;
}

ul.shortcuts {
  font-size: 0.9rem;
  margin-top: 0;
  margin-bottom: 0;
}
.shortcuts li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
}
.shortcuts li+li {
    border-top: 1px solid var(--bs-card-border-color);
}
kbd {
  background-color: var(--card-background);
  color: var(--color-text);
  border: 1px solid var(--color-text);
}
kbd + kbd {
  margin-left: 0.3rem;
}
</style>