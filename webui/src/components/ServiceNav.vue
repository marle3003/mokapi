<template>
  <b-nav card-header tabs>
    <b-nav-item>
        <img src="@/assets/endpoint.png" width="24" height="24" /> Endpoints
    </b-nav-item>
  </b-nav>
</template>

<script>
export default {
  name: 'serviceNav',
  props: ['service'],
  methods: {
    routerLinkToEndpoints (item, index, event) {
      this.$router.push({ name: 'endpoints' })
    },
    routerLinkToModels (item, index, event) {
      this.$router.push({ name: 'models' })
    }
  }
}
</script>

<style>
  .operations{
    cursor: pointer;
  }
</style>

<style scoped>
  .card-header-tabs{
      font-size: 1.2rem;
  }
</style>
