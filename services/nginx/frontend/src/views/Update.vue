<!-- Package file upload page -->
<template>
  <div class="d-flex justify-center">
    <div class="d-flex flex-column w-75 mt-12">
      <h2 class="text-blue-grey-darken-3">Upload CSV</h2>
      <component
        :is="Modal"
        v-if="showModal"
        @close="showModal = false"
        :payload="modalPayload"
      >
        <template v-slot:header>
          <h2>Processing Failed</h2>
          <p>{{ modalMessage }}</p>
        </template>
        <component :is="CSV" v-if="modalCSV" :text="modalData" />
        <div v-else>{{ modalData }}</div>
      </component>
      <div id="drag-drop" />
      <v-row class="mt-4 justify-center">
        <v-btn
          @click="resetPressed"
          width="200"
          color="blue-grey-darken-3"
          class="mr-1"
          >Reset</v-btn
        >
      </v-row>
      <div class="align-self-center mt-1" v-if="false">
        <v-checkbox
          id="autobox"
          v-model="autoUpload"
          label="Auto-Process"
        ></v-checkbox>
      </div>
      <br />
      <div v-if="csv.length > 0">
        <h4 v-if="csvError.length == 0">CSV Processed: {{ csv }}</h4>
        <h4 v-else style="color: red">CSV Not Processed: {{ csv }}</h4>
      </div>
      <p style="color: red">{{ csvError }}</p>

      <table class="no-border mt-6">
        <tr class="no-border">
          <td class="no-border">
            <div v-if="elementsExist" class="file-list">
              <template v-if="processing">
                <h4 :style="{ display: 'inline-block' }">Processing Files</h4>
                <div
                  :style="{
                    display: 'inline-block',
                    height: '20px',
                    width: '20px',
                  }"
                  class="spinner"
                />
              </template>
              <h4 v-else>Uploaded Files {{ files.length }}</h4>
              <ul>
                <li v-for="(file, index) in files" :key="index">
                  {{ file.filename }}
                </li>
              </ul>
            </div>
          </td>

          <td class="no-border">
            <div v-if="elementsExist" class="upload-list">
              <h4>Upload Progress {{ elementsLength }} Left</h4>
              <ul>
                <li v-for="(element, index) in incompleteUploads" :key="index">
                  {{ element.data.fullPath }}: {{ element.perc }}%
                </li>
              </ul>
            </div>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Ref, ref, computed, onMounted, onBeforeUnmount } from "vue";

import Modal from "@/components/Modal.vue";
import CSV from "@/components/CSV.vue";
import download from "downloadjs";

import Uppy from "@uppy/core";
import XHRUpload from "@uppy/xhr-upload";
import DrapDrop from "@uppy/drag-drop";

import "@uppy/core/dist/style.css";
import "@uppy/drag-drop/dist/style.css";

type file = {
  id: string;
  filename: string;
  sha1: string;
  uploadName: string;
  contentType: string;
  header: {
    Filename: string;
    Header: Record<string, string[]>;
  };
};
type XHRUploadResponse = {
  status: number;
  body: {
    Filename: string;
    Uploadname: string;
    Sha1: string;
    isMeta: boolean;
    "Content-Type": string;
    Header: {
      Filename: string;
      Header: {
        "Content-Disposition": string[];
        "Content-Type": string[];
      };
    };
    Extra?: string;
  };
  uploadURL?: string;
};

type XHRUploadProgress = {
  bytesUploaded: number;
  bytesTotal: number;
  uploadStarted: null | number; // null or UNIX timestamp
  percentage: number; // Integer [0, 100]
};

interface UppyFilePerc extends Uppy.UppyFile {
  perc?: string;
}

type ModalPayload = {
  data: string;
  filename: string;
  mime: string;
};

const files: Ref<file[]> = ref([]);
const uploadList: Ref<UppyFilePerc[]> = ref([]);
const uploadDict: Ref<Record<string, number>> = ref({});
const autoUpload: Ref<boolean> = ref(false);
const processing: Ref<boolean> = ref(false);
const newFile: Ref<boolean> = ref(false);

