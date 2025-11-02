<template lang="pug">
  div
    b-cart(title="Proxy List").m-2
      LogList

  
</template>

<script>
import LogList from '@/components/LogList.vue'

export default {
  name: 'LogView',
  components: {
    LogList,
  }
}
</script>