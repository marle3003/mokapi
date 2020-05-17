<template>
  <b-tabs content-class="mt-3 ml-2" class="responses" align="left">
    <b-tab v-for="response in sorted" :key="response.status">
      <template v-slot:title>
        <b-icon icon="circle-fill" class="icon mr-1" variant="success" v-if="response.status >= 200 && response.status < 300"></b-icon>
        <b-icon icon="circle-fill" class="icon mr-1" variant="warning" v-if="response.status >= 300 && response.status < 400"></b-icon>
        <b-icon icon="circle-fill" class="icon mr-1 client-error" v-if="response.status >= 400 && response.status < 500"></b-icon>
        <b-icon icon="circle-fill" class="icon mr-1" variant="danger" v-if="response.status >= 500 && response.status < 600"></b-icon>
        {{ response.status }}
      </template>
      <p class="label">Description</p>
      <p>{{ response.description }}</p>
      <div v-if="response.contentTypes != null && response.contentTypes.length === 1">
          <div v-for="content in response.contentTypes" :key="content.type">
            <p class="label">Content Type</p>
            <p>{{ content.type }}</p>
            <p v-if="content.schema != null" class="label">Schema</p>
            <schema v-if="content.schema != null" v-bind:schema="content.schema"></schema>
          </div>
      </div>
      <b-card no-body v-if="response.contentTypes != null && response.contentTypes.length > 1">
        <b-tabs pills card vertical nav-class="p-0">
          <b-tab v-for="content in response.contentTypes" :key="content.type" :title="content.type" >
            <b-card-text>
              <p class="label">Schema</p>
              <schema v-bind:schema="content.schema"></schema>
            </b-card-text>
          </b-tab>
        </b-tabs>
      </b-card>
    </b-tab>
  </b-tabs>
</template>

<script>
import Schema from "@/components/Schema"

export default {
    name: "response",
    components: { 'schema': Schema,},
    props: ['responses'],
    computed: {
      sorted: function(){
        if (this.responses == null){
          return [];
        }
        
        function compare(a, b) {
          return a.status - b.status;
        }

        return  this.responses.sort(compare);
      }
    }
}
</script>

<style scoped>
.responses .icon{
    vertical-align: middle;
    font-size: 0.6rem;
}
.client-error{
    color: var(--orange);
}
.label{
    color: #a0a1a7;
    margin-bottom: 0
  }
</style>