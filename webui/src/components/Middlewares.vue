<template>
  <b-table small hover head-variant="dark" :items="middlewares">
  </b-table>
</template>

<script>
export default {
    name: "middlewares",
    props: ['middlewares'],
}
</script>

<style scoped>

</style>