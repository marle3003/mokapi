<template>
  <b-table small hover head-variant="dark" :items="parameters" :fields="fields">
    <template v-slot:cell(show_details)="row">
        <div @click="row.toggleDetails">
            <b-icon v-if="row.detailsShowing" icon="dash-square"></b-icon>
            <b-icon v-else icon="plus-square"></b-icon>
          </div>
      </template>
     <template v-slot:row-details="row">
        <b-card>
          <b-row class="mb-2">
            <b-col sm="3" class="text-sm-right"><b>Description:</b></b-col>
            <b-col><vue-simple-markdown :source="row.item.description" /></b-col>
          </b-row>
        </b-card>
      </template>
  </b-table>
</template>

<script>
export default {
    name: "parameters",
    props: ['operation'],
    data() {
      return {
        fields: [{key: 'show_details', label: ''}, 'name', 'in', 'type'],
      }
    },
    computed: {
      parameters: function () {
        if (this.operation == null){
          return [];
        }

        let result = [];

        this.operation.parameters.forEach((parameter, index) => {
          result.push(parameter);
        });
        
        return result;
      }
    }
}
</script>

<style scoped>

</style>