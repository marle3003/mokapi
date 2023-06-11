<script setup lang="ts">
import { useMails } from '@/composables/mails'
import { onMounted, ref } from 'vue'

const props = defineProps({
    messageId: { type: String, required: true },
    body: { type: String, required: true },
    contentType: { type: String, required: true },
  }
)

const { attachmentUrl } = useMails()

const showBodyLabel = ref<boolean>(false)

function parseHtmlBody(body: string) {
    body = body.replace(/cid:(?<name>[^"]*)/g, function(_: string, name: string){
      return document.baseURI + attachmentUrl(props.messageId, name).substring(1)
    })
    return body
}

onMounted(() => {
  const body = document.querySelector('#body')
  if (body) {
    const shadowRoot = body.attachShadow({mode: 'open'});
    shadowRoot.innerHTML = parseHtmlBody(props.body)
  }
})
</script>

<template>
  <div class="card-group" @mouseover="showBodyLabel = true" @mouseleave="showBodyLabel = false">
    <div class="card">
      <div class="card-body">
        <div class="card-title text-center bodyLabel" :style="[showBodyLabel ? '' : 'display: none']">Body</div>
        <div class="row">
          <p class="col">
            <div id="body" v-if="contentType.startsWith('text/html')" style="display: inline-block;width: 100%"></div>
            <!-- <iframe frameborder="0"  style="height: 100vh; width:100vh; border: none" v-if="conten     tType.startsWith('text/html')" :src="'data:'+contentType+','+parseHtmlBody(body)"></iframe> -->
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
</style>

