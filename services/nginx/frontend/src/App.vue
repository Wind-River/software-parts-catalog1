<script setup lang="ts">
//Initialize navigation drawer links
const items = [
  { text: "Home", link: "/" },
  { text: "Add", link: "/file" },
  { text: "Update", link: "/update" },
  { text: "About", link: "/about" },
];
</script>
<template>
  <v-app>
    <v-app-bar>
      <v-list
        class="d-flex flex-row w-100 justify-end px-6 bg-blue-grey-darken-3"
      >
        <v-list-item
          v-for="(item, i) in items"
          :key="i"
          :value="item"
          :to="item.link"
        >
          <v-list-item-title v-text="item.text" />
        </v-list-item>
      </v-list>
    </v-app-bar>
    <v-main>
      <router-view></router-view>
    </v-main>
  </v-app>
</template>
