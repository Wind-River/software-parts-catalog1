<template>
  <div id="header">
    <h1>{{title}}</h1>
  </div>
</template>

<script>
export default {
  name: 'Header',
  props: ['title']
}
</script>

<style scoped>
  #header {
    background-color: black;
    color: white;
    margin: 0 !important;
    display: inline-block !important;
    width: 100%;
    clear: both;
  }
</style>
