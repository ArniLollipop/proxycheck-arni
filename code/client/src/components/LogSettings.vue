<template lang="pug">
  b-form(inline)
    h4.mr-3 Settings 
    label(for="log-settings-input-interval") Import Interval 
    input(type="number", id="log-settings-input-interval", v-model="logSettings.interval").ml-1
    b-button(variant="primary").ml-2 Save
</template>

<script>
export default {
  name: 'LogSettings',
  data() {
    return {
      logSettings: {
        interval: 24
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
