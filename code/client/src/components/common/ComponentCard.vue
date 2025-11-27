<template>
  <div
    :class="[
      'rounded-2xl border border-gray-200 bg-white',
      className,
    ]"
  >
    <!-- Card Header -->
    <div class="px-6 py-5">
      <h3 class="text-base font-medium text-gray-800">
        {{ title }}
      </h3>
      <p v-if="desc" class="mt-1 text-sm text-gray-500">
        {{ desc }}
      </p>
    </div>

    <!-- Card Body -->
    <div class="p-4 border-t border-gray-100 sm:p-6">
      <div class="space-y-5">
        <slot></slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'

interface Props {
  title: string
  className?: string
  desc?: string
}

defineProps<Props>()
</script>
