<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />
    <div class="space-y-5 sm:space-y-6">
      <ComponentCard title="Failure History">
        <div class="space-y-4">
          <!-- Filters -->
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
            <div>
              <label class="mb-2 block text-sm font-medium text-black dark:text-white">Proxy</label>
              <select
                v-model="filters.proxyId"
                @change="fetchLogs"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none dark:border-strokedark dark:bg-meta-4"
              >
                <option value="">All Proxies</option>
                <option v-for="proxy in proxies" :key="proxy.id" :value="proxy.id">
                  {{ proxy.name || proxy.ip }}
                </option>
              </select>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black dark:text-white">Error Type</label>
              <select
                v-model="filters.errorType"
                @change="fetchLogs"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none dark:border-strokedark dark:bg-meta-4"
              >
                <option value="">All Types</option>
                <option value="ping_failed">Ping Failed</option>
                <option value="speed_check_failed">Speed Check Failed</option>
                <option value="ip_check_failed">IP Check Failed</option>
              </select>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black dark:text-white">Start Date</label>
              <input
                v-model="filters.startDate"
                @change="fetchLogs"
                type="date"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none dark:border-strokedark dark:bg-meta-4"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black dark:text-white">End Date</label>
              <input
                v-model="filters.endDate"
                @change="fetchLogs"
                type="date"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none dark:border-strokedark dark:bg-meta-4"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black dark:text-white">Page Size</label>
              <select
                v-model.number="filters.pageSize"
                @change="fetchLogs"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none dark:border-strokedark dark:bg-meta-4"
              >
                <option :value="25">25</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
            </div>
          </div>

          <!-- Statistics -->
          <div v-if="filters.proxyId && stats" class="grid grid-cols-2 gap-4 rounded-lg border border-stroke p-4 sm:grid-cols-5 dark:border-strokedark">
            <div class="text-center">
              <p class="text-sm text-bodydark">Total Failures</p>
              <p class="text-2xl font-bold text-black dark:text-white">{{ stats.total_failures }}</p>
            </div>
            <div class="text-center">
              <p class="text-sm text-bodydark">Ping Failures</p>
              <p class="text-2xl font-bold text-danger">{{ stats.ping_failures }}</p>
            </div>
            <div class="text-center">
              <p class="text-sm text-bodydark">Speed Failures</p>
              <p class="text-2xl font-bold text-warning">{{ stats.speed_failures }}</p>
            </div>
            <div class="text-center">
              <p class="text-sm text-bodydark">IP Check Failures</p>
              <p class="text-2xl font-bold text-secondary">{{ stats.ip_check_failures }}</p>
            </div>
            <div class="text-center">
              <p class="text-sm text-bodydark">Failure Rate</p>
              <p class="text-2xl font-bold text-black dark:text-white">{{ stats.failure_rate.toFixed(1) }}/day</p>
            </div>
          </div>

          <!-- Logs table -->
          <div class="overflow-x-auto">
            <table class="w-full table-auto">
              <thead>
                <tr class="bg-gray-2 text-left dark:bg-meta-4">
                  <th class="px-4 py-3 font-medium text-black dark:text-white">Timestamp</th>
                  <th class="px-4 py-3 font-medium text-black dark:text-white">Proxy</th>
                  <th class="px-4 py-3 font-medium text-black dark:text-white">Error Type</th>
                  <th class="px-4 py-3 font-medium text-black dark:text-white">Error Message</th>
                  <th class="px-4 py-3 font-medium text-black dark:text-white">Latency</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="log in logs"
                  :key="log.id"
                  class="border-b border-stroke dark:border-strokedark"
                >
                  <td class="px-4 py-3 text-black dark:text-white">
                    {{ formatDate(log.timestamp) }}
                  </td>
                  <td class="px-4 py-3">
                    <span
                      @click="filterByProxy(log.proxy_id)"
                      class="cursor-pointer text-primary hover:underline"
                    >
                      {{ getProxyName(log.proxy_id) }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <span
                      :class="{
                        'bg-danger text-white': log.error_type === 'ping_failed',
                        'bg-warning text-white': log.error_type === 'speed_check_failed',
                        'bg-secondary text-white': log.error_type === 'ip_check_failed'
                      }"
                      class="inline-flex rounded-full px-3 py-1 text-xs font-medium"
                    >
                      {{ formatErrorType(log.error_type) }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-black dark:text-white">
                    <span class="max-w-md truncate" :title="log.error_msg">
                      {{ log.error_msg }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-black dark:text-white">{{ log.latency }}ms</td>
                </tr>
                <tr v-if="logs.length === 0">
                  <td colspan="5" class="px-4 py-8 text-center text-bodydark">
                    No failure logs found. {{ filters.proxyId || filters.errorType || filters.startDate ? 'Try adjusting your filters.' : 'Your proxies are working perfectly!' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div v-if="totalPages > 1" class="flex items-center justify-between">
            <p class="text-sm text-bodydark">
              Showing {{ (currentPage - 1) * filters.pageSize + 1 }} to {{ Math.min(currentPage * filters.pageSize, total) }} of {{ total }} entries
            </p>
            <div class="flex gap-2">
              <button
                @click="changePage(currentPage - 1)"
                :disabled="currentPage === 1"
                class="rounded-md border border-stroke px-3 py-1 hover:bg-gray disabled:opacity-50 dark:border-strokedark"
              >
                Previous
              </button>
              <button
                v-for="page in visiblePages"
                :key="page"
                @click="changePage(page)"
                :class="{
                  'bg-primary text-white': page === currentPage,
                  'border-stroke hover:bg-gray dark:border-strokedark': page !== currentPage
                }"
                class="rounded-md border px-3 py-1"
              >
                {{ page }}
              </button>
              <button
                @click="changePage(currentPage + 1)"
                :disabled="currentPage === totalPages"
                class="rounded-md border border-stroke px-3 py-1 hover:bg-gray disabled:opacity-50 dark:border-strokedark"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      </ComponentCard>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue'
import AdminLayout from '@/components/layout/AdminLayout.vue'
import ComponentCard from '@/components/common/ComponentCard.vue'
import axios from 'axios'

const currentPageTitle = ref('Failure Logs')
const logs = ref([])
const proxies = ref([])
const stats = ref(null)
const total = ref(0)
const currentPage = ref(1)

const filters = ref({
  proxyId: '',
  errorType: '',
  startDate: '',
  endDate: '',
  pageSize: 50
})

const totalPages = computed(() => Math.ceil(total.value / filters.value.pageSize))

const visiblePages = computed(() => {
  const pages = []
  const maxVisible = 5
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2))
  let end = Math.min(totalPages.value, start + maxVisible - 1)

  if (end - start < maxVisible - 1) {
    start = Math.max(1, end - maxVisible + 1)
  }

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

const formatDate = (timestamp: string) => {
  const date = new Date(timestamp)
  return date.toLocaleString()
}

const formatErrorType = (type: string) => {
  return type
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

const getProxyName = (proxyId: string) => {
  const proxy = proxies.value.find((p: any) => p.id === proxyId)
  return proxy ? (proxy.name || proxy.ip) : proxyId
}

const filterByProxy = (proxyId: string) => {
  filters.value.proxyId = proxyId
  fetchLogs()
  fetchStats()
}

const fetchProxies = async () => {
  try {
    const response = await axios.get('/api/proxy')
    proxies.value = response.data.data || []
  } catch (error) {
    console.error('Failed to fetch proxies:', error)
  }
}

const fetchLogs = async () => {
  try {
    const params: any = {
      page: currentPage.value,
      page_size: filters.value.pageSize
    }

    if (filters.value.proxyId) params.proxy_id = filters.value.proxyId
    if (filters.value.errorType) params.error_type = filters.value.errorType
    if (filters.value.startDate) params.start_date = filters.value.startDate
    if (filters.value.endDate) params.end_date = filters.value.endDate

    const response = await axios.get('/api/failureLogs', { params })
    logs.value = response.data.data || []
    total.value = response.data.total || 0
  } catch (error) {
    console.error('Failed to fetch failure logs:', error)
  }
}

const fetchStats = async () => {
  if (!filters.value.proxyId) {
    stats.value = null
    return
  }

  try {
    const response = await axios.get(`/api/failureStats/${filters.value.proxyId}`)
    stats.value = response.data.data
  } catch (error) {
    console.error('Failed to fetch stats:', error)
    stats.value = null
  }
}

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchLogs()
}

onMounted(async () => {
  // Check if proxy_id is in URL query params
  const urlParams = new URLSearchParams(window.location.search)
  const proxyId = urlParams.get('proxy_id')
  if (proxyId) {
    filters.value.proxyId = proxyId
  }

  await fetchProxies()
  await fetchLogs()
  if (filters.value.proxyId) {
    await fetchStats()
  }
})
</script>
