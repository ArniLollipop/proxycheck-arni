
<template lang="pug">
  div
    b-card(title="Speed History").m-2
      SpeedLogList
</template>

<script>
import SpeedLogList from '@/components/SpeedLogList.vue';

export default {
  name: 'SpeedLogView',
  components: {
    SpeedLogList,
  },
}
</script>
