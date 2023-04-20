<!-- Package file upload page -->
<template>
  <v-container>
    <div class="d-flex flex-column align-center">
      <v-card class="d-flex flex-column pa-4 mt-4 bg-secondary w-75">
        <h3 class="px-8">Upload Parts</h3>
        <Upload
          type="application/*,.bz2,.xz"
          message="Click to select files"
          icon="mdi-upload"
          :processing="processing"
          @sendFiles="handleUpload"
        />
        <div class="text-subtitle align-self-end">
          {{ processedFiles + "/" + fileCount }}
        </div>
      </v-card>
      <v-table
        v-if="uploadedArchives.length > 0 || incompleteUploads.length > 0"
        class="ma-4"
      >
        <thead>
          <tr>
            <th>{{ processedFiles + "/" + fileCount }}</th>
          </tr>
          <tr>
            <th>Name</th>
            <th>Checksum</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(archive, index) in uploadedArchives" :key="index">
            <td>{{ archive.name }}</td>
            <td>{{ archive.sha256 ? archive.sha256 : archive.sha1 }}</td>
            <td>
              <v-icon color="primary">{{
                archive.part ? "mdi-check" : "mdi-update"
              }}</v-icon>
            </td>
          </tr>
          <tr v-for="(name, index) in incompleteUploads" :key="index">
            <td>{{ name }}</td>
            <td>Processing</td>
            <td>
              <v-icon color="primary"> mdi-update </v-icon>
            </td>
          </tr>
        </tbody>
      </v-table>
    </div>
  </v-container>
  <v-dialog v-model="showDialog" transition="scale-transition">
    <v-card width="50%" class="align-self-center">
      <v-btn @click="downloadCSV" color="primary">Download CSV</v-btn>
      <v-btn @click="showDialog = false">Close</v-btn>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useMutation, useQuery } from "@urql/vue"
import { onBeforeMount, Ref, ref } from "vue"
import download from "downloadjs"
import Upload from "@/components/Upload.vue"
import { useRoute } from "vue-router"
import Papa from "papaparse"

//Defines the data types expected to be returned from the catalog
type Archive = {
  name: string
  insert_date: string
  sha256: string
  sha1: string
  part_id: string
  part: Part
}
type Part = {
  id: string
  file_verification_code: string
  type: string
  name: string
  version: string
  label: string
  description: string
  family_name: string
  license: string
  license_rationale: string
  comprised: string
  aliases: string[]
}

//Various refs used in file upload processing
const uploadedArchives: Ref<Archive[]> = ref([])
const processing: Ref<boolean> = ref(false)
const showDialog: Ref<boolean> = ref(false)
const fileCount: Ref<number> = ref(0)
const processedFiles: Ref<number> = ref(0)

//Mutation is responsible for uploading files to the catalog and returning archive data
const uploadMutation = useMutation(`
  mutation($file: Upload!){
    uploadArchive(file: $file){
      archive{
        sha256
        sha1
        Size
        md5
        insert_date
        part_id
        part{
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
      }
    }
  }
`)

//Responsible for handling archives on first upload to catalog
const incompleteUploads: Ref<string[]> = ref([])
const currentName: Ref<string> = ref("")
const archiveQuery = useQuery({
  query: `
  query($archiveName: String){
    archive(name: $archiveName){
      sha256
      sha1
      Size
      md5
      insert_date
      part_id
      part{
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
    }
  }`,
  variables: { archiveName: currentName },
})
const queryResponse = archiveQuery.data
const queryError = archiveQuery.error
const queryFetching = archiveQuery.fetching

//Processes any archives that returned invalid data typically due to first upload
async function processIncomplete() {
  for (const name of incompleteUploads.value) {
    const result = await retrieveArchive(name)
    uploadedArchives.value.push(result)
  }
  processing.value = false
  if (pid.value != undefined) {
    addToPartList(uploadedArchives.value)
  }
  showDialog.value = true
  incompleteUploads.value = []
}

//Retrieves information about given archive from catalog
async function retrieveArchive(name: string) {
  currentName.value = name
  await archiveQuery.executeQuery()
  if (queryResponse.value.archive === null) {
    retrieveArchive(name)
  }
  if (queryError.value) {
    console.log(queryError.value)
  }
  if (queryFetching.value) {
    console.log(queryFetching.value)
  }
  if (queryResponse.value.archive) {
    queryResponse.value.archive.name = name
    return queryResponse.value.archive
  }
  return
}

//Converts returned archive and part data into downloadable csv
function convertToCSV(arr: Archive[]) {
  const array = [
    "part_id",
    "file_verification_code",
    "type",
    "name",
    "version",
    "label",
    "description",
    "family_name",
    "license",
    "license_rationale",
    "comprised",
    "\n",
  ]

  const parsedArr = arr
    .map((archive) => {
      if (archive.part !== null) {
        return [
          archive.part.id,
          archive.part.file_verification_code,
          archive.part.type,
          archive.part.name,
          archive.part.version,
          archive.part.label,
          archive.part.description,
          archive.part.family_name,
          archive.part.license,
          archive.part.license_rationale,
          archive.part.comprised,
        ].toString()
      } else return
    })
    .join("\n")

  return array.toString() + parsedArr
}

//Converts data into csv format and then allows user to download csv file
function downloadCSV() {
  download(
    Papa.unparse(uploadedArchives.value.map((value) => value.part)),
    "catalog-prefilled",
    "text/csv",
  )
  showDialog.value = false
}

//Uses file upload component to upload files to catalog
async function handleUpload(files: File[]) {
  processing.value = true
  fileCount.value += files.length
  for (const file of files) {
    processedFiles.value++
    await uploadMutation
      .executeMutation({ file: file })
      .then((value) => {
        if (value.error) {
          console.log(value.error)
          if (
            value.error?.message ===
            "[GraphQL] the requested element is null which the schema does not allow"
          ) {
            incompleteUploads.value.push(file.name)
          }
        }
        return value
      })
      .then((value) => {
        if (value.data.uploadArchive.archive) {
          const uploadedArchive = value.data.uploadArchive.archive
          uploadedArchive.name = file.name
          uploadedArchives.value.push(uploadedArchive)
        }
      })
  }
  processing.value = false
  if (pid.value != undefined && uploadedArchives.value.length > 0) {
    addToPartList(uploadedArchives.value)
  }
  if (uploadedArchives.value.length > 0) {
    showDialog.value = true
  }
}

//If a partlist has been selected to add parts this will perform that mutation
const partListMutation = useMutation(`
mutation($id: Int64!, $parts: [UUID]){
  updatePartList(id: $id, parts: $parts){
    id
    name
    parent_id
  }
}
`)

async function addToPartList(archives: Archive[]) {
  const partIDS: string[] = archives.map((value) => value.part_id)
  await partListMutation
    .executeMutation({ id: pid.value, parts: partIDS })
    .then((value) => {
      if (value.error) {
        console.log(value.error)
      }
      if (value.data) {
        console.log(value.data)
      }
    })
}

//Responsible for checking if a partlist has been selected to add parts to
const route = useRoute()
const pid: Ref<number | undefined> = ref()

onBeforeMount(function () {
  var id: string
  if (route.params.id === undefined) {
    return
  } else if (typeof route.params.id === "string") {
    id = route.params.id
  } else {
    id = route.params.id[0]
  }

  pid.value = parseInt(id)
})
</script>
