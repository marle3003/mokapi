<template>
<b-container fluid>
  <b-row class="mb-2">
    <b-col cols="auto" class="mr-auto pl-0">
    <h3>{{ $route.params.path }}</h3>
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
      <b-tab v-for="item in items" :key="item.method">
        <template v-slot:title>
        <b-badge pill :variant="item.variant" >{{ item.method }}</b-badge>
      </template>
        <b-card-text>
          <p class="label">Summary</p>
          <p>{{ item.summary }}</p>
          <p class="label">Description</p>
          <p>{{ item.description }}</p>
          <parameters v-bind:parameters="item.parameters" />
        </b-card-text>
      </b-tab>
    </b-tabs>
  </b-card>
  </b-row>
</b-container>
</template>

<script>
import Parameters from "@/components/Parameters"

export default {
    name: "endpoint",
    components: {},
    data() {
      return {
        items: [
          { method: "GET", summary: "Returns a list of users.", description: "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est"
          , variant:"primary",
          parameters: [{
            type: "path",
            name: "id",
            schema: "integer"
          }] }, 
          { method: "POST", summary: "Hello World", variant:"success" },
        ]
      }
    },
    components: {    
      "parameters": Parameters,
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