<template>
  <slot></slot>
</template>

<script setup lang="ts">
import { ref, provide, onMounted, watch, computed } from 'vue'

type Theme = 'light' | 'dark'

const theme = ref<Theme>('light')
const isInitialized = ref(false)

const isDarkMode = computed(() => false) // Always false - dark mode disabled

const toggleTheme = () => {
  // Theme toggle disabled - always use light mode
}

onMounted(() => {
  // Force light theme
  theme.value = 'light'
  document.documentElement.classList.remove('dark')
  localStorage.setItem('theme', 'light')
  isInitialized.value = true
})

watch([theme, isInitialized], ([newTheme, newIsInitialized]) => {
  if (newIsInitialized) {
    // Always force light theme
    localStorage.setItem('theme', 'light')
    document.documentElement.classList.remove('dark')
  }
})

provide('theme', {
  isDarkMode,
  toggleTheme,
})
</script>

<script lang="ts">
import { inject } from 'vue'

export function useTheme() {
  const theme = inject('theme')
  if (!theme) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return theme
}
</script>
