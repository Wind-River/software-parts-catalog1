<!-- This page shows details of selected part list -->
<template>
  <v-container>
    <v-card class="pa-4">
      <div class="d-flex justify-end align-center mt-2">
        <v-card-title v-if="parts" class="text-primary me-auto">{{
          parts.partlist.name
        }}</v-card-title>

        <v-btn class="mx-2" color="primary" size="small" @click="addParts"
          >Add Parts</v-btn
        >
        <v-btn
          v-if="parts && parts.partlist_parts.length > 0"
          color="primary"
          size="small"
          @click="downloadCSV"
          >Download CSV</v-btn
        >
        <v-btn
          class="mx-2"
          color="primary"
          size="small"
          @click="showDeleteConfirmation"
          >Delete</v-btn
        >
      </div>
      <v-table>
        <thead>
          <tr>
            <th>Part Name</th>
            <th>License</th>
            <th>Verification Code</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-if="parts && parts.partlist_parts.length > 0"
            v-for="(item, index) in parts.partlist_parts"
            :key="index"
          >
            <td style="cursor: pointer" @click="redirect(item.id)">
              {{ item.label }}
            </td>
            <td>{{ item.license }}</td>
            <td>{{ item.file_verification_code.substring(14, 24) }}</td>
            <td>
              <v-btn class="me-2" size="small" color="primary" icon="mdi-delete" @click="deletePartFromList(item.id)"></v-btn>
              <v-btn color="primary" @click="redirect(item.id)"
                >Open</v-btn
              >
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card>
    <v-dialog v-model="deleteDialogVisible">
      <v-card class="w-50 pa-4 mx-auto">
        <h3 class="text-center mb-2">
          Are you sure you want to delete this part list?
        </h3>
        <div class="d-flex justify-center">
          <v-btn
            class="mx-2"
            color="primary"
            size="small"
            @click="deleteDialogVisible = false"
            >Cancel</v-btn
          >
          <v-btn
            class="mx-2"
            color="primary"
            size="small"
            @click="deletePartList"
            >Confirm</v-btn
          >
        </div>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { usePartListStore } from "@/stores/partlist"
import { useMutation, useQuery } from "@urql/vue"
import download from "downloadjs"
import Papa from "papaparse"
import { onBeforeMount, onMounted, ref, Ref } from "vue"
import { useRoute, useRouter } from "vue-router"

//Handles route parameters used in getting part list details
const partListID: Ref<number> = ref(0)
const route = useRoute()
const router = useRouter()
const partListStore = usePartListStore()

//Retrieves a partlist from the catalog and its parts
const partQuery = useQuery({
  query: `
  query($id: Int64!){
    partlist_parts(id: $id){
      id
      type
      name
      version
      label
      description
      family_name
      file_verification_code
      size
      license
      license_rationale
      comprised
    }
    partlist(id: $id){
      name
    }
  }
  `,
  variables: { id: partListID },
})

const parts = partQuery.data

//This section handles deletion of a part list
const deleteDialogVisible: Ref<boolean> = ref(false)

function showDeleteConfirmation() {
  deleteDialogVisible.value = true
}

const deleteMutation = useMutation(`
mutation($id: Int64!){
  deletePartList(id: $id){
  id
  name
  parent_id
  }
}
`)

async function deletePartList() {
  await deleteMutation
    .executeMutation({ id: partListID.value })
    .then((value) => {
      if (value.error) {
        console.log(value.error)
      }
      if (value.data) {
        partListStore.deletePartList(partListID.value)
        router.push({
          name: "PartLists",
        })
      }
    })
}

const deletePartFromListMutation = useMutation(`
  mutation($list_id: Int64!, $part_id: UUID!){
    deletePartFromList(list_id: $list_id, part_id: $part_id){
      id
      name
    }
  }
`)

async function deletePartFromList(part_id: string){
  await deletePartFromListMutation
    .executeMutation({ list_id: partListID.value, part_id: part_id})
    .then((value) => {
      if (value.error) {
        console.log(value.error)
      }
      if (value.data) {
        partQuery.executeQuery()
      }
    })
}


//This will redirect router to upload page with part list as params
function addParts() {
  router.push({
    name: "PartListAdd",
    params: { id: partListID.value },
  })
}

function downloadCSV() {
  download(
    Papa.unparse(parts.value.partlist_parts),
    parts.value.partlist.name + "-" + new Date().toISOString().slice(0,length-8),
    "text/csv",
  )
}

//Collects id from route params
onBeforeMount(function () {
  var id: string
  if (typeof route.params.id === "string") {
    id = route.params.id
  } else {
    id = route.params.id[0]
  }

  partListID.value = parseInt(id)
})

//Redirects browser to package detail for selected part
function redirect(id: number): void {
  router.push({
    name: "Package Detail",
    params: { id: id },
  })
}
</script>
