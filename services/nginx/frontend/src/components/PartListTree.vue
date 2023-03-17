<!-- Custom component for displaying part lists in a nested structure -->
<template>
  <v-list-group>
    <template v-slot:activator="{ props }">
      <v-list-item v-bind="props">
        <v-list-item-title
          @click="redirect(partlist.id)"
          style="cursor: pointer"
          >{{ partlist.name }}</v-list-item-title
        >
      </v-list-item>
    </template>

    <PartListTree
      v-if="partlist.children"
      v-for="list in partlist.children"
      :partlist="list"
    >
    </PartListTree>
  </v-list-group>
</template>
<script setup lang="ts">
import { useRouter } from "vue-router"

type PartList = {
  id: number
  name: string
  parent_id: number
  children: PartList[]
}

const router = useRouter()

//Receives a given part list as parent
const props = defineProps<{
  partlist: PartList
}>()

//Redirects browser to part list detail page for selected list
const redirect = (id: number): void => {
  router.push({
    name: "Part List Detail",
    params: { id: id },
  })
}
</script>
