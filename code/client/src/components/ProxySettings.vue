<template lang="pug">
  b-form(inline)
    label(for="proxy-settings-input-url") URL 
    input(type="text", id="proxy-settings-input-url", v-model="proxySettings.url").ml-1
    label(for="proxy-settings-input-timeout").ml-2 Timeout 
    input(type="number", id="proxy-settings-input-timeout", v-model="proxySettings.timeout").ml-1
    label(for="proxy-settings-input-threads").ml-2 Threads
    input(type="number", id="proxy-settings-input-threads", v-model="proxySettings.threads").ml-1
    label(for="proxy-settings-input-repeat").ml-2 Repeat
    input(type="number", id="proxy-settings-input-repeat", v-model="proxySettings.repeat").ml-1
    b-button(variant="primary").ml-2 Save
</template>

<script>
export default {
  name: 'ProxySettings',
  data() {
    return {
      proxySettings: {
        url: 'http://example.com',
        timeout: 5000,
        threads: 10,
        repeat: 1
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
