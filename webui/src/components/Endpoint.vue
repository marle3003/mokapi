<template>
  <b-row class="ml-4 mr-4">
    <b-col>
      <b-container fluid v-if="endpoint != null">
        <b-row class="mb-2">
          <b-col cols="auto" class="mr-auto pl-0">
          <h3>{{ endpoint.path }}</h3>
          </b-col>
          <b-col cols="auto" class="pr-0">
          <div class="close" @click="$router.go(-1)">
            <b-icon icon="x" class="border rounded p-1"></b-icon>
          </div>
          </b-col>
        </b-row>
        <b-row>
        <b-card no-body class="w-100">
          <b-tabs card>
            <b-tab v-for="operation in endpoint.operations" :key="operation.method">
              <template v-slot:title>
              <b-badge pill class="operation" :class="operation.method" >{{ operation.method }}</b-badge>
              </template>
              <b-card-text>
                <p class="label">Summary</p>
                <p>{{ operation.summary }}</p>
                <p class="label">Description</p>
                <p>{{ operation.description }}</p>
                <h5>Parameters</h5>
                <parameters v-bind:parameters="operation.parameters" />
                <hr v-if="operation.requestBody != null" />
                <h5 v-if="operation.requestBody != null">Request Body</h5>
                
                <hr />
                <h5>Response</h5>
                <response v-bind:responses="operation.responses" />
                
                <hr v-if="operation.middlewares != null" />
                <h5 v-if="operation.middlewares != null">Middlewares</h5>
                <middlewares v-if="operation.middlewares != null" v-bind:middlewares="operation.middlewares" />

                <hr v-if="operation.resources != null" />
                <h5 v-if="operation.resources != null">Resources</h5>
                <resources v-if="operation.resources != null" v-bind:resources="operation.resources" />
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
import Parameters from '@/components/Parameters'
import Schema from '@/components/Schema'
import Response from '@/components/Response'
import Resources from '@/components/Resources'
import Middlewares from '@/components/Middlewares'

export default {
    name: "endpoint",
    props: ['service'],
    computed: {
      endpoint: function(){
        if (this.service == null){
          return null;
        }

        let path = this.$route.params.path;

        for (let i = 0; i < this.service.endpoints.length; i++){
          let endpoint = this.service.endpoints[i]
          if (endpoint.path === path){
            return endpoint
          }
        }

        return null
      }
    },
    components: {    
      'parameters': Parameters,
      'response': Response,
      'resources': Resources,
      'middlewares': Middlewares
    },
    methods:{
      doCommand(e) {
          let cmd = e.key.toLowerCase();
          if (cmd == "escape"){
            this.$router.go(-1)
          }
        }
    },
    created() {
      window.addEventListener('keyup', this.doCommand);
    },
    destroyed() {
      window.removeEventListener('keyup', this.doCommand);
    },
}
</script>

<style scoped>
.label{
    color: #a0a1a7;
    margin-bottom: 0
  }
.close {
  font-size: 2.2rem;
  cursor: pointer;
}
</style>