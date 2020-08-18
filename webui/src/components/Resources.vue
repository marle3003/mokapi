<template>
  <b-table small hover head-variant="dark" :items="resources" :fields="fields">
    <template v-slot:cell(status)="data">
      <b-icon icon="check-circle" class="mr-1" variant="success" v-if="data.item.status.length == 0"></b-icon>
      <span v-else>{{ data.item.status }}</span>
    </template>
  </b-table>
</template>

<script>
export default {
    name: "resources",
    props: ['resources'],
    data() {
      return {
        fields: ['name', 'condition', 'status'],
      }
    }
}
</script>

<style scoped>

</style>