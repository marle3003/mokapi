<template>
  <b-row class="ml-4 mr-4">
    <b-col>
    <operations :service="service"></operations>
    </b-col>
  </b-row>
</template>

<script>
import Operations from '@/components/Endpoints'
import Models from '@/components/Models'

export default {
    name: 'serviceOverview',
    props: ["service"],
    components: {   
      'operations': Operations,
      'models': Models
    }
}
</script>