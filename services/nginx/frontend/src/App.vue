<template>
  <v-app>
    <v-app-bar density="compact">
      <v-list
        class="d-flex flex-row w-100 justify-end px-4 bg-blue-grey-darken-3"
      >
        <v-list-item
          v-for="(item, i) in items"
          :key="i"
          :value="item"
          :to="item.link"
        >
          <div slot:title class="text-subtitle-2">{{ item.text }}</div>
        </v-list-item>
      </v-list>
    </v-app-bar>
    <v-main>
      <router-view></router-view>
    </v-main>
  </v-app>
</template>
<script setup lang="ts">
import { provide } from "vue"
//Initialize navigation drawer links
const items = [
  { text: "Home", link: "/" },
  { text: "Add", link: "/file" },
  { text: "Update", link: "/update" },
  { text: "About", link: "/about" },
]
provide("version", 1.0)
</script>
