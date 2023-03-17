<!-- This page is for part list navigation and various part list operations -->
<template>
  <v-container>
    <div class="d-flex my-4">
      <h1 class="me-auto">Part Lists</h1>
      <v-btn class="align-self-center" @click="showDialog" color="primary"
        >Add</v-btn
      >
    </div>
    <v-divider></v-divider>
    <div class="d-flex">
      <h3 v-if="rootPartLists.length === 0">No Part Lists Found</h3>
      <v-list>
        <Tree v-for="partlist in rootPartLists" :partlist="partlist"></Tree>
      </v-list>
    </div>
  </v-container>
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
import { Ref, ref, onBeforeMount } from "vue"
import { PartList, usePartListStore } from "@/stores/partlist"
import Tree from "@/components/PartListTree.vue"
import { useMutation, useQuery } from "@urql/vue"

//Coordinates part lists with pinia store
const partListStore = usePartListStore()
const rootPartLists: Ref<PartList[]> = ref([])

//Handles adding a new part list
const dialogVisible: Ref<boolean> = ref(false)
const newListName: Ref<string> = ref("")

const addPartListMutation = useMutation(`
  mutation($name: String!){
    addPartList(name: $name){
      id
      name
      parent_id
    }
  }
`)

//Retrieves part lists from catalog
const partListQuery = useQuery({
  query: `
  query{
    partlists(parent_id: 0){
      id
      name
      parent_id
    }
  }
  `
})

//Functions for executing add new part list
function showDialog() {
  dialogVisible.value = true
}
const showNoListMessage: Ref<boolean> = ref(false)

async function addPartList() {
  if(newListName.value !== "") {
    await addPartListMutation.executeMutation({name: newListName.value}).then((value)=>{
      if (value.error) {
        console.log(value.error)
      }
      if (value.data) {
        partListStore.addPartList(value.data.addPartList)
      }
    })
    dialogVisible.value = false
  }
}

//Function for retrieving part lists and synchronizes with pinia store
async function fetchPartLists() {
  await partListQuery.executeQuery().then((value) => {
    partListStore.setPartLists(value.data.value.partlists)
  })
  rootPartLists.value = partListStore.partLists
}

onBeforeMount(() => {
  fetchPartLists()
})
</script>
