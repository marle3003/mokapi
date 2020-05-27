<template>
  <b-table small hover head-variant="dark" :items="resources" :fields="fields">
  </b-table>
</template>

<script>
export default {
    name: "resources",
    props: ['resources'],
    data() {
      return {
        fields: ['name', 'condition'],
      }
    }
}
</script>

<style scoped>

</style>