<template lang="pug">
  div
    b-card( bg-variant="light").m-2
      ProxySettings
    b-cart(title="Proxy List").m-2
      ProxyList

  
</template>

<script>
import ProxySettings from '@/components/ProxySettings.vue'
import ProxyList from '@/components/ProxyList.vue';

export default {
  name: 'ProxyView',
  components: {
    ProxySettings,
    ProxyList
  }
}
</script>
