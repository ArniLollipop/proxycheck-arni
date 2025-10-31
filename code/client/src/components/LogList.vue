<template lang="pug">
  b-container(fluid)
    // Toolbar
    b-row.mb-2(align-h="between")
      b-col(cols="auto")
        b-button-group
          b-button(variant="success" size="sm" @click="onNew") Process Logs 
          b-button(variant="info" size="sm" @click="onExport") Export
      b-col(cols="auto")
        b-input-group(size="sm" class="w-100")
          b-form-input(v-model="search" placeholder="Search by  IP, domain" type="search")
          b-input-group-append
            b-button(variant="secondary" size="sm" @click="search = ''") Clear
          b-form-select(v-model="filters.proxyName" :options="proxyList")
    // Table
    b-table(:items="filteredLogs" :fields="fields" small responsive striped hover)
      template(#cell(actions)="row")
        b-button(size="sm" variant="info" @click="viewRaw(row.item)") View RAW Log

    // Pagination
    b-pagination(v-model="currentPage" :per-page="perPage" :total-rows="filteredLogs.length" align="center" size="sm" class="my-2")
</template>

<script>
export default {
  data() {
    return {
      search: '',
      filters: {
        proxyName: '',
        ip: '',
        tag: ''
      },
      perPage: 10,
      currentPage: 1,
      logs: [
        {
          datetime: '2025-10-31 09:00:00',
          sourceIP: '192.168.1.1',
          targetIP: '10.0.0.1',
          port: '443',
          domain: 'example.com',
          proxyName: 'Proxy-1',
          tag: 'Main'
        },
        {
          datetime: '2025-10-31 09:10:12',
          sourceIP: '192.168.1.2',
          targetIP: '10.0.0.2',
          port: '8080',
          domain: 'test.org',
          proxyName: 'Proxy-2',
          tag: 'Backup'
        }
      ],
      proxyList: [ { value: '', text: 'Proxy' }, { value: '8080', text: 'Proxy1' }, { value: '9090', text: 'Proxy2' } ],
      fields: [
        { key: 'datetime', label: 'Date/Time', sortable: true },
        { key: 'sourceIP', label: 'Source IP', sortable: true },
        { key: 'targetIP', label: 'Target IP', sortable: true },
        { key: 'port', label: 'Port', sortable: true },
        { key: 'domain', label: 'Domain', sortable: true },
        { key: 'actions', label: 'Actions' }
      ]
    }
  },
  computed: {
    filteredLogs() {
      const s = this.search.toLowerCase()
      return this.logs.filter(l => {
        return (
          (!this.filters.proxyName || l.proxyName.toLowerCase().includes(this.filters.proxyName.toLowerCase())) &&
          (!this.filters.ip || l.sourceIP.includes(this.filters.ip) || l.targetIP.includes(this.filters.ip)) &&
          (!this.filters.tag || l.tag.toLowerCase().includes(this.filters.tag.toLowerCase())) &&
          (!s ||
            l.proxyName.toLowerCase().includes(s) ||
            l.sourceIP.includes(s) ||
            l.targetIP.includes(s) ||
            l.domain.toLowerCase().includes(s) ||
            l.tag.toLowerCase().includes(s))
        )
      })
    }
  },
  methods: {
    viewRaw(item) {
      alert(`RAW log for ${item.domain} (from ${item.sourceIP} to ${item.targetIP})`)
    },
    onNew() { alert('New log clicked') },
    onImport() { alert('Import clicked') },
    onExport() { alert('Export clicked') },
    onExportSelected() { alert('Export selected clicked') }
  }
}
</script>
