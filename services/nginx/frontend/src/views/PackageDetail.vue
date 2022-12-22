<template>
  <v-container>
    <component
      :is="Modal"
      v-if="showModal"
      @close="showModal = false"
      :payload="modalPayload"
    >
      <template v-slot:header>
        <h2>Download Failed</h2>
      </template>
      <p>{{ error }}</p>
    </component>
    <v-progress-circular
      indeterminate
      size="50"
      color="primary"
      v-if="fcFetching"
    ></v-progress-circular>
    <h3 v-if="fcError">{{ fcError }}</h3>
    <h2 v-if="fcData" class="mb-6">{{ fcData.archives[0].name }}</h2>
    <v-table v-if="fcData">
      <thead class="bg-primary">
        <tr>
          <th>Field</th>
          <th>Value</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>File Count</td>
          <td>{{ fcData.file_count }}</td>
        </tr>
        <tr>
          <td>Insert Date</td>
          <td>
            {{
              new Date(fcData.file_collection.insert_date).toLocaleDateString()
            }}
          </td>
        </tr>
        <tr>
          <td>License</td>
          <td>
            {{
              fcData.file_collection.license
                ? fcData.file_collection.license.name
                : fcData.file_collection.license
            }}
          </td>
        </tr>
        <tr>
          <td>Rationale</td>
          <td>{{ fcData.file_collection.license_rationale }}</td>
        </tr>
        <tr>
          <td>Verification Code</td>
          <td>{{ fcData.file_collection.verification_code_two }}</td>
        </tr>
      </tbody>
    </v-table>

    <h3 class="my-4" v-if="fcData">Available Archives</h3>

    <v-table v-if="fcData">
      <thead class="bg-primary">
        <tr>
          <th>Name</th>
          <th>Insert Date</th>
          <th>Checksum</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(archive, index) in fcData.archives" :key="index">
          <td>{{ archive.name }}</td>
          <td>{{ new Date(archive.insert_date).toLocaleDateString() }}</td>
          <td>
            {{
              archive.sha256
                ? "SHA256:" + archive.sha256.substring(0, 10) + "..."
                : "SHA1:" + archive.sha1.substring(0, 10) + "..."
            }}
          </td>
          <td v-if="archive.path != null">
            <v-btn
              @click="downloadArchive(archive.id, archive.name)"
              variant="tonal"
              color="primary"
            >
              Download
            </v-btn>
          </td>
          <td v-else></td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
</template>

<script setup lang="ts">
import { Ref, ref, computed, onBeforeMount } from "vue"
import Modal from "@/components/Modal.vue"

import download from "downloadjs"
import { useRoute } from "vue-router"
import { useQuery } from "@urql/vue"

type ModalPayload = {
  data: string
  filename: string
  mime: string
}
const cid: Ref<number> = ref(0)
const loaded: Ref<boolean> = ref(false)
const data: Ref<any> = ref({})
const showModal: Ref<boolean> = ref(false)
const modalPayload: Ref<ModalPayload | null> = ref(null)

const fileCollectionQuery = useQuery({
  query: `
  query($cid: Int64!){
  file_collection(id: $cid){
    id
    insert_date
    group_container_id
    flag_extract
    flag_license_extract
    license_id
    license {
      name
    }
    license_rationale
    analyst_id
    license_expression
    license_notice
    copyright
    verification_code_one
    verification_code_two
  }
  archives(id: $cid){
    name
    sha256
    sha1
    insert_date
    path
    id
  }
  file_count(id: $cid)
}
`,
  variables: { cid },
})
const fcData = fileCollectionQuery.data
const fcError = fileCollectionQuery.error
const fcFetching = fileCollectionQuery.fetching

const error: Ref<string> = ref("Uninitialized Error")
const route = useRoute()

function downloadArchive(id: number, name: string, depth = 0) {
  // var ok = true
  var ctype = ""
  fetch(
    new Request(`/api/container/download/${id}`, {
      method: "GET",
      mode: "same-origin",
    }),
  )
    .then((response) => {
      ctype = response.headers.get("Content-Type") || ""
      // ok = response.ok
      if (!response.ok) {
        throw response.statusText
      }

      return response.blob()
    })
    .catch((err) => {
      if (depth < 3) {
        downloadArchive(id, name, depth + 1)
      } else {
        error.value = "Unable to retrieve download from server"
        showModal.value = true
      }
    })
    .then((blob) => {
      if (blob instanceof Blob) {
        download(blob, name, ctype)
      }
    })
}

onBeforeMount(function () {
  var id: string
  if (typeof route.params.id === "string") {
    id = route.params.id
  } else {
    id = route.params.id[0]
  }

  cid.value = parseInt(id)
})
</script>

<style></style>
