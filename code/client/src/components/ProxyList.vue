<template lang="pug">
.card.p-3
  //- Toolbar and Filters
  .d-flex.justify-content-between.flex-wrap.mb-3
    b-button-group
      b-button(variant="primary", @click="onNew")
        | New
      b-button(variant="secondary", @click="onImport") 
        | Import
      b-button(variant="success", @click="onExport") 
        | Export
      b-button(variant="info", @click="onExportSelected") 
        | Export Selected
      b-button(variant="info", @click="onVerifySelected") 
        | Verify Selected
      b-button(variant="danger", @click="onDeleteSelected") 
        | Delete Selected
      b-button(@click="togglePasswordVisibility")
        | {{ passwordsVisible ? 'Hide' : 'Show' }} Passwords
    b-form(inline)
      b-input-group(size="sm", class="flex-grow-1")
        b-form-input(v-model="search", placeholder="Search...", size="sm")
        template(#append)
          b-input-group-text
            b-icon(icon="search")
      b-form-select(v-model="filters.port", :options="portOptions", class="mr-2", size="sm")
      b-form-select(v-model="filters.operator", :options="operatorOptions", class="mr-2", size="sm")
      b-form-select(v-model="filters.realCountry", :options="countryOptions", class="mr-2", size="sm")

  //- Proxies Table
  b-table(
    :items="filteredData" 
    :fields="fields" 
    :current-page="currentPage"
    :per-page="perPage"
    small 
    responsive="sm" 
    selectable 
    select-mode="multi"
    ref="selectableTable"
    @row-selected="onRowSelected"
  )
    template(#cell(index)="{ index }")
      span {{ (currentPage - 1) * perPage + index + 1 }}
    template(#cell(password)="data")
      span(v-if="passwordsVisible") {{ data.item.password }}
      span(v-else) ******
    template(#cell(lastStatus)="data")
      b-badge(v-if="data.item.lastStatus === 1", variant="success") Ok
      b-badge(v-else-if="data.item.lastStatus === 2", variant="danger") Error
      b-badge(v-else-if="data.item.lastStatus === 3", variant="warning") Stuck
      span(v-else) -
    template(#cell(uptime)="data")
      span {{ formatUptime(data.item.uptime) }}
    template(#cell(realIP)="data")
      div.d-flex.justify-content-between.align-items-center
        span {{ data.item.realIP }}
        router-link(:to="{ path: '/ip_logs', query: { proxy_id: data.item.id } }", target="_blank")
          b-button(size="sm", variant="outline-secondary")
            b-icon(icon="clock-history")
    template(#cell(speed)="data")
      div.d-flex.justify-content-between.align-items-center
        span {{ formatSpeed(data.item.speed) }}
        router-link(:to="{ path: '/speed_logs', query: { proxy_id: data.item.id } }", target="_blank")
          b-button(size="sm", variant="outline-secondary")
            b-icon(icon="clock-history")
    template(#cell(upload)="data")
      span {{ formatSpeed(data.item.upload) }}
    template(#cell(actions)="data")
      router-link(:to="{ path: '/visit_logs', query: { proxy_id: data.item.id } }", target="_blank")
        b-button(size="sm", variant="outline-secondary", class="mr-1")
          b-icon(icon="clock-history")
      b-button(size="sm", variant="warning", @click="() => onChange(data.item)")
        b-icon(icon="pencil-square")
      b-button(size="sm", variant="danger", @click="() => onDelete(data.item)")
        b-icon(icon="trash")
    template(#cell(selected)='{ rowSelected }')
      template(v-if='rowSelected')
        b-icon(icon='check')
      template(v-else='')
        span(aria-hidden='true') &nbsp;

  //- Pagination and Info
  .d-flex.justify-content-between.align-items-center.mt-3
    small {{ filteredData.length }} results
    b-pagination(v-model="currentPage", :total-rows="filteredData.length", :per-page="perPage", size="sm")

  ProxySettingsModal(v-model="newModalShown", title="New Proxy", @proxy-created="handleProxyCreated", @hidden="onModalHidden")
  ProxySettingsModal(v-model="editModalShown", title="Edit Proxy", :proxy="proxyToEdit", @proxy-updated="handleProxyUpdated", @hidden="onModalHidden")
  
  //- Скрытый инпут для выбора файла
  input(type="file", ref="fileInput", @change="handleFileSelected", style="display: none",)

</template>

<script>
import ProxySettingsModal from '@/components/ProxySettingsModal.vue';
import { getProxies,verifyProxy, verifyBatch, deleteProxy, importProxies, exportAllProxies, exportSelectedProxies } from '@/api/proxy.js';

export default {
  name: 'ProxyList',
  components:{
    ProxySettingsModal
  },
  data() {
    return {
      search: '',
      proxies: [], // Для хранения оригинальных данных с API
      selectedRows: [],
      newModalShown: false,
      editModalShown: false,
      proxyToEdit: null,
      pendingProxy: null, // Временно храним прокси для обновления списка
      passwordsVisible: false,
      filters: { port: null, operator: null, realCountry: null },
      currentPage: 1,
      perPage: 100,
      fields: [
        { key: 'index', label: '#'},
        { key: 'selected', label: '', selectable: true },
        { key: 'name', label: 'Name', sortable: true },
        { key: 'ip', label: 'Local IP', sortable: true },
        { key: 'port', label: 'Port', sortable: true },
        { key: 'phone', label: 'Phone', sortable: true },
        { key: 'contacts', label: 'Contacts', sortable: true },
        { key: 'realIP', label: 'Real IP', sortable: true },
        { key: 'realCountry', label: 'Real Country', sortable: true },
        { key: 'username', label: 'Username' },
        { key: 'password', label: 'Password' },
        { key: 'operator', label: 'Operator', sortable: true },
        { key: 'lastIPChange', label: 'Last IP Change', sortable: true },
        { key: 'uptime', label: 'Uptime', sortable: true },
        { key: 'lastStatus', label: 'Status', sortable: true },
        { key: 'lastLatency', label: 'Latency (ms)', sortable: true },
        { key: 'speed', label: 'Download (mb/s)', sortable: true },
        { key: 'upload', label: 'Upload (mb/s)', sortable: true },
        { key: 'failures', label: 'Failures', sortable: true },
        { key: 'actions', label: 'Actions' }
      ],
    }
  },
  computed: {
    // Динамически создаем опции для фильтров на основе загруженных данных
    portOptions() {
      const ports = [...new Set(this.proxies.map(p => p.port))];
      const options = ports.map(port => ({ value: port, text: port }));
      return [{ value: null, text: 'Port' }, ...options];
    },
    operatorOptions() {
      const operators = [...new Set(this.proxies.map(p => p.operator).filter(Boolean))];
      const options = operators.map(op => ({ value: op, text: op }));
      return [{ value: null, text: 'Operator' }, ...options];
    },
    countryOptions() {
      const countries = [...new Set(this.proxies.map(p => p.realCountry).filter(Boolean))];
      const options = countries.map(country => ({ value: country, text: country }));
      return [{ value: null, text: 'Country' }, ...options];
    },
    filteredData() {
      const oneDay = 24 * 60 * 60 * 1000; // 24 часа в миллисекундах
      const now = new Date();

      const filtered = this.proxies.filter(item => {
        let matched = true;
        const query = this.search.toLowerCase();

        if (query) {
          const searchIn = [
            item.name,
            item.ip,
            item.port,
            item.phone,
            item.operator,
            item.realIP,
            item.username,
            item.contacts,
          ];
          matched = searchIn.some(field => field && field.toLowerCase().includes(query));
        }

        if (matched && this.filters.port) {
          matched = item.port === this.filters.port;
        }

        if (matched && this.filters.operator) {
          matched = item.operator === this.filters.operator;
        }

        if (matched && this.filters.realCountry) {
          matched = item.realCountry === this.filters.realCountry;
        }

        return matched;
      });

      return filtered.map(item => {
        const newItem = { ...item };
        if (newItem.lastIPChange) {
          const lastChangeDate = new Date(newItem.lastIPChange);
          // Проверяем, прошло ли больше 24 часов
          if (now - lastChangeDate > oneDay) {
            newItem._rowVariant = 'danger'; // Используем 'danger' для красного цвета
          }
          // Рассчитываем uptime в минутах
          const uptimeInMinutes = Math.floor((now - lastChangeDate) / (1000 * 60));
          newItem.uptime = uptimeInMinutes;
        } else {
          newItem.uptime = null;
        }
        return newItem;
      });
    }
  },
  methods: {
    formatSpeed(kbps) {
      if (kbps === null || isNaN(kbps) || kbps === 0) {
        return '-';
      }
      const mbps = kbps / 1000;
      return mbps.toFixed(2);
    },
    formatUptime(minutes) {
      if (minutes === null || isNaN(minutes)) {
        return '-';
      }
      if (minutes <= 6) {
        return `${minutes}m`;
      }

      const days = Math.floor(minutes / (60 * 24));
      const hours = Math.floor((minutes % (60 * 24)) / 60);
      const mins = minutes % 60;

      let result = '';
      if (days > 0) {
        result += `${days}d `;
      }
      if (hours > 0) {
        result += `${hours}h `;
      }
      // Показываем минуты, если нет других единиц или они не равны нулю
      if (mins > 0 || result === '') {
        result += `${mins}m`;
      }

      return result.trim();
    },
    togglePasswordVisibility() {
      this.passwordsVisible = !this.passwordsVisible;
    },
    handleProxyCreated(newProxy) {
      // Временно сохраняем новый прокси, не обновляя основной список
      this.pendingProxy = newProxy;
    },
    handleProxyUpdated(updatedProxy) {
      // Временно сохраняем обновленный прокси
      this.pendingProxy = updatedProxy;
    },
    onNew() { this.newModalShown = true } ,
    onImport() {
      // Открываем диалог выбора файла
      this.$refs.fileInput.click();
    },
    async handleFileSelected(event) {
      const file = event.target.files[0];
      if (!file) {
        return;
      }

      try {
        const result = await importProxies(file);
        alert(`Import finished. Imported: ${result.importedCount}, Failed: ${result.failedCount}`);
        this.proxies = await getProxies(); // Обновляем список
      } catch (error) {
        console.error('Failed to import proxies:', error);
        alert('Failed to import proxies.');
      } finally {
        // Сбрасываем значение инпута, чтобы можно было выбрать тот же файл снова
        event.target.value = '';
      }
    },
    async onExport() {
      try {
        await exportAllProxies();
      } catch (error) {
        console.error('Failed to export all proxies:', error);
        alert('Failed to export proxies.');
      }
    },
    async onExportSelected() {
      const ids = this.selectedRows.map(item => item.id);
      if (ids.length === 0) {
        alert('No proxies selected for export.');
        return;
      }
      try {
        await exportSelectedProxies(ids);
      } catch (error) {
        console.error('Failed to export selected proxies:', error);
        alert('Failed to export selected proxies.');
      }
    },
    async onVerifySelected() {
      const ids = this.selectedRows.map(item => item.id);
      if (ids.length === 0) {
        alert('No proxies selected');
        return;
      }
      try {
        await verifyBatch(ids);
        alert('Proxies verified successfully');
        this.proxies = await getProxies();
        this.$refs.selectableTable.clearSelected();
      } catch (error) {
        console.error('Failed to verify selected proxies:', error);
        alert('Failed to verify selected proxies.');
      }
    },
    async onDeleteSelected() {
      const ids = this.selectedRows.map(item => item.id);
      if (ids.length === 0) {
        alert('No proxies selected');
        return;
      }

      if (!confirm(`Are you sure you want to delete ${ids.length} selected proxies?`)) {
        return;
      }

      try {
        for (const id of ids) {
          await deleteProxy(id);
        }
      } catch (error) {
        console.error('An error occurred during batch deletion:', error);
        alert('An error occurred while deleting proxies. The operation was interrupted.');
      } finally {
        this.proxies = await getProxies();
        this.$refs.selectableTable.clearSelected();
      }
    },
    async onVerify(item) {
      try {
        await verifyProxy(item.id);
        this.proxies = await getProxies();
      } catch (error) {
        console.error(`Failed to verify proxy ${item.id}:`, error);
        alert(`Failed to verify proxy ${item.ip}:${item.port}.`);
      }
    },
    onChange(item) {
      this.proxyToEdit = item;
      this.editModalShown = true;
    },
    onReset(item) { alert('Reset: ' + item.ipPort) },
    async onDelete(item) {
      if (!confirm(`Are you sure you want to delete proxy ${item.ip}:${item.port}?`)) {
        return;
      }
      try {
        await deleteProxy(item.id);
        this.proxies = await getProxies();
      } catch (error) {
        console.error(`Failed to delete proxy ${item.id}:`, error);
        alert(`Failed to delete proxy ${item.ip}:${item.port}.`);
      }
    },
    onRowSelected(items) {
      this.selectedRows = items;
    },
    onModalHidden() {
      this.$nextTick(() => {
        if (this.pendingProxy) {
          const index = this.proxies.findIndex(p => p.id === this.pendingProxy.id);
          if (index !== -1) {
            // Редактирование
            this.$set(this.proxies, index, this.pendingProxy);
          } else {
            // Создание
            this.proxies.unshift(this.pendingProxy);
          }
        }
        
        // Сбрасываем все временные состояния
        this.proxyToEdit = null;
        this.pendingProxy = null;
      });
    }
  },
   async created() {
    const proxies = await getProxies();
    this.proxies = proxies;
  }
}
</script>

<style scoped>
.gap-2 > * {
  margin-right: 0.5rem;
}
</style>
