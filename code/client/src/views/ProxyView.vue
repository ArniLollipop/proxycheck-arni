<template lang="pug">
  b-card(title="Settings" bg-variant="light").m-2
    ProxySettings

  
</template>

<script>
import ProxySettings from '@/components/ProxySettings.vue'

export default {
  name: 'ProxyView',
  components: {
    ProxySettings
  }
}
</script>
