<!-- Part update page -->
<template>
  <v-container>
    <div class="d-flex flex-column align-center">
      <v-card class="d-flex flex-column pa-4 mt-4 bg-secondary w-75">
        <h3 class="px-8">Update Part Details</h3>
        <Upload
          type="text/csv"
          message="Click to select CSV"
          icon="mdi-table-edit"
          :processing="processing"
          @sendFiles="handleUpload"
        />
      </v-card>
      <v-table v-if="uploadedCSV.length > 0 && !graphqlError" class="my-4">
        <thead>
          <tr>
            <th>Name</th>
            <th>Verification Code</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(csv, index) in uploadedCSV" :key="index">
            <td>{{ csv.name }}</td>
            <td>{{ csv.file_verification_code }}</td>
            <td><v-icon>mdi-check</v-icon></td>
          </tr>
        </tbody>
      </v-table>
    </div>
  </v-container>
  <v-dialog v-model="showDialog" transition="scale-transition">
    <v-card width="50%" class="align-self-center">
      <h3 class="mx-6 mt-6">{{ dialogMessage }}</h3>
      <v-btn @click="showDialog = false" color="primary" class="ma-6"
        >Done</v-btn
      >
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useMutation } from "@urql/vue"
import { Ref, ref } from "vue"
import Upload from "@/components/Upload.vue"
import Papa from "papaparse"

//Defines the structure of an update csv file
type UpdateCSV = {
  comprised: string
  family_name: string
  file_verification_code: string
  license: string
  license_notice: string
  license_rationale: string
  name: string
  id: string
  type: string
  version: string
}

type PartInput = {
  id: string
  type?: string
  name?: string
  version?: string
  family_name?: string
  file_verification_code: string
  license?: string
  license_rationale?: string
  license_notice?: string
  comprised?: string
}

//Handles csv files and related processing elements
const uploadedCSV: Ref<UpdateCSV[]> = ref([])
const processing: Ref<boolean> = ref(false)
const showDialog: Ref<boolean> = ref(false)
const dialogMessage: Ref<string> = ref("Uploaded CSV has completed processing.")

//Updates the part record within the catalog
const updateMutation = useMutation(`
  mutation($partInput: PartInput!){
    updatePart(partInput: $partInput){
      file_verification_code
      license
      license_rationale
    }
  }
`)

const graphqlError = updateMutation.error

//Parses csv file into structure used for update mutations
function parseCSV(file: File): Promise<UpdateCSV[]> {
  return new Promise<UpdateCSV[]>((resolve, reject) => {
    Papa.parse<UpdateCSV>(file, {
      complete: (results) => {
        resolve(results.data)
      },
      error: (error) => {
        reject(error)
      },
      skipEmptyLines: true,
      header: true,
    })
    //   const reader = new FileReader()
    //   reader.onload = (event) => {
    //     if (typeof event.target?.result === "string") {
    //       const resultString: string = event.target!.result
    //       uploadedCSV.value.push(...csvToArray(resultString))
    //       resolve(uploadedCSV.value)
    //     } else {
    //       reject("error reading uploaded file")
    //     }
    //   }
    //   reader.onerror = (event) => {
    //     reject(event.target?.error)
    //   }
    //   reader.readAsText(file)
    // })
  })
}

function csvToArray(csvString: string): UpdateCSV[] {
  const rows = csvString.slice(csvString.indexOf("\n") + 1).split("\n")
  const arr = rows.reduce((arr: UpdateCSV[], row: string) => {
    const fields = row.split(",")
    const id = fields[0]
    const file_verification_code = fields[1]
    const type = fields[2]
    const name = fields[3]
    const version = fields[4]
    const family_name = fields[5]
    const license = fields[6]
    const license_rationale = fields[7]
    const license_notice = fields[8]
    const comprised = fields[9]
    arr.push({
      id,
      file_verification_code,
      type,
      name,
      version,
      family_name,
      license,
      license_rationale,
      license_notice,
      comprised,
    })
    return arr
  }, [])
  return arr.filter((value) => value.id !== "")
}

//Function parses csv into processable format and then executes mutations
async function handleUpload(files: File[]) {
  processing.value = true
  uploadedCSV.value = []
  for (const file of files) {
    await parseCSV(file).then((value) => {
      uploadedCSV.value = [...uploadedCSV.value, ...value]
    }).catch((error) => {
      console.log(error)
    })
  }
  console.log(uploadedCSV.value)
  for (const csv of uploadedCSV.value) {
    if (csv.file_verification_code === undefined || "") {
      dialogMessage.value = "Verification code required to update parts."
    } else {
      let updatePartInput: PartInput = {
        id: csv.id,
        file_verification_code: csv.file_verification_code,
      }
      if (csv.type) {
        updatePartInput.type = csv.type
      }
      if (csv.name) {
        updatePartInput.name = csv.name
      }
      if (csv.version) {
        updatePartInput.version = csv.version
      }
      if (csv.family_name) {
        updatePartInput.family_name = csv.family_name
      }
      if (csv.license) {
        updatePartInput.license = csv.license
      }
      if (csv.license_rationale) {
        updatePartInput.license_rationale = csv.license_rationale
      }
      if (csv.license_notice) {
        updatePartInput.license_notice = csv.license_notice
      }
      if (
        csv.comprised &&
        csv.comprised !== "00000000-0000-0000-0000-000000000000"
      ) {
        updatePartInput.comprised = csv.comprised
      }
      updateMutation
        .executeMutation({
          partInput: updatePartInput,
        })
        .then((result) => {
          if (
            result.error?.message === "[GraphQL] no data was given to update"
          ) {
            dialogMessage.value = "No data was given to update."
          } else if (result.error) {
            dialogMessage.value =
              "Error processing uploaded CSV, please check formatting."
          } else {
            dialogMessage.value = "Uploaded CSV has completed processing."
          }
        })
        .catch((error) => {
          console.log(error)
        })
    }
    processing.value = false
    showDialog.value = true
  }
}
</script>
