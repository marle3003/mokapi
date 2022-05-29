<template>
  <div v-if="workflows !== null">
    <p class="label">Workflows</p>
    <b-table small hover class="dataTable" :items="workflows" :fields="fields">
      <template v-slot:cell(name)="data">
        {{ data.item.tags.name }}
      </template>
      <template v-slot:cell(duration)="data">
        {{ data.item.duration | duration }}
      </template>
      <template v-slot:cell(tags)="data">
        <div v-for="(v, k) in data.item.tags" :key="k" class="line">
          {{ k + ": " + v }}
        </div>
      </template>
    </b-table>
  </div>
</template>

<script>
import Filters from '@/mixins/Filters'

export default {
  name: 'Workflows',
  props: ['workflows'],
  mixins: [Filters],
  data () {
    return {
      fields: ['name', 'duration', 'tags'],
    }
  },
  methods: {
  }
}
</script>

<style scoped>
.step{
  line-height: 2rem;
  border-radius: 6px;
  padding-left: 8px;
}
.step.not-collapsed{
  background-color: var(--var-bg-color-secondary);
}
.step svg {
  -moz-transition: all .3s linear;
  -webkit-transition: all .3s linear;
  transition: all .3s linear;
}
.step.not-collapsed svg {
  -moz-transform:rotate(90deg);
  -webkit-transform:rotate(90deg);
  transform:rotate(90deg);
}
.logs .line {
  font-size: 0.65rem;
  display: flex;
  padding-bottom: 3px;
}
.logs .line .number {
  padding-right: 8px;
  color: var(--var-color-line-num);
}
.line svg {
  -moz-transition: all .3s linear;
  -webkit-transition: all .3s linear;
  transition: all .3s linear;
}
.line.not-collapsed svg {
  -moz-transform:rotate(90deg);
  -webkit-transform:rotate(90deg);
  transform:rotate(90deg);
}
</style>
