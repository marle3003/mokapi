<template>
  <b-card-group deck>
    <b-card class="w-100">
      <b-card-title class="info text-center">Recent Mails</b-card-title>
      <b-table
        hover
        :items="mails"
        :fields="mailFields"
        class="dataTable selectable"
        @row-clicked="mailClickHandler"
      >
        <template v-slot:cell(from)="data">
                <div
                  v-for="from in data.item.from"
                  :key="from.Address"
                >
                  <span v-if="from.Name !== ''">{{ from.Name }} &lt;</span><span>{{ from.Address }}</span><span v-if="from.Name !== ''">&gt;</span>
                </div>
              </template>
              <template v-slot:cell(to)="data">
                <div
                  v-for="to in data.item.to"
                  :key="to.Address"
                >
                  <span v-if="to.Name !== ''">{{ to.Name }} &lt;</span><span>{{ to.Address }}</span><span v-if="to.Name !== ''">&gt;</span>
                </div>
              </template>
              <template v-slot:cell(time)="data">
                {{ data.item.time | moment}}
              </template>
      </b-table>
    </b-card>
  </b-card-group>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'

export default {
  mixins: [Api, Filters, Refresh],
  data () {
    return {
      mails: [],
      mailFields: [
        'from',
        'to',
        { key: 'subject', class: 'subject' },
        'time'
      ],
    }
  },
  methods: {
    async getData () {
      this.$http.get(this.baseUrl + '/api/events?namespace=smtp').then(
        r => {
          this.mails = r.data
        },
        r => {
          this.mails = []
        }
      )
    },
    mailClickHandler (record) {
      this.$router.push({ name: 'smptMail', params: { id: record.id } })
    }
  }
}
</script>

<style scoped>
.card {
  border-color: var(--var-border-color);
  margin: 7px;
}
.info {
  font-size: 0.7rem;
  font-weight: 300;
}
</style>
