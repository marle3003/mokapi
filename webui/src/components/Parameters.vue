<template>
  <b-table striped hover :items="parameters" :fields="fields">
  </b-table>
</template>

<script>
export default {
    name: "parameters",
    props: ['parameters'],
    data() {
      return {
        fields: ['name', 'type', 'schema'],
      }
    }
}
</script>

<style scoped>

</style>