<!-- Custom component for displaying part lists in a nested structure -->
<template>
  <v-list-group>
    <template v-slot:activator="{ props }">
      <v-list-item v-bind="props" @click="getChildren">
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
    <v-list-item>
      <v-btn color="primary" size="x-small" @click="dialogVisible = true"
        >Add List</v-btn
      >
    </v-list-item>
  </v-list-group>
  <v-dialog v-model="dialogVisible">
    <v-card class="w-50 mx-auto pa-4">
      <h3 class="mb-4">Add Part List</h3>
      <v-text-field
        label="List Name"
        v-model="newListName"
        color="primary"
        variant="outlined"
      ></v-text-field>
      <div class="d-flex justify-end">
        <v-btn class="mx-2" color="primary" @click="dialogVisible = false"
          >Cancel</v-btn
        >
        <v-btn class="mx-2" color="primary" @click="addPartList">Submit</v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>
<script setup lang="ts">
import { usePartListStore } from "@/stores/partlist"
import { useMutation, useQuery } from "@urql/vue"
import { Ref, ref } from "vue"
import { useRouter } from "vue-router"

type PartList = {
  id: number
  name: string
  parent_id: number
  children: PartList[]
}

//Receives a given part list as parent
const props = defineProps<{
  partlist: PartList
}>()

const partListStore = usePartListStore()

const dialogVisible: Ref<boolean> = ref(false)
const newListName: Ref<string> = ref("")

const subListMutation = useMutation(`
mutation($name: String!, $parentID: Int64!){
  addPartList(name: $name, parent_id: $parentID){
    id
    name
    parent_id
  }
}
`)

async function addPartList() {
  if (newListName.value.length > 0) {
    await subListMutation
      .executeMutation({ name: newListName.value, parentID: props.partlist.id })
      .then((value) => {
        if (value.data) {
          getChildren()
        }
      })
  }
  dialogVisible.value = false
}

const subListQuery = useQuery({
  query: `
  query($partlistID: Int64!){
    partlists(parent_id: $partlistID){
      id
      name
      parent_id
    }
  }
  `,
  pause: true,
  variables: { partlistID: props.partlist.id },
})

const subPartLists = subListQuery.data
const subPartListQueryError = subListQuery.error

async function getChildren() {
  subListQuery.executeQuery().then((value) => {
    if (subPartListQueryError.value) {
      console.log(value.error)
    } else if (subPartLists.value) {
      props.partlist.children = subPartLists.value.partlists
    }
  })
}

const router = useRouter()

//Redirects browser to part list detail page for selected list
const redirect = (id: number): void => {
  router.push({
    name: "Part List Detail",
    params: { id: id },
  })
}
</script>
