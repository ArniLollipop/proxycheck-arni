<template>
  <div class="space-y-6">
    <!-- Normal Textarea -->
    <div>
      <label class="mb-1.5 block text-sm font-medium text-gray-700">
        Description
      </label>
      <textarea
        v-model="normalDescription"
        placeholder="Enter a description..."
        rows="6"
        class=" w-full rounded-lg border border-gray-300 bg-transparent px-4 py-2.5 text-sm text-gray-800 shadow-theme-xs placeholder:text-gray-400 focus:border-brand-300 focus:outline-hidden focus:ring-3 focus:ring-brand-500/10"
      ></textarea>
    </div>

    <!-- Disabled Textarea -->
    <div>
      <label class="mb-1.5 block text-sm font-medium text-gray-300">
        Description
      </label>
      <textarea
        v-model="disabledDescription"
        placeholder="Enter a description..."
        rows="6"
        disabled
        class=" w-full rounded-lg border border-gray-300 bg-transparent px-4 py-2.5 text-sm text-gray-800 shadow-theme-xs placeholder:text-gray-400 focus:border-brand-300 focus:shadow-focus-ring focus:outline-hidden focus:ring-0 disabled:border-gray-100 disabled:bg-gray-50 disabled:placeholder:text-gray-300"
      ></textarea>
    </div>

    <!-- Error State Textarea -->
    <div>
      <label class="mb-1.5 block text-sm font-medium text-gray-700">
        Description
      </label>
      <textarea
        v-model="errorDescription"
        placeholder="Enter a description..."
        rows="6"
        class=" w-full rounded-lg border border-error-300 bg-transparent px-4 py-2.5 text-sm text-gray-800 shadow-theme-xs placeholder:text-gray-400 focus:border-error-300 focus:outline-hidden focus:ring-3 focus:ring-error-500/10"
      ></textarea>
      <p class="mt-1.5 text-theme-xs text-error-500">Please enter a message in the textarea.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const normalDescription = ref('')
const disabledDescription = ref('This textarea is disabled')
const errorDescription = ref('')
</script>
