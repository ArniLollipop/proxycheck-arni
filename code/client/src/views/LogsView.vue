<template lang="pug">
  div
    b-card( bg-variant="light").m-2
      LogSettings
    b-cart(title="Proxy List").m-2
      LogList

  
</template>

<script>
import LogList from '@/components/LogList.vue'
import LogSettings from '@/components/LogSettings.vue'

export default {
  name: 'LogView',
  components: {
    LogList,
    LogSettings,
  }
}
</script>