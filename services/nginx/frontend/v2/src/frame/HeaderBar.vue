<template>
  <div id="header-bar">
    {{title}}
    <router-link to="/">Home</router-link>
    <template v-for="(route, index) in routes">
      |
      <router-link :to="route.path" :key="index">{{route.title}}</router-link>
    </template>
  </div>
</template>

<script>
export default {
  name: 'Header',
  props: ['title', 'routes']
}
</script>

<style scoped>
  #header-bar {
    background-color: red;
    color: white;
    margin: 0 !important;
    width: 100%;
    clear: both;
  }

  a {
      color: white;
      margin-right: 5px;
      margin-left: 5px;
      text-decoration: none;
      text-shadow: 1px 1px 0 #771C1C;
  }

  a.legacy {
    text-decoration: underline !important;
    color: red;
    text-shadow: none !important;
  }

  a:hover {
    color: black;
    border-color: red;
    text-shadow: 1px 1px 0 white;
  }
  a.router-link-exact-active {
    color: black;
  }

  a.router-link-exact-active:hover {
    color: black;
    text-shadow: 1px 1px 0 #771C1C;
  }
</style>
