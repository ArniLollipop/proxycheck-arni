<template lang="pug">
.card.p-3
  h4 IP Change History
  //- Filters
  b-form.d-flex.justify-content-start.flex-wrap.mb-3(@submit.prevent="applyFilters")
    b-form-group(label="Proxy", label-for="proxy-filter", class="mr-2")
      b-form-select(id="proxy-filter", v-model="filters.proxy_id", :options="proxyOptions", size="sm")
    b-form-group(label="Start Date", label-for="start-date-filter", class="mr-2")
      b-form-datepicker(id="start-date-filter", v-model="filters.start_date", size="sm")
    b-form-group(label="End Date", label-for="end-date-filter", class="mr-2")
      b-form-datepicker(id="end-date-filter", v-model="filters.end_date", size="sm")
    .d-flex.align-items-end
      b-button(type="submit", variant="primary", size="sm", class="mr-2") Apply
      b-button(variant="secondary", size="sm", @click="resetFilters") Reset

  //- IP Logs Table
  b-table(
    :items="logs"
    :fields="fields"
    :busy.sync="isBusy"
    small
    responsive="sm"
    @sort-changed="handleSort"
  )
    template(#cell(proxy_name)="{ item }")
      | {{ getProxyName(item.proxy_id) }}
    template(#cell(timestamp)="{ value }")
      | {{ new Date(value).toLocaleString() }}

  //- Pagination and Info
  .d-flex.justify-content-between.align-items-center.mt-3
    small {{ totalRows }} results
    b-pagination(
      v-model="currentPage"
      :total-rows="totalRows"
      :per-page="perPage"
      size="sm"
      @change="handlePageChange"
    )

</template>

<script>
import { getIpLogs, getProxies } from '@/api/proxy.js';

export default {
  name: 'IpLogList',
  data() {
    return {
      logs: [],
      isBusy: false,
      filters: {
        proxy_id: null,
        start_date: '',
        end_date: ''
      },
      proxyOptions: [{ value: null, text: 'All Proxies' }],
      currentPage: 1,
      perPage: 15,
      totalRows: 0,
      sortBy: 'timestamp',
      sortDesc: true,
      fields: [
        { key: 'proxy_name', label: 'Proxy Name', sortable: false },
        { key: 'timestamp', label: 'Date/Time', sortable: true },
        { key: 'ip', label: 'New IP', sortable: true },
        { key: 'old_ip', label: 'Old IP', sortable: false },
        { key: 'isp', label: 'ISP', sortable: true },
      ]
    };
  },
  async created() {
    const proxyIdFromUrl = this.$route.query.proxy_id;
    if (proxyIdFromUrl) {
      this.filters.proxy_id = proxyIdFromUrl;
    }
    await this.loadProxies();
    await this.fetchLogs();
  },
  methods: {
    async fetchLogs() {
      this.isBusy = true;
      try {
        const params = {
          page: this.currentPage,
          page_size: this.perPage,
          sort_field: this.sortBy,
          sort_desc: this.sortDesc,
          ...this.filters
        };

        // Remove empty filters
        Object.keys(params).forEach(key => {
          if (params[key] === null || params[key] === '') {
            delete params[key];
          }
        });

        const { data, total } = await getIpLogs(params);
        this.logs = data;
        this.totalRows = total;
      } catch (error) {
        alert('Failed to load IP logs.');
      } finally {
        this.isBusy = false;
      }
    },
    getProxyName(proxyId) {
      const proxy = this.proxyOptions.find(p => p.value === proxyId);
      return proxy ? proxy.text : 'Unknown';
    },
    async loadProxies() {
      try {
        const data= await getProxies();
        this.proxyOptions.push(...data.map(p => ({ value: p.id, text: p.name })));
        console.log('Loaded proxies:', data);
      } catch (error) {
        console.error('Failed to load proxies for filter:', error);
      }
    },
    applyFilters() {
      this.currentPage = 1;
      this.fetchLogs();
    },
    resetFilters() {
      this.filters = {
        proxy_id: null,
        start_date: '',
        end_date: ''
      };
      this.applyFilters();
    },
    handlePageChange(page) {
      this.currentPage = page;
      this.fetchLogs();
    },
    handleSort(ctx) {
      this.sortBy = ctx.sortBy;
      this.sortDesc = ctx.sortDesc;
      this.currentPage = 1; // Reset to first page on sort
      this.fetchLogs();
    }
  }
};
</script>

<style scoped>
.card {
  margin-top: 1rem;
}
</style>