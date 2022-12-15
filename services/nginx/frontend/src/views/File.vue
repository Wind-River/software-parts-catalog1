<!-- Package file upload page -->
<template>
  <v-container>
    <v-card class="d-flex flex-column pa-8 mt-12 bg-secondary">
      <h2 class="px-8">Upload Parts</h2>
      <Upload
        type="application/*"
        message="Click to select files"
        :processing="processing"
        @sendFiles="handleUpload"
      />
    </v-card>
    <v-table v-if="uploadedArchives.length > 0" class="ma-12">
      <thead>
        <tr>
          <th>Name</th>
          <th>Checksum</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(archive, index) in uploadedArchives" :key="index">
          <td>{{ archive.name }}</td>
          <td>{{ archive.sha256 ? archive.sha256 : archive.sha1 }}</td>
          <td><v-icon icon="mdi-check" color="primary"></v-icon></td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
  <v-dialog v-model="showDialog" transition="scale-transition">
    <v-card width="50%" class="align-self-center">
      <v-btn @click="downloadCSV" color="primary">Download CSV</v-btn>
      <v-btn @click="showDialog = false">Close</v-btn>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useMutation } from "@urql/vue"
import { Ref, ref } from "vue"
import download from "downloadjs"
import Upload from "@/components/Upload.vue"

type Archive = {
  id: number
  name: string
  insert_date: string
  sha256: string
  sha1: string
  extract_status: string
  file_collection_id: string
  file_collection: FileCollection
}
type FileCollection = {
  verification_code_one: string
  verification_code_two: string
  license_expression: string
  license_rationale: string
  license_notice: string
  copyright: string
}

const uploadedArchives: Ref<Archive[]> = ref([])
const processing: Ref<boolean> = ref(false)
const showDialog: Ref<boolean> = ref(false)

const uploadMutation = useMutation(`
  mutation($file: Upload!){
    uploadArchive(file: $file){
      archive{
        id
        name
        insert_date
        sha256
        sha1
        extract_status
        file_collection_id
        file_collection{
          verification_code_one
          verification_code_two
          license_expression
          license_rationale
          license_notice
          copyright
        }
      }
    }
  }
`)

function convertToCSV(arr: Archive[]) {
  const array = [
    "name",
    "insert_date",
    "checksum",
    "verification_code",
    "license",
    "license_rationale",
    "license_notice",
    "copyright",
    "\n",
  ]

  const parsedArr = arr
    .map((archive) => {
      if (archive.file_collection !== null) {
        return [
          archive.name,
          archive.insert_date,
          archive.sha256 ? archive.sha256 : archive.sha1,
          archive.file_collection.verification_code_two
            ? archive.file_collection.verification_code_two
            : archive.file_collection.verification_code_one,
          archive.file_collection.license_expression,
          archive.file_collection.license_rationale,
          archive.file_collection.license_notice,
          archive.file_collection.copyright,
        ].toString()
      } else
        return [
          archive.name,
          archive.insert_date,
          archive.sha256 ? archive.sha256 : archive.sha1,
        ]
    })
    .join("\n")

  return array.toString() + parsedArr
}

function downloadCSV() {
  download(convertToCSV(uploadedArchives.value), "tk-prefilled", "text/csv")
  showDialog.value = false
}

async function handleUpload(files: File[]) {
  processing.value = true
  console.log(files)
  uploadedArchives.value = []
  let retry = false
  for (const file of files) {
    await uploadMutation
      .executeMutation({ file: file })
      .then((value) => {
        if (value.error) {
          console.log(value.error)
          retry = true
        }
        return value
      })
      .then((value) => {
        if (value.data.uploadArchive.archive) {
          uploadedArchives.value.push(value.data.uploadArchive.archive)
        }
      })
  }
  if (retry) {
    handleUpload(files)
  } else {
    processing.value = false
    showDialog.value = true
  }
}
</script>
