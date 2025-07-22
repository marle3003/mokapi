<script setup lang="ts">
import { useMails } from '@/composables/mails'
import { onMounted, ref, watch } from 'vue'

const props = defineProps({
    messageId: { type: String, required: true },
    body: { type: String, required: true },
    contentType: { type: String, required: true },
  }
)

const { attachmentUrl } = useMails()
const isDark = ref<boolean>( document.documentElement.getAttribute('data-theme') == 'dark')
const mailBodyTheme = localStorage.getItem('mail-body-theme')
if (mailBodyTheme === 'light') {
  isDark.value = false
}

function parseHtmlBody(body: string) {
    body = body.replace(/cid:(?<name>[^"]*)/g, function(_: string, name: string){
      return document.baseURI + attachmentUrl(props.messageId, name).substring(1)
    })
    body = `<style>:host{ all: initial; }</style>${body}`

    return body
}

let shadowRoot: ShadowRoot;
onMounted(() => {
  const body = document.querySelector('#body')
  if (body) {
    shadowRoot = body.attachShadow({mode: 'open'});
    setBody()
  }
})

watch(isDark, setBody)

function setBody() {
  shadowRoot.innerHTML = parseHtmlBody(props.body)

  for (const sheet of shadowRoot.styleSheets) {
    for (const rule of sheet.cssRules) {
      const r: any = rule
      if (r.media?.mediaText === '(prefers-color-scheme: dark)') {
        r.media.mediaText = isDark.value ? '' : '(original-prefers-color-scheme)'
      }
      if (r.media?.mediaText === '(prefers-color-scheme: light)') {
        r.media.mediaText = isDark.value ? '(original-prefers-color-scheme)' : ''
      }
    }
  }
}

function switchTheme() {
  const body = document.getElementById('mail-body')!
  isDark.value = !isDark.value
  const theme = isDark.value ? 'dark' : 'light'
  localStorage.setItem('mail-body-theme', theme)
  
}
</script>

<template>
  <div class="card-group mail-body" id="mail-body" :data-theme="isDark ? 'dark' : 'light'">
    <div class="card">
      <div class="card-body">
        <div class="float-end tools">
          <i class="bi bi-brightness-high-fill" @click="switchTheme" v-if="isDark"></i>
          <i class="bi bi-moon-fill" @click="switchTheme" v-if="!isDark"></i>
        </div>
        <div class="row">
          <p class="col">
            <div id="body" v-if="contentType.startsWith('text/html')" style="display: inline-block; width: 100%;" :style="isDark ? 'color: #fff;' : 'color: #000;'"></div>
            <p v-else>{{ body }}</p>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .bodyLabel {
    position: absolute;
    left: 50%;
  }
  .dashboard .card-group .mail-body[data-theme="light"] > .card {
    background-color: white;
  }

  .dashboard .card-group.mail-body p {
    padding: 0;
    margin: 0;
  }
  .tools i {
    cursor: pointer;
  }
</style>

