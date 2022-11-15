<template>
  <div id="content">
    <slot/>
  </div>
</template>

<script>
export default {
  name: 'Content'
}
</script>

<style>
  #content {
    min-height: 500px;
  }
</style>
