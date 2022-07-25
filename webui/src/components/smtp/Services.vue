<template>
  <b-card-group deck v-if="services !== null">
    <b-card
      body-class="info-body"
      class="text-center"
    >
      <b-card-title class="info">Smtp Servers</b-card-title>
      <b-table
        :items="services"
        :fields="fields"
        table-class="dataTable selectable"
        @row-clicked="clickHandler"
      >
        <template v-slot:cell(mails)="data">
          <span>{{ metric(data.item.metrics, 'smtp_mails_total') }}</span>
        </template>
      </b-table>
    </b-card>
  </b-card-group>
</template>

<script>
import Metrics from '@/mixins/Metrics'

export default {
  name: 'SmtpServices',
  mixins: [Metrics],
  props: ['services'],
  data () {
    return {
      fields: [
        { key: 'name', class: 'text-left' },
        'mails'
      ]
    }
  },
  methods: {
    clickHandler (record) {
      this.$router.push({ name: 'smtpMail', params: { id: record.id } })
    }
  }
}
</script>

<style scoped>
</style>
