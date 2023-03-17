<!-- Reusable component for uploading files -->
<template>
  <v-col class="d-flex flex-column align-center">
    <v-row class="w-100 ma-2">
      <v-file-input
        density="compact"
        :label="message"
        :accept="type"
        :prepend-icon="icon"
        multiple
        v-model="files"
        variant="outlined"
        bg-color="white"
      >
        <template v-slot:selection="{ fileNames }">
          <template v-for="(fileName, index) in fileNames" :key="fileName">
            <v-chip
              v-if="index < 7"
              size="small"
              label
              color="primary"
              class="mr-1"
            >
              {{ fileName.length > 24? fileName.substring(0,6) +".."+ fileName.substring(fileName.length - 7) : fileName}}
            </v-chip>
            <span v-else-if="index === 7">
              +{{ fileNames.length - 7 }} others...
            </span>
          </template>
        </template>
      </v-file-input>
    </v-row>
    <v-row class="align-center">
      <v-btn
        density="compact"
        color="primary"
        @click="emit('sendFiles', files)"
      >
        Upload
      </v-btn>
      <v-progress-circular
        indeterminate
        color="primary"
        class="mx-2"
        :size="20"
        :width="2"
        v-if="processing"
      ></v-progress-circular>
    </v-row>
  </v-col>
</template>

<script setup lang="ts">
import { Ref, ref } from "vue"

const files: Ref<File[]> = ref([])

// Defines the upload component with message, file type, and progress information
const props = defineProps<{
  message: string
  type: string
  icon: string
  processing: boolean
}>()

// Emits when the user activates the form
const emit = defineEmits<{
  (event: "sendFiles", files: File[]): void
}>()
</script>
