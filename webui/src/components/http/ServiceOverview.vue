<template>
<b-card no-body>
    <b-tabs pills card vertical class="service" content-class="tabs">
      <b-tab title-link-class="serviceNav">
        <template #title>
          <img src="@/assets/endpoint.png" width="24" height="24" />
        </template>
        <b-card-text class="endpoints"><endpoints :service="service"></endpoints></b-card-text>
      </b-tab>
      <b-tab title-link-class="serviceNav">
        <template #title>
          <img src="@/assets/models.png" width="24" height="24" />
        </template>
        <b-card-text><models :service="service"></models></b-card-text>
      </b-tab>
    </b-tabs>
  </b-card>
</template>

<script>
import Endpoints from '@/components/http/Endpoints'
import Models from '@/components/http/Models'

export default {
  name: 'serviceOverview',
  props: ['service'],
  components: {
    'endpoints': Endpoints,
    'models': Models
  }
}
</script>

<style>
.serviceNav.nav-link{
  font-weight: 500;
}
.service .nav {
  padding:0 0 0 0;
}
</style>
