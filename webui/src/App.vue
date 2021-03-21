<template>
  <div id="app">
    <div id="nav">
      <top-header></top-header>
    </div>
    <router-view/>
  </div>
</template>

<script>
import TopHeader from '@/components/TopHeader.vue'

export default {
  name: 'App',
  components: {
    'top-header': TopHeader,

  }
}
</script>
<style>
  @import './assets/vars.css';
  @import url('https://fonts.googleapis.com/css2?family=Roboto&display=swap');

  body{
    background-color: var(--var-bg-color);
    height: 100%;
    color:var(--var-color-primary);
    border-color: #141414;
    font-family: 'Roboto', sans-serif;
    font-weight: 400;
    line-height: normal;
    font-size: 1.0rem;
  }
  #app {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  h1 {
    font-size: 1.75rem;
    font-weight: 600;
  }
  h2 {
    font-size: 1.25rem;
    font-weight: 500;
  }
  pre {
    color:var(--var-color-primary);
  }
  .operation{
      text-transform: uppercase;
      color:var(--var-color-primary);
  }
  .operation.put, .operation.patch{
    background-color: var(--orange);
  }
  .operation.get, .operation.options{
    background-color: var(--blue);
  }
  .operation.delete{
    background-color: var(--red);
  }
  .operation.post{
    background-color: var(--green);
  }
  .operation.head, .operation.trace{
    background-color: var(--yellow);
  }
  .success{
    color: var(--green);
  }
  .warning{
    color: #ffc107;
  }
  .danger{
    color: #dc3545;
  }
  .client-error{
      color: var(--orange);
  }
  .card {
    border: 1px solid #141414;
    border-radius: 10px;
    background-color: #292933;
  }
  table.dataTable {
    border-left-style: hidden;
    border-right-style: hidden;
    color:var(--var-color-primary);
    font-size: 0.8rem;
  }
  table.dataTable thead tr{
    color: var(--var-table-header-color);
    background-color: var(--var-table-header-bg-color);   
  }
  table.dataTable thead th{
    padding: 3px 0 3px 12px;
    border-color: var(--var-table-border-color);
    border-top-width: 2px;
    border-bottom-width: 2px;
  }
  table.dataTable td{
    border-width: 2px;
    border-color: var(--var-table-border-color);
    border-style: solid;
    padding: 4px 0 3px 12px;
  }
  .table-hover tbody tr:hover td, .table-hover tbody tr:hover th {
    color: var(--var-color-primary);
  }
  .markdown-body {
    color:var(--var-color-primary);
    font-size: 0.8rem;
    line-height: normal;
  }
  .nav-tabs .nav-item .nav-link.active,
  .nav-tabs .nav-item.show .nav-link {
    border-color: var(--var-table-border-color);
    background-color: var(--var-tab-bg-color);
  }
  .nav-pills .nav-link.active,
  .nav-pills .show > .nav-link{
    border-color: var(--var-table-border-color);
    background-color: var(--var-tab-bg-color);
  }
  .nav-pills .nav-link:hover,
  .nav-pills .show > .nav-link:hover{
    border-color: var(--var-table-border-color);
    background-color: var(--var-tab-bg-color);
  }
 .card-header, .nav-tabs {
    border-color: var(--var-table-border-color);
    border-width: 2px;
    
  }
  .label {
    font-size: 0.6rem;
    margin: 0;
    margin-bottom: 0.3rem;
  }
</style>
