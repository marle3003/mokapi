<template>
  <b-card-group deck>
    <b-card class="w-100">
      <b-card-title class="info text-center">Recent Requests</b-card-title>
      <b-table
        hover
        :items="requests"
        :fields="requestFields"
        class="dataTable selectable"
        @row-clicked="requestClickHandler"
      >
        <template v-slot:cell(method)="data">
          <b-badge
            pill
            class="operation"
            :class="data.item.data.request.method.toLowerCase()"
          >{{ data.item.data.request.method }}</b-badge>
        </template>
        <template v-slot:cell(statusCode)="data">
          <b-icon
            icon="circle-fill"
            class="response icon mr-1"
            variant="success"
            v-if="data.item.data.response.statusCode >= 200 && data.item.data.response.statusCode < 300"
          ></b-icon>
          <b-icon
            icon="circle-fill"
            class="response icon mr-1"
            variant="warning"
            v-if="data.item.data.response.statusCode >= 300 && data.item.data.response.statusCode < 400"
          ></b-icon>
          <b-icon
            icon="circle-fill"
            class="response icon mr-1 client-error"
            v-if="data.item.data.response.statusCode >= 400 && data.item.data.response.statusCode < 500"
          ></b-icon>
          <b-icon
            icon="circle-fill"
            class="response icon mr-1"
            variant="danger"
            v-if="data.item.data.response.statusCode >= 500 && data.item.response.statusCode < 600"
          ></b-icon>
          {{ data.item.data.response.statusCode }}
        </template>
        <template v-slot:cell(url)="data">
          {{ data.item.data.request.url }}
        </template>
        <template v-slot:cell(time)="data">
          {{ data.item.time | moment }}
        </template>
        <template v-slot:cell(duration)="data">
          {{ data.item.data.duration | duration }}
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
      requests: [],
      requestFields: [
        'method',
        'statusCode',
        { key: 'url', tdClass: 'break', label: 'Url' },
        'time',
        'duration'
      ]
    }
  },
  methods: {
    async getData () {
      this.$http.get(this.baseUrl + '/api/events?namespace=http').then(
        r => {
          this.requests = r.data
        },
        r => {
          this.requests = []
        }
      )
    },
    requestClickHandler (record) {
      this.$router.push({ name: 'httpRequest', params: { id: record.id } })
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
