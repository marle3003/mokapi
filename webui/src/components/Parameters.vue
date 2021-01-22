<template>
  <b-table small hover head-variant="dark" :items="parameters" :fields="fields">
    <template v-slot:cell(show_details)="row">
        <div @click="toggleDetails(row)">
            <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
            <b-icon v-else icon="plus-square"></b-icon>
          </div>
      </template>
     <template v-slot:row-details="row">
        <b-card class="w-100">
          <b-row class="mb-2">
            <b-col sm="3" class="text-sm-right"><b>Description:</b></b-col>
            <b-col><vue-simple-markdown :source="row.item.description" /></b-col>
          </b-row>
          <b-row class="mb-2">
            <b-col sm="3" class="text-sm-right"><b>Schema:</b></b-col>
            <b-col><schema v-bind:schema="row.item.schema"></schema></b-col>
          </b-row>
        </b-card>
      </template>
      <template v-slot:cell(schema)="data">
      {{ data.item.schema ? data.item.schema.type : 'undefined' }}
    </template>
  </b-table>
</template>

<script>
import Schema from '@/components/Schema'

export default {
  name: 'parameters',
  components: {'schema': Schema,},
  props: ['operation'],
  data () {
    return {
      fields: [{key: 'show_details', label: ''}, 'name', 'location', 'schema'],
      detailsShown: []
    }
  },
  computed: {
    parameters: function () {
      if (this.operation === null) {
        return []
      }

      let result = []

      if (this.operation.parameters) {
        this.operation.parameters.forEach((parameter, index) => {
          parameter['key'] = parameter.name +":"+parameter.location
          parameter['_showDetails'] = this.detailsShown.indexOf(parameter.key) >= 0
          result.push(parameter)
        })
      }

      return result
    }
  },
  methods: {
    toggleDetails(row) {
      row.toggleDetails()
      const index = this.detailsShown.indexOf(row.item.key);

      if (row.item._showDetails) {
        this.detailsShown.push(row.item.key)
      } else {
      	this.detailsShown.splice(index, 1)
      }
    }
  }
}
</script>

<style scoped>

</style>
