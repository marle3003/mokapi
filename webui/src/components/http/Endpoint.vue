<template>
  <b-row v-if="pathItem != null">
    <b-col>
      <b-container fluid>
        <b-row class="mb-2">
          <b-col cols="auto" class="mr-auto pl-0">
          <h3>{{ pathItem.path }}</h3>
          </b-col>
          <b-col cols="auto" class="pr-0">
          <div class="close" @click="$router.go(-1)">
            <b-icon icon="x" class="border rounded p-1"></b-icon>
          </div>
          </b-col>
        </b-row>
        <b-row>
          <b-card class="w-100 mb-2">
            <b-row>
              <b-col class="col-image">
                <img src="@/assets/endpoint.png" />
              </b-col>
              <b-col>
                <p class="label">Summary</p>
                <p class="title">{{ pathItem.summary }}</p>
              </b-col>
              </b-row>
              <b-row>
                <b-col class="col-image"></b-col>
              <b-col>
                <p class="label">Description</p>
                <vue-simple-markdown :source="pathItem.description" />
              </b-col>
            </b-row>
          </b-card>
        </b-row>
        <b-row>
          <b-card no-body class="w-100">
            <b-tabs card>
              <b-tab v-for="operation in pathItem.operations" :key="operation.method">
                <template v-slot:title>
                <b-badge pill class="operation" :class="operation.method" >{{ operation.method }}</b-badge>
                </template>
                <b-card-text>
                  <p class="label">Summary</p>
                  <p>{{ operation.summary }}</p>
                  <p class="label">Description</p>
                  <p><vue-simple-markdown :source="operation.description" /></p>

                  <h2>Parameters</h2>
                  <parameters v-bind:operation="operation" />

                  <requestBody v-bind:operation="operation" />

                  <hr />
                  <h2>Response</h2>
                  <response v-bind:responses="operation.responses" />
                </b-card-text>
              </b-tab>
            </b-tabs>
          </b-card>
        </b-row>
      </b-container>
    </b-col>
  </b-row>
</template>

<script>
import Parameters from '@/components/http/Parameters'
import Response from '@/components/http/Response'
import RequestBody from '@/components/http/RequestBody'

export default {
  name: 'endpoint',
  props: ['service'],
  computed: {
    pathItem: function () {
      if (this.service == null) {
        return null
      }

      let path = this.$route.params.path
      path = decodeURIComponent(path)

      for (let i = 0; i < this.service.paths.length; i++) {
        let pathItem = this.service.paths[i]
        if (pathItem.path === path) {
          return pathItem
        }
      }

      return null
    }
  },
  components: {
    'parameters': Parameters,
    'response': Response,
    'requestBody': RequestBody
  },
  methods: {
    doCommand (e) {
      let cmd = e.key.toLowerCase()
      if (cmd === 'escape') {
        this.$router.go(-1)
      }
    }
  },
  created () {
    window.addEventListener('keyup', this.doCommand)
  },
  destroyed () {
    window.removeEventListener('keyup', this.doCommand)
  }
}
</script>

<style scoped>
.close {
  font-size: 2.2rem;
  cursor: pointer;
  border-color: var(--var-border-color);
  color: var(--var-color-primary);
}
.col-image{
    width: 100px;
    flex: 0 0 auto
  }
</style>
<style>

</style>
