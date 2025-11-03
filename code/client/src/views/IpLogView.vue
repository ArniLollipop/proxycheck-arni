<template lang="pug">
  div
    b-card(title="IP Change History").m-2
      IpLogList
</template>

<script>
import IpLogList from '@/components/IpLogList.vue';

export default {
  name: 'IpLogView',
  components: {
    IpLogList,
  },
}
</script>