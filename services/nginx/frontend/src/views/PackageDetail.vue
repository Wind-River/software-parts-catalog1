<!-- This page displays details about a given part -->
<template>
  <v-alert
    v-model="copyAlertVisible"
    class="elevation-5"
    type="success"
    icon="mdi-content-copy"
    width="100vw"
    position="fixed"
  >
    {{ copyAlertMessage }}
  </v-alert>
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
      v-if="partFetching"
    ></v-progress-circular>
    <h3 v-if="partError">{{ partError }}</h3>
    <h2 v-if="partData" class="mb-6">{{ partData.part.name }}</h2>
    <v-card v-if="partData">
      <v-row>
        <v-col cols="8" class="pe-0">
          <v-table v-if="partData" height="264" fixed-header>
            <thead>
              <tr>
                <th class="bg-primary">Information</th>
                <th class="bg-primary"></th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td class="font-weight-bold">File Count</td>
                <td>{{ partData.file_count }}</td>
              </tr>
              <tr>
                <td class="font-weight-bold">License</td>
                <td>
                  {{ partData.part.license }}
                </td>
              </tr>
              <tr>
                <td class="font-weight-bold">Description</td>
                <td>
                  {{
                    partData.part.license_rationale &&
                    partData.part.license_rationale !== "null"
                      ? JSON.parse(partData.part.license_rationale).rationale
                      : ""
                  }}
                </td>
              </tr>
              <tr>
                <td class="font-weight-bold">Verification Code</td>
                <td
                  @click="copyText(partData.part.file_verification_code)"
                  style="cursor: pointer"
                >
                  {{ partData.part.file_verification_code.slice(-10) }}
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-col>
        <v-divider vertical></v-divider>
        <v-col cols="4" class="ps-0">
          <v-table fixed-header height="264">
            <thead>
              <tr>
                <th class="bg-primary">Profiles</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-if="partData && partData.part.profiles.length > 0"
                v-for="(profile, index) in partData.part.profiles"
                :key="index"
              >
                <td
                  @click="showProfile(pid, profile.key)"
                  style="cursor: pointer"
                  class="text-blue"
                >
                  <v-icon class="me-1">mdi-file-document-outline</v-icon>
                  <span class="text-decoration-underline">{{
                    profile.key.slice(0, 1).toUpperCase() + profile.key.slice(1)
                  }}</span>
                </td>
              </tr>
              <tr v-else>
                <td>No Profiles</td>
              </tr>
            </tbody>
          </v-table>
        </v-col>
      </v-row>
    </v-card>

    <h3
      class="my-4"
      v-if="
        partData &&
        partData.part.comprised &&
        partData.part.comprised !== '00000000-0000-0000-0000-000000000000'
      "
    >
      Comprised Of
      <v-btn
        size="small"
        class="ms-2"
        color="primary"
        @click="showComprisedParts()"
        >Show/Hide</v-btn
      >
    </h3>
    <v-table v-if="partData && comprisedVisible">
      <thead>
        <tr>
          <th class="bg-primary">ID</th>
          <th class="bg-primary"></th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>{{ partData.part.comprised }}</td>
          <td>
            <v-btn
              @click="redirect(partData.part.comprised)"
              variant="tonal"
              color="primary"
            >
              Open
            </v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>

    <v-divider
      v-if="
        partData &&
        partData.part.comprised &&
        partData.part.comprised !== '00000000-0000-0000-0000-000000000000'
      "
    ></v-divider>

    <h3 class="my-4" v-if="partData && partData.part.sub_parts.length > 0">
      Contains
      <v-btn
        size="small"
        class="ms-2"
        color="primary"
        @click="showContainedParts()"
        >Show/Hide</v-btn
      >
    </h3>
    <v-table v-if="partData && containedVisible">
      <thead>
        <tr>
          <th class="bg-primary">Name</th>
          <th class="bg-primary"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(subpart, index) in partData.part.sub_parts" :key="index">
          <td>{{ subpart.part.name }}</td>
          <td>
            <v-btn
              @click="redirect(subpart.part.id)"
              variant="tonal"
              color="primary"
            >
              Open
            </v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>

    <v-divider
      v-if="partData && partData.part.sub_parts.length > 0"
    ></v-divider>

    <h3 class="my-4" v-if="partData && partData.archives.length > 0">
      Available Archives
      <v-btn
        size="small"
        class="ms-2"
        color="primary"
        @click="showAvailableArchives()"
        >Show/Hide</v-btn
      >
    </h3>
    <v-table v-if="partData && availableArchivesVisible">
      <thead class="bg-primary">
        <tr>
          <th class="bg-primary">Name</th>
          <th class="bg-primary">Insert Date</th>
          <th class="bg-primary">Checksum</th>
          <th class="bg-primary"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(archive, index) in partData.archives" :key="index">
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
              size="small"
            >
              Download
            </v-btn>
          </td>
          <td v-else>
            <v-btn size="small" variant="tonal" color="primary">
              Download
            </v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
  <!-- <v-divider vertical></v-divider>
        <v-col class="px-0">
          <v-table fixed-header height="264">
            <thead>
              <tr>
                <th class="bg-primary">Aliases</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-if="partData && partData.part.aliases.length > 0"
                v-for="(alias, index) in partData.part.aliases"
                :key="index"
              >
                <td>
                  {{ alias }}
                </td>
              </tr>
              <tr v-else>
                <td>No Aliases Found</td>
              </tr>
            </tbody>
          </v-table>
        </v-col> -->
</template>

<script setup lang="ts">
import { Ref, ref, computed, onBeforeMount, getCurrentInstance } from "vue"
import Modal from "@/components/Modal.vue"

import download from "downloadjs"
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router"
import { useQuery } from "@urql/vue"

type ModalPayload = {
  data: string
  filename: string
  mime: string
}
const pid: Ref<string> = ref("")
const loaded: Ref<boolean> = ref(false)
const data: Ref<any> = ref({})
const showModal: Ref<boolean> = ref(false)
const modalPayload: Ref<ModalPayload | null> = ref(null)

//This query retrieves part data along with available archives and file count
const fileCollectionQuery = useQuery({
  query: `
  query($pid: UUID!){
  part(id: $pid){
    id
    type
    name
    version
    family_name
    file_verification_code
    size
    license
    license_rationale
    license_notice
    automation_license
    automation_license_rationale
    aliases
    comprised
    sub_parts{
      part{
        id
        name
        type
      }
    }
    profiles{
      key
    }
  }
  archives(id: $pid){
    name
    sha256
    sha1
    insert_date
  }
  file_count(id: $pid)
}
`,
  variables: { pid },
})
const partData = fileCollectionQuery.data
const partError = fileCollectionQuery.error
const partFetching = fileCollectionQuery.fetching

//This section handles display of available archives, comprised parts, and contained parts
const availableArchivesVisible: Ref<boolean> = ref(false)
const showAvailableArchives = () => {
  availableArchivesVisible.value = !availableArchivesVisible.value
}
const comprisedVisible: Ref<boolean> = ref(false)
const showComprisedParts = () => {
  comprisedVisible.value = !comprisedVisible.value
}

const containedVisible: Ref<boolean> = ref(false)
const showContainedParts = () => {
  containedVisible.value = !containedVisible.value
}

// This section handles the copying of file_verification codes on click event
const copyAlertVisible: Ref<boolean> = ref(false)
const copyAlertMessage: Ref<string> = ref("")
const copyText = (value: string) => {
  navigator.clipboard.writeText(value)
  showCopyAlert("File Verification Code Copied to Clipboard")
}
const showCopyAlert = (message: string) => {
  copyAlertMessage.value = message
  copyAlertVisible.value = true
  setTimeout(() => {
    copyAlertVisible.value = false
  }, 2000)
}

const error: Ref<string> = ref("Uninitialized Error")
const route = useRoute()
const router = useRouter()

//This function redirects browser to detail page on selected part
function redirect(id: string): void {
  router.push({
    name: "Package Detail",
    params: { id: id },
  })
}

//This function will redirect browser to selected profile detail page
function showProfile(id: string, key: string) {
  router.push({
    name: "Profile Detail",
    params: { id: id, key: key },
  })
}

//Downloads an available archive
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

//Responsible for pulling part id from router params
onBeforeMount(function () {
  var id: string
  if (typeof route.params.id === "string") {
    id = route.params.id
  } else {
    id = route.params.id[0]
  }

  pid.value = id
})

//This forces update to reload component with new part information
onBeforeRouteUpdate((to) => {
  var id: string
  if (typeof to.params.id === "string") {
    id = to.params.id
  } else {
    id = to.params.id[0]
  }

  pid.value = id
  containedVisible.value = false
  comprisedVisible.value = false
  availableArchivesVisible.value = false
})
</script>

<style></style>