const modalMessage: Ref<string> = ref("");
const modalData: Ref<null | string> = ref(null);
const modalPayload: Ref<ModalPayload | null> = ref(null);
const modalCSV: Ref<boolean> = ref(false);
const showModal: Ref<boolean> = ref(false);

const csv: Ref<string> = ref("");
const csvError: Ref<string> = ref("");

const uppy: Ref<Uppy.Uppy<"strict"> | null> = ref(null);

const processingAllowed = computed(function getProcessingAllowed(): boolean {
  return files.value.length > 0 && !processing.value;
});

const uploadElements = computed(function getUploadElements(): UppyFilePerc[] {
  return uploadList.value
    .filter(function (value: UppyFilePerc): boolean {
      return value.progress?.percentage !== 1;
    })
    .map(function (value: UppyFilePerc): UppyFilePerc {
      const perc = ((value.progress?.percentage || 0) * 100).toFixed(0);
      value.perc = perc;
      return value;
    });
});

const elementsLength = computed(function getElementsLength(): number {
  return uploadElements.value.length;
});

const elementsExist = computed(function getElementsExist(): boolean {
  return files.value.length > 0 || uploadElements.value.length > 0;
});

const incompleteUploads = computed(
  function getIncompleteUploads(): UppyFilePerc[] {
    return uploadElements.value.filter((value: UppyFilePerc): boolean => {
      return value.progress?.percentage !== 1;
    });
  }
);

function process(): void {
  if (!processing.value && files.value.length > 0) {
    processing.value = true;
    var ok = true;
    // var status = 200
    fetch(
      new Request("/api/upload/process", {
        method: "POST",
        mode: "same-origin",
        body: JSON.stringify(files.value),
      })
    )
      .catch((error) => {
        // Error making fetch request
        alert(JSON.stringify(error));
        processing.value = false;
      })
      .then((response) => {
        // Process fetch response
        if (response instanceof Response) {
          ok = response.ok;
          return response.text();
        } else {
          processing.value = false;
        }
      })
      .then((text) => {
        if (ok && typeof text === "string") {
          download(text, "tk-prefilled.csv", "text/csv");
          reset();
        } else {
          modalData.value = text || "null";
          modalPayload.value = {
            data: modalData.value,
            filename: "tk-error.csv",
            mime: "text/csv",
          };
          modalMessage.value =
            "Errors encountered processing the following files";
          modalCSV.value = true;
          showModal.value = true;
        }

        processing.value = false;
      })
      .catch((result) => {
        alert(JSON.stringify(result));
        processing.value = false;
      });
  }
}

function reset(): void {
  console.log("this.reset()");
  files.value = [];
  uploadList.value = [];
  uploadDict.value = {};
  // leave this.autoUpload unchanged
  processing.value = false;
  newFile.value = false;
  csv.value = "";
  csvError.value = "";

  if (uppy.value instanceof Uppy.Uppy) {
    uppy.value.reset();
  }
}

function processPressed(): void {
  process();
}

function resetPressed(): void {
  reset();
}

function debug(): void {
  if (uppy.value instanceof Uppy.Uppy) {
    console.log(uppy.value.getState());
  }
}

onMounted(function (): void {
  if (uppy.value === null) {
    mountUppy().run();
  } else {
    uppy.value.run();
  }
});

