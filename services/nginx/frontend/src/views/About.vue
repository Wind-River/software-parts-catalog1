<!-- About Page -->
<template>
  <div class="d-flex justify-center">
    <div class="w-75">
      <v-img src="catalog-logo2.png" height="200" class="mt-12" />
      <h4 class="my-6 px-6">
        Maintaining a catalog of software components (parts) is a requirement
        whether generating SBOMs for managing license compliance, security
        assurance, export control, or safety certification. We developed a
        highly searchable scalable software parts database (catalog) enabling
        companies to seamlessly manage 1000s, 10,000s if not 100,000s of
        software parts from which their products are comprised.
      </h4>
      <div class="d-flex justify-center">
        <v-card width="50%">
          <v-card-title class="bg-primary text-center px-8">
            Catalog Name: {{ catalogName }}
          </v-card-title>
          <v-card-text class="mt-4">
            Total Parts Stored: {{ totalParts }}
          </v-card-text>
          <v-divider></v-divider>
          <v-card-text>UI Version: {{ uiVersion.toPrecision(3) }}</v-card-text>
          <v-divider></v-divider>
          <v-card-text>
            Server Version: {{ serverVersion.toPrecision(3) }}
          </v-card-text>
        </v-card>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
const catalogName: string = "https://software-parts.org";
const serverVersion: number = 1.0;
const uiVersion: number = 1.0;
const totalParts: number = 0;
</script>
