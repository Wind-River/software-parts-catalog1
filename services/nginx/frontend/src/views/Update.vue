<!-- Package file upload page -->
<template>
  <div class="d-flex justify-center">
    <Upload
      type="text/csv"
      message="Click to select CSV"
      :processing="processing"
      @sendFiles="handleUpload"
    />
  </div>
  <v-table v-if="uploadedCSV.length > 0" class="ma-12">
    <thead>
      <tr>
        <th>Name</th>
        <th>Verification Code</th>
        <th>License</th>
        <th>Rationale</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(csv, index) in uploadedCSV" :key="index">
        <td>{{ csv.name }}</td>
        <td>{{ csv.verification_code }}</td>
        <td>{{ csv.license }}</td>
        <td>{{ csv.license_rationale }}</td>
        <td><v-icon icon="mdi-check" color="primary"></v-icon></td>
      </tr>
    </tbody>
  </v-table>
  <v-dialog v-model="showDialog" transition="scale-transition">
    <v-card width="50%" class="align-self-center">
      <h3 class="mx-6 mt-6">{{dialogMessage}}</h3>
      <v-btn @click="showDialog = false" color="primary" class="ma-6"
        >Done</v-btn
      >
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useMutation } from "@urql/vue";
import { Ref, ref } from "vue";
import Upload from "@/components/Upload.vue";

type UpdateCSV = {
  checksum: string;
  copyright: string;
  insert_date: string;
  license: string;
  license_notice: string;
  license_rationale: string;
  name: string;
  verification_code: string;
};

const uploadedCSV: Ref<UpdateCSV[]> = ref([]);
const processing: Ref<boolean> = ref(false);
const showDialog: Ref<boolean> = ref(false);
const dialogMessage: Ref<string> = ref("Uploaded CSV has completed processing.");

const updateMutation = useMutation(`
  mutation($verificationCode: String!, $license: String, $licenseRationale: String){
    updateFileCollection(verificationCode: $verificationCode, license: $license, licenseRationale: $licenseRationale){
      verification_code_one
      verification_code_two
      license_expression
      license_rationale
    }
  }
`);

function parseCSV(file: File): Promise<UpdateCSV[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (event) => {
      if (typeof event.target?.result === "string") {
        const resultString: string = event.target!.result;
        uploadedCSV.value.push(...csvToArray(resultString));
        resolve(uploadedCSV.value);
      } else {
        reject("error reading uploaded file");
      }
    };
    reader.onerror = (event) => {
      reject(event.target?.error);
    };
    reader.readAsText(file);
  });
}

function csvToArray(csvString: string) {
  const headers = csvString.slice(0, csvString.indexOf("\n")).split(",");
  const rows = csvString.slice(csvString.indexOf("\n") + 1).split("\n");
  const arr = rows.map(function (row) {
    const values = row.split(",");
    const el: { [key: string]: string | number } = headers.reduce(function (
      object: { [key: string]: string | number },
      header,
      index
    ) {
      object[header] = values[index];
      return object;
    },
    {});
    return el as UpdateCSV;
  });
  return arr.filter((value) => {
    return value.name !== "";
  });
}

async function handleUpload(files: File[]) {
  processing.value = true;
  uploadedCSV.value = [];
  for (const file of files) {
    await parseCSV(file).catch((error) => {
      console.log(error);
    });
  }
  for (const csv of uploadedCSV.value) {
    updateMutation
      .executeMutation({
        verificationCode: csv.verification_code,
        license: csv.license,
        licenseRationale: csv.license_rationale,
      })
      .then((result) => {
        if(result.error){
          dialogMessage.value = "Error processing uploaded CSV, please check formatting."
        }
        else{
          dialogMessage.value = "Uploaded CSV has completed processing."
        }
        processing.value = false;
        showDialog.value = true;
      })
      .catch((error) => {
        console.log(error);
      });
  }
}
</script>
