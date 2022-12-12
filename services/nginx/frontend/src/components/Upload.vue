<template>
  <v-col class="d-flex flex-column align-center">
    <v-row class="w-75 ma-6">
      <v-file-input
        :label="message"
        :accept="type"
        multiple
        v-model="files"
        variant="outlined"
      >
        <template v-slot:selection="{ fileNames }">
          <template v-for="fileName in fileNames" :key="fileName">
            <v-chip size="large" label color="primary" class="mr-2">
              {{ fileName }}
            </v-chip>
          </template>
        </template>
      </v-file-input>
    </v-row>
    <v-row>
      <v-btn width="150" color="primary" @click="emit('sendFiles', files)">
        Upload
      </v-btn>
      <v-progress-circular
        indeterminate
        color="primary"
        class="mx-4"
        v-if="processing"
      ></v-progress-circular>
    </v-row>
  </v-col>
</template>

<script setup lang="ts">
import { Ref, ref } from "vue"

const files: Ref<File[]> = ref([])
const props = defineProps<{
  message: string
  type: string
  processing: boolean
}>()

const emit = defineEmits<{
  (event: "sendFiles", files: File[]): void
}>()
</script>