function mountUppy(): Uppy.Uppy<"strict"> {
  uppy.value = Uppy<Uppy.StrictTypes>({
    meta: {
      type: "binary",
    },
    autoProceed: true,
  });

  uppy.value.use(XHRUpload, {
    endpoint: "/api/upload",
    method: "post",
    formData: true,
    fieldName: "file",
    timeout: 0,
    limit: 3,
  });

  uppy.value.use(DrapDrop, {
    target: "#drag-drop",
    width: "100%",
    height: "100%",
    locale: {
      strings: {
        dropHereOr: "Drop CSV here or %{browse}",
        browse: "browse",
      },
    },
  });

  uppy.value.on("file-added", (file: Uppy.UppyFile): void => {
    const ind = uploadList.value.push(file) - 1;
    uploadDict.value[file.id] = ind;
  });

  uppy.value.on("complete", (result) => {
    if (files.value.length === uploadList.value.length) {
      if (uppy.value instanceof Uppy.Uppy) {
        uppy.value.reset();
        console.log("this.uppy.reset()");
      } else {
        console.error("cannot reset null uppy");
      }
      if (newFile.value && autoUpload.value) {
        newFile.value = false;
        process();
      }
    }
  });

  uppy.value.on(
    "upload-success",
    (file: Uppy.UppyFile, response: XHRUploadResponse) => {
      // console.log(`upload-success(${file.id}, ${JSON.stringify(response)})`)
      if (response.body.isMeta === false) {
        const f = {
          id: file.id,
          filename: response.body.Filename,
          sha1: response.body.Sha1,
          uploadName: response.body.Uploadname,
          contentType: response.body["Content-Type"],
          header: response.body.Header,
        };

        files.value.push(f);
        newFile.value = true;
      } else {
        if (Object.prototype.hasOwnProperty.call(response.body, "Extra")) {
          modalData.value = null; // this.modalData = 'test'
          modalPayload.value = {
            data: response.body.Extra || "",
            filename: "tk-csv-error.csv",
            mime: "text/csv",
          };
          modalMessage.value =
            "Errors were encountered while updating. Details can be found by Downloading the attatched CSV file.";
          modalCSV.value = false;
          showModal.value = true;
        }

        csv.value = response.body.Filename;
        console.log(`Set csv to ${csv.value}`);
      }
    }
  );

  uppy.value.on(
    "upload-progress",
    (file: Uppy.UppyFile, progress: XHRUploadProgress) => {
      const f = uploadList.value[uploadDict.value[file.id]];
      if (f.progress === undefined) {
        f.progress = {
          bytesTotal: 0,
          bytesUploaded: 0,
          percentage: 0,
          uploadStarted: null,
          uploadComplete: false,
        };
      }

      f.progress.bytesTotal = progress.bytesTotal;
      f.progress.bytesUploaded = progress.bytesUploaded;
      f.progress.percentage = f.progress.bytesUploaded / f.progress.bytesTotal;

      if (f.progress.percentage === 1) {
        f.progress.uploadComplete = true;
      }
    }
  );

  return uppy.value;
}

onBeforeUnmount(function (): void {
  if (uppy.value !== null) {
    uppy.value.close();
    // this.uppy = null
  }
});
</script>

<style lang="scss">
.spinner {
  -webkit-animation: spin 1.5s linear infinite;
  animation: spin 1.5s linear infinite;
  border: 3px solid #ddd;
  border-top-width: 3px;
  border-top-style: solid;
  border-radius: 50%;
}

@-webkit-keyframes spin {
  0% {
    border-top-color: #42a5f5;
  }
  50% {
    border-top-color: #ec407a;
  }
  100% {
    border-top-color: #42a5f5;
    -webkit-transform: rotate(360deg);
    transform: rotate(360deg);
  }
}

@keyframes spin {
  0% {
    border-top-color: #42a5f5;
  }
  50% {
    border-top-color: #ec407a;
  }
  100% {
    border-top-color: #42a5f5;
    -webkit-transform: rotate(360deg);
    transform: raotate(360deg);
  }
}

table {
  width: 100%;
}

.no-border {
  border: none !important;
}
td.no-border {
  vertical-align: top;
  width: 50%;
}

.uppy-DragDrop-arrow {
  width: 60px;
  height: 60px;
  fill: black; // fill: lighten(gray, 30%);
  margin-bottom: 17px;
}

.uppy-DragDrop-note {
  font-size: 1em;
  //color: lighten(gray, 10%);
  color: black;
}
</style>
