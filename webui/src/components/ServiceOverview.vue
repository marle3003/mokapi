<template>
<div>
  <b-row class="mb-4">
    <b-col>
    <operations :service="service"></operations>
    </b-col>
  </b-row>
  <b-row>
    <b-col>
    <models :service="service"></models>
    </b-col>
  </b-row>
</div>
</template>

<script>
import Operations from '@/components/Endpoints'
import Models from '@/components/Models'

export default {
  name: 'serviceOverview',
  props: ['service'],
  components: {
    'operations': Operations,
    'models': Models
  }
}
</script>
