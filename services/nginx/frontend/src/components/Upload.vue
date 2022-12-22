<template>
  <v-col class="d-flex flex-column align-center">
    <v-row class="w-100 ma-2">
      <v-file-input
        density="compact"
        :label="message"
        :accept="type"
        multiple
        v-model="files"
        variant="outlined"
        bg-color="white"
      >
        <template v-slot:selection="{ fileNames }">
          <template v-for="fileName in fileNames" :key="fileName">
            <v-chip size="small" label color="primary" class="mr-1">
              {{ fileName }}
            </v-chip>
          </template>
        </template>
      </v-file-input>
    </v-row>
    <v-row class="align-center">
      <v-btn density="compact" color="primary" @click="emit('sendFiles', files)">
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
const props = defineProps<{
  message: string
  type: string
  processing: boolean
}>()

const emit = defineEmits<{
  (event: "sendFiles", files: File[]): void
}>()
</script>
