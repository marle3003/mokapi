<template>
<b-card no-body>
    <b-tabs pills card vertical class="service" content-class="tabs">
      <b-tab title-link-class="serviceNav">
        <template #title>
          <img src="@/assets/endpoint.png" width="24" height="24" /> Endpoints
        </template>
        <b-card-text><endpoints :service="service"></endpoints></b-card-text>
      </b-tab>
      <b-tab title-link-class="serviceNav">
        <template #title>
          <img src="@/assets/models.png" width="24" height="24" /> Models
        </template>
        <b-card-text><models :service="service"></models></b-card-text>
      </b-tab> 
    </b-tabs>
  </b-card>
</template>

<script>
import Endpoints from '@/components/Endpoints'
import Models from '@/components/Models'

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
.serviceNav.nav-link.active{
  background-color: rgb(0,0,0,0.1);
color: #212529;
}
.serviceNav.nav-link{
  color: #212529;
  font-weight: 500;
}
.service .nav-item{

  border-bottom: solid;
  border-width: 2px;
  border-color: rgb(222, 226, 230)
}
.service .nav {
  padding:0 0 0 0;
}
.tabs .tab-pane{
  padding-top: 0;
}
</style>