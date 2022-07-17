<template>
  <div v-if="server !== null">
    <b-card-group deck>
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p class="label">Smtp Server</p>
            <p style="font-size: 1.25rem; font-weight: 500;">{{ server.name }}</p>
          </b-col>
          <b-col>
            <p class="label">Version</p>
            <p>{{ server.version }}</p>
          </b-col>
        </b-row>
        <b-row>

        </b-row>
        <b-row v-if="server.description !== ''">
          <b-col>
            <p class="label">Description</p>
            <vue-simple-markdown :source="server.description" />
          </b-col>
        </b-row>
      </b-card>
    </b-card-group>
    <b-card-group deck>
      <b-card
        body-class="info-body"
        class="text-center"
      >
        <b-card-title class="info">Mails</b-card-title>
        <b-card-text class="text-center value">{{ metric(cluster.metrics, 'smtp_mails_total') }}</b-card-text>
      </b-card>
    </b-card-group>

    <b-card-group deck>
      <b-card class="w-100">
        <b-card-title class="info text-center">Mails</b-card-title>
        <b-table
          small
          hover
          class="dataTable selectable"
          :items="mails"
          :fields="mailFields"
          style="table-layout: fixed"
          @row-clicked="mailClickHandler"
        >
        <template v-slot:cell(partitions)="data">
            <span>{{ data.item.partitions.map(x => x.id).join(", ") }}</span>
          </template>
          <template v-slot:cell(messages)="data">
            {{ metric(cluster.metrics, 'kafka_messages_total', {name: 'topic', value: data.item.name}) }}
          </template>
          <template v-slot:cell(time)="data">
            <span>{{ metric(cluster.metrics, 'kafka_message_timestamp', {name: 'topic', value: data.item.name}) | moment }}</span>
          </template>
        </b-table>
      </b-card>
    </b-card-group>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Filters from '@/mixins/Filters'
import Refresh from '@/mixins/Refresh'
import Metrics from '@/mixins/Metrics'
import Shortcut from '@/mixins/Shortcut'

import Header from '@/components/dashboard/Header'

export default {
  name: 'Topic',
  mixins: [Api, Filters, Refresh, Metrics, Shortcut],
    components: {
    'dashboard-header': Header
  },
  data () {
    return {
      cluster: null,
      mailFields: [
        'subject',
        'from',
        'to',
        'time'
      ]
    }
  },
  methods: {
    async getData () {
      const name = this.$route.params.cluster
      if (!name){
        return
      }

      this.$http.get(this.baseUrl + '/api/services/kafka/' + name).then(
        r => {
          this.cluster = r.data
        },
        r => {
          this.cluster = null
        }
      )
    },
    topicClickHandler (item) {
      this.$router.push({
        name: 'kafkaTopic',
        params: { cluster: this.$route.params.cluster, topic: item.name },
        query: { refresh: '5' }
      })
    },
    compareTopic (x, y) {
      return x.name.localeCompare(y.name)
    }
  }
}
</script>

<style scoped>
.dashboard {
  width: 90%;
  margin: 42px auto auto;
}
.page-header h2 {
  font-weight: 400;
}
.card {
  border-color: var(--var-border-color);
  margin: 7px;
}
.info {
  font-size: 0.7rem;
  font-weight: 300;
}
.info-body {
  padding: 0.8rem;
}
.value {
  font-size: 2.25rem;
  font-weight: 300;
}
</style>
<style>
.kafka-message {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 50%;
  width: 50%;
}
</style>
