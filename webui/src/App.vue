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
    'top-header': TopHeader
  }
}
</script>
<style>
  @import './assets/vars.css';
  @import './assets/prism.css';

  body{
    background-color: var(--var-bg-color);
    height: 100%;
    color:var(--var-color-primary);
    border-color: #141414;
    font-family: 'Poppins', sans-serif;
    font-weight: 400;
    line-height: normal;
    font-size: 0.9rem;
  }
  #app {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  a {
    color: var(--var-link-color);
  }
  a:hover{
    color: var(--var-link-color);
  }
  h1 {
    font-size: 1.75rem;
    font-weight: 600;
  }
  h2 {
    font-size: 1.25rem;
    font-weight: 500;
  }
  h3 {
    font-size: 1.1rem;
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
  .break{
    word-break: break-all;
  }
  table.dataTable {
    border-left-style: hidden;
    border-right-style: hidden;
    color:var(--var-color-primary);
    font-size: 0.8rem;
  }
  table.dataTable.selectable td {
    cursor: pointer;
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
    font-weight: 500;
  }
  table.dataTable td{
    border-top-width: 2px;
    border-bottom-width: 2px;
    border-color: var(--var-table-border-color);
    border-style: solid;
    padding: 4px 0 3px 12px;
    border-left-style: hidden;
    border-right-style: hidden;
    font-size: 0.75rem;
  }
  .table-hover tbody tr:hover:not(.b-table-details) td, .table-hover tbody tr:hover th {
    color: var(--var-color-primary);
  }
  .table-hover tbody tr.b-table-details:hover {
    background-color: var(--var-bg-color-primary);
    color:var(--var-color-primary);
  }
  .nav-tabs .nav-item .nav-link.active,
  .nav-tabs .nav-item.show .nav-link {
    border-color: var(--var-table-border-color);
    background-color: var(--var-tab-bg-color);
    color: var(--var-tab-header-color);
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
  .card {
    border: 1px solid #141414;
    border-radius: 10px;
    background-color: var(--var-bg-color-primary);
  }
 .card-header, .nav-tabs, .popover-header {
    border-color: var(--var-table-border-color);
    border-width: 2px;
  }
  .popover {
    border: 1px solid #141414;
    border-radius: 10px;
    background-color: var(--var-bg-color-primary);
  }
  .popover-header {
    color:var(--var-color-primary);
    background-color: var(--var-bg-color-primary);
  }
  .popover-body {
    color:var(--var-color-primary);
  }
  .label {
    font-size: 0.6rem;
    margin: 0 0 0.3rem;
  }
  .list-group-item {
    background-color: var(--var-bg-color);
  }
  .list-group-item-action {
    color:var(--var-color-primary);
  }
  .router-link-active {
    background-color: var(--var-bg-color);
  }
  .router-link-exact-active {
    background-color: var(--var-bg-color);
    text-decoration: underline;
  }
  .list-group-item-action:hover,
  .router-link-active:hover {
    background-color: var(--var-bg-color);
    color:var(--var-color-primary);
    text-decoration: underline;
  }
  .list-group-item-action:visited,
  .router-link-active:visited {
    background-color: var(--var-bg-color);
    color:var(--var-color-primary);
    text-decoration: underline;
  }
  .list-group-item:focus {
    outline: none;
  }
  .markdown-body {
    color:var(--var-color-primary);
    font-size: 0.9rem;
    line-height: normal;
  }
  .markdown-body a {
    color: var(--var-link-color);
  }
  /* markdownit */
  div.code-toolbar > .toolbar{
    opacity: 1;
  }
  code[class*="language-"], pre[class*="language-"] {
    font-size: 0.8rem;
    background-color: var(--var-code-bg-color);
    text-shadow: none;
    color: var(--orange);
  }
  .token.atrule, .token.attr-value, .token.keyword {
    color: var(--var-code-blue);
  }
  .token.comment, .token.prolog, .token.doctype, .token.cdata{
    color: var(--var-code-green);
  }
  .token.property, .token.tag, .token.boolean, .token.number, .token.constant, .token.symbol, .token.deleted{
    color: var(--orange);
  }
  .token.operator, .token.entity, .token.url, .language-css .token.string, .style .token.string{
    background-color: var(--var-code-bg-color);
  }
</style>
