<template>
  <div
    @click="copyText"
    style="cursor: pointer"
  >
    <slot></slot>
    <v-icon size="x-small" color="primary" class="ms-1">{{ copyIcon }}</v-icon>
  </div>
</template>
<script setup lang="ts">
import { ref, Ref } from "vue"

const props = defineProps<{
    copytext: string
}>()
const copyIcon: Ref<string> = ref("mdi-content-copy")
const copyText = () => {
  copyIcon.value = "mdi-check"
  navigator.clipboard.writeText(props.copytext)
  setTimeout(() => {
    copyIcon.value = "mdi-content-copy"
  }, 1500)
}
</script>
