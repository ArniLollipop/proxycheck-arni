<template lang="pug">
  b-modal(
    :visible="show"
    @change="(val) => $emit('change', val)"
    :title="title"
    @hidden="resetForm"
    @ok="handleOk"
    @cancel="handleCancel"
  )
    form(ref="form" @submit.stop.prevent="handleSubmit")
      b-form-group(label="ID:" v-if="id")
        b-form-input(v-model="id" disabled)
      b-form-group(label="IP:")
        b-form-input(v-model="ip" required)
      b-form-group(label="Port:")
        b-form-input(v-model="port" required)
      b-form-group(label="Username:")
        b-form-input(v-model="username")
      b-form-group(label="Password:")
        b-form-input(v-model="password" type="password")
      b-form-group(label="Name:")
        b-form-input(v-model="name")
      b-form-group(label="Phone:")
        b-form-input(v-model="phone")
      b-form-group(label="Contacts:")
        b-form-textarea(v-model="contacts" max-rows="5" rows="3")
</template>

<script>
import { createProxy, updateProxy } from '@/api/proxy.js';

export default {
  name: 'ProxySettingsModal',
  model: {
    prop: 'show',
    event: 'change'
  },
  props: {
    show: {
      type: Boolean,
      default: false
    },
    title: String,
    proxy: Object
  },
   watch: {
    proxy(newVal) {
      if (newVal) {
        this.ip = newVal.ip || ''
        this.port = newVal.port || ''
        this.username = newVal.username || ''
        this.password = newVal.password || ''
        this.name = newVal.name || ''
        this.phone = newVal.phone || ''
        this.contacts = newVal.contacts || ''
        this.id = newVal.id || ''
      } else {
        this.resetForm();
      }
    }
  },
  data() {
    return {
        ip: '',
        port: '',
        username: '',
        password: '',
        name: '',
        phone: '',
        contacts: '',
        id: ''
    }
  },
  methods: {
    resetForm() {
      this.ip = ''
      this.port = ''
      this.username = ''
      this.password = ''
      this.name = ''
      this.phone = ''
      this.contacts = ''
    },
    handleOk(bvModalEvent) {
      bvModalEvent.preventDefault()
      this.handleSubmit()
    },
    handleCancel() {
      // Просто закрываем модальное окно
      this.$emit('change', false)
    },
    async handleSubmit() {
      const proxyData = {
        ip: this.ip,
        port: this.port,
        username: this.username,
        password: this.password,
        name: this.name,
        phone: this.phone,
        contacts: this.contacts,
      };

      try {
        let savedProxy;
        if (this.proxy && this.proxy.id) {
          // Режим редактирования
          savedProxy = await updateProxy(this.proxy.id, proxyData);
          this.$emit('proxy-updated', savedProxy);
        } else {
          // Режим создания
          savedProxy = await createProxy(proxyData);
          this.$emit('proxy-created', savedProxy);
        }
        // Закрываем модальное окно после успешного сохранения
        this.$emit('change', false);
      } catch (error) {
        console.error('Failed to save proxy:', error);
        alert('Не удалось сохранить прокси. Пожалуйста, проверьте консоль для деталей.');
      }
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>