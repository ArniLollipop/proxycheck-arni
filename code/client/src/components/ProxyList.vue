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
    b-form(inline)
      b-input-group(size="sm", class="flex-grow-1")
        b-form-input(v-model="search", placeholder="Search...", size="sm")
        template(#append)
          b-input-group-text
            b-icon(icon="search")
      b-form-select(v-model="filters.port", :options="portOptions", class="mr-2", size="sm")
      b-form-select(v-model="filters.operator", :options="operatorOptions", class="mr-2", size="sm")

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
    template(#cell(actions)="data")
      b-button(size="sm", variant="primary", @click="() => onVerify(data.item)") Verify
      b-button(size="sm", variant="warning", @click="() => onChange(data.item)") Details/Change
      b-button(size="sm", variant="danger", @click="() => onDelete(data.item)") Delete
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
</template>

<script>
import ProxySettingsModal from '@/components/ProxySettingsModal.vue';
import { getProxies,verifyProxy, verifyBatch, deleteProxy } from '@/api/proxy.js';

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
      filters: { port: null, operator: null },
      currentPage: 1,
      perPage: 15, // Можно настроить
      fields: [
        { key: 'selected', label: '', selectable: true },
        { key: 'name', label: 'Name', sortable: true },
        { key: 'ip', label: 'IP', sortable: true },
        { key: 'port', label: 'Port', sortable: true },
        { key: 'phone', label: 'Phone', sortable: true },
        { key: 'realIP', label: 'Real IP', sortable: true },
        { key: 'username', label: 'Username' },
        { key: 'password', label: 'Password' },
        { key: 'operator', label: 'Operator', sortable: true },
        { key: 'lastStatus', label: 'Status', sortable: true },
        { key: 'lastLatency', label: 'Latency (ms)', sortable: true },
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
    filteredData() {
      return this.proxies.filter(item => {
        let matched  = true
        const query = this.search.toLowerCase()
        matched = query? item.name.toLowerCase().includes(query) ||
                         item.ip.toLowerCase().includes(query) ||
                         item.port.toLowerCase().includes(query) ||
                         item.phone.toLowerCase().includes(query) ||
                         item.operator.toLowerCase().includes(query)
               : true
        matched = this.filters.port? item.port === this.filters.port : matched
        matched = this.filters.operator? item.operator === this.filters.operator : matched
        return matched
      })
    }
  },
  methods: {
    handleProxyCreated(newProxy) {
      // Временно сохраняем новый прокси, не обновляя основной список
      this.pendingProxy = newProxy;
    },
    handleProxyUpdated(updatedProxy) {
      // Временно сохраняем обновленный прокси
      this.pendingProxy = updatedProxy;
    },
    onNew() { this.newModalShown = true } ,
    onImport() { alert('Import clicked') },
    onExport() { alert('Export clicked') },
    onExportSelected() { alert('Export selected clicked') },
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
