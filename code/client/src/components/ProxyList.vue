<template lang="pug">
.card.p-3
  .d-flex.justify-content-between.flex-wrap.mb-3
    b-button-group
      b-button(variant="primary", @click="onNew")
        | New
        b-icon(icon="plus")
      b-button(variant="secondary", @click="onImport") 
        | Import
        b-icon(icon="arrow-down")
      b-button(variant="success", @click="onExport") 
        | Export
        b-icon(icon="arrow-up")
      b-button(variant="info", @click="onExportSelected") 
        | Export Selected
        b-icon(icon="check")
      b-button(variant="danger", @click="onExportSelected") 
        | Delete Selected
        b-icon(icon="trash")
    b-form(inline)
      b-input-group(size="sm", class="flex-grow-1")
        b-form-input(v-model="search", placeholder="Search by name, tag, port, phone, operator...", size="sm")
        template(#append)
          b-input-group-text
            b-icon(icon="search")
      b-form-select(v-model="filters.port", :options="portOptions", class="mr-2", size="sm", placeholder="Port")
      b-form-select(v-model="filters.tag", :options="tagOptions", class="mr-2", size="sm", placeholder="Tag")
      b-form-select(v-model="filters.operator", :options="operatorOptions", class="mr-2", size="sm", placeholder="Operator")

  b-table(:items="filteredData" :fields="fields" small responsive="sm" selectable select-mode="multi")
    template(#cell(actions)="data")
      b-button(size="sm", variant="primary", @click="() => onVerify(data.item)") Verify
      b-button(size="sm", variant="warning", @click="() => onChange(data.item)") Change
      b-button(size="sm", variant="primary", @click="() => onChange(data.item)") History
      b-button(size="sm", variant="info", @click="() => onReset(data.item)") Details
      b-button(size="sm", variant="danger", @click="() => onDelete(data.item)") Delete
    template(#cell(selected)='{ rowSelected }')
      template(v-if='rowSelected')
        b-icon(icon='check')
      template(v-else='')
        span(aria-hidden='true') &nbsp;

  //- Pagination
  .d-flex.justify-content-between.align-items-center.mt-3
    small {{ filteredData.length }} results
    b-pagination(v-model="currentPage", :total-rows="filteredData.length", :per-page="perPage", size="sm")
</template>

<script>
export default {
  name: 'ProxyList',
  data() {
    
    return {
      search: '',
      filters: { port: '', tag: '', operator: '' },
      currentPage: 1,
      perPage: 10,
      fields: [
        { key: 'selected', label: '', selectable: true },
        { key: 'name', label: 'Name' },
        { key: 'phone', label: 'Phone' },
        { key: 'ipPort', label: 'IP:Port' },
        { key: 'realIP', label: 'RealIP' },
        { key: 'username', label: 'UserName' },
        { key: 'password', label: 'Password' },
        { key: 'operator', label: 'Operator' },
        { key: 'tag', label: 'Tag' },
        { key: 'status', label: 'Status' },
        { key: 'latency', label: 'Latency' },
        { key: 'failures', label: 'Failures' },
        { key: 'uptime', label: 'Uptime' },
        { key: 'actions', label: 'Actions' }
      ],
      items: [
        { name: 'Server A', phone: '+1234567', ipPort: '192.168.1.1:8080', operator: 'Admin', tag: 'prod', status: 'OK' },
        { name: 'Server B', phone: '+7654321', ipPort: '192.168.1.2:9090', operator: 'User', tag: 'test', status: 'Error' }
      ],
      portOptions: [ { value: '', text: 'Port' }, { value: '8080', text: '8080' }, { value: '9090', text: '9090' } ],
      tagOptions: [ { value: '', text: 'Tag' }, { value: 'prod', text: 'prod' }, { value: 'test', text: 'test' } ],
      operatorOptions: [ { value: '', text: 'Operator' }, { value: 'Admin', text: 'Admin' }, { value: 'User', text: 'User' } ]
    }
  },
  computed: {
    filteredData() {
      return this.items.filter(item => {
        const query = this.search.toLowerCase()
        const matchesSearch = [item.name, item.tag, item.ipPort, item.phone, item.operator]
          .some(field => field.toLowerCase().includes(query))
        const matchesPort = !this.filters.port || item.ipPort.includes(this.filters.port)
        const matchesTag = !this.filters.tag || item.tag === this.filters.tag
        const matchesOperator = !this.filters.operator || item.operator === this.filters.operator
        return matchesSearch && matchesPort && matchesTag && matchesOperator
      })
    }
  },
  methods: {
    onNew() { alert('New clicked') },
    onImport() { alert('Import clicked') },
    onExport() { alert('Export clicked') },
    onExportSelected() { alert('Export selected clicked') },
    onVerify(item) { alert('Verify: ' + item.name) },
    onChange(item) { alert('Change: ' + item.name) },
    onReset(item) { alert('Reset: ' + item.name) },
    onDelete(item) { alert('Delete: ' + item.name) }
  }
}
</script>

<style scoped>
.gap-2 > * {
  margin-right: 0.5rem;
}
</style>
