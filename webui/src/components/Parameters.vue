<template>
  <b-table small hover head-variant="dark" :items="parameters" :fields="fields">
  </b-table>
</template>

<script>
export default {
    name: "parameters",
    props: ['parameters'],
    data() {
      return {
        fields: ['name', 'in', 'type'],
      }
    }
}
</script>

<style scoped>

</style>