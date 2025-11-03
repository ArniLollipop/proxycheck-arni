<template lang="pug">
  b-form(inline, @submit.prevent="saveSettings")
    h4.mr-3 Settings
    label.mr-1(for="settings-url") URL
    b-form-input#settings-url.mr-2(type="text", v-model="url", style="width: 200px;")

    label.mr-1(for="settings-timeout") Timeout
    b-form-input#settings-timeout.mr-2(type="number", v-model.number="timeout", style="width: 80px;")

    label.mr-1(for="settings-check-ip-interval") CheckIP Interval
    b-form-input#settings-check-ip-interval.mr-2(type="number", v-model.number="checkIPInterval", style="width: 80px;")

    label.mr-1(for="settings-speed-check-interval") Speed Check Interval
    b-form-input#settings-speed-check-interval.mr-2(type="number", v-model.number="speedCheckInterval", style="width: 80px;")

    b-button(type="submit", variant="primary") Save
</template>

<script>
import { getSettings, updateSettings } from '@/api/settings.js';

export default {
  name: 'ProxySettings',
  data() {
    return {
        url: '',
        timeout: 0,
        checkIPInterval: 0,
        speedCheckInterval: 0,
    }
  },
  async created() {
    await this.loadSettings();
  },
  methods: {
    async loadSettings() {
      try {
        const settings = await getSettings();
        console.log('Loaded settings:', settings);
        // Исправлено: используем camelCase, как в ответе API
        this.url = settings.url;
        this.timeout = settings.timeout;
        this.checkIPInterval = settings.checkIPInterval;
        this.speedCheckInterval = settings.speedCheckInterval;
      } catch (error) {
        console.error('Failed to load settings:', error);
        alert('Failed to load settings.');
      }
    },
    async saveSettings() {
      try {
      
        const settingsToSave = {
          Url: this.url,
          Timeout: this.timeout,
          CheckIPInterval: this.checkIPInterval,
          SpeedCheckInterval: this.speedCheckInterval,
        };
        await updateSettings(settingsToSave);
        alert('Settings saved successfully!');
      } catch (error) {
        console.error('Failed to save settings:', error);
        alert('Failed to save settings.');
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>