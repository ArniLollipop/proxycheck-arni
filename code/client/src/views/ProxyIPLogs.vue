<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />
    <div class="space-y-5 sm:space-y-6">
      <ComponentCard title="IP Change History">
        <div class="space-y-4">
          <!-- Filters -->
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Proxy</label
              >
              <select
                v-model="filters.proxyId"
                @change="fetchLogs"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none">
                <option value="">All Proxies</option>
                <option
                  v-for="proxy in proxies"
                  :key="proxy.id"
                  :value="proxy.id">
                  {{ proxy.name || proxy.ip }}
                </option>
              </select>
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Start Date</label
              >
              <input
                v-model="filters.startDate"
                @change="fetchLogs"
                type="date"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >End Date</label
              >
              <input
                v-model="filters.endDate"
                @change="fetchLogs"
                type="date"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Page Size</label
              >
              <select
                v-model.number="filters.pageSize"
                @change="fetchLogs"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none">
                <option :value="25">25</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
            </div>
          </div>

          <!-- Logs table -->
          <div class="overflow-x-auto">
            <table class="w-full table-auto">
              <thead>
                <tr class="bg-gray-2 text-left">
                  <th class="px-4 py-3 font-medium text-black">Timestamp</th>
                  <th class="px-4 py-3 font-medium text-black">Proxy</th>
                  <th class="px-4 py-3 font-medium text-black">IP Change</th>
                  <th class="px-4 py-3 font-medium text-black">
                    Country Change
                  </th>
                  <th class="px-4 py-3 font-medium text-black">ISP Change</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="log in logs"
                  :key="log.id"
                  class="border-b border-stroke">
                  <td class="px-4 py-3 text-black">
                    {{ formatDate(log.timestamp) }}
                  </td>
                  <td class="px-4 py-3">
                    <span
                      @click="filterByProxy(log.proxy_id)"
                      class="cursor-pointer text-primary hover:underline">
                      {{ getProxyName(log.proxy_id) }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex flex-col">
                      <span
                        v-if="log.old_ip"
                        class="text-xs text-bodydark line-through">
                        {{ log.old_ip }}
                      </span>
                      <span
                        class="font-mono text-sm"
                        :class="
                          log.stack
                            ? 'text-red-500 font-semibold'
                            : 'text-black'
                        ">
                        {{ log.ip }}
                      </span>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex flex-col">
                      <span
                        v-if="log.old_country && log.old_country != log.country"
                        class="text-xs text-bodydark line-through">
                        {{ log.old_country }}
                      </span>
                      <span class="text-sm text-black">
                        {{ log.country }}
                      </span>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex flex-col">
                      <span
                        v-if="log.old_isp && log.old_isp != log.isp"
                        class="text-xs text-bodydark line-through">
                        {{ log.old_isp }}
                      </span>
                      <span class="text-sm text-black">
                        {{ log.isp }}
                      </span>
                    </div>
                  </td>
                </tr>
                <tr v-if="logs.length === 0">
                  <td colspan="6" class="px-4 py-8 text-center text-bodydark">
                    No IP change logs found.
                    {{
                      filters.proxyId || filters.startDate
                        ? "Try adjusting your filters."
                        : "No IP changes recorded yet."
                    }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div v-if="totalPages > 1" class="flex items-center justify-between">
            <p class="text-sm text-bodydark">
              Showing {{ (currentPage - 1) * filters.pageSize + 1 }} to
              {{ Math.min(currentPage * filters.pageSize, total) }} of
              {{ total }} entries
            </p>
            <div class="flex gap-2">
              <button
                @click="changePage(currentPage - 1)"
                :disabled="currentPage === 1"
                class="rounded-md border border-stroke px-3 py-1 hover:bg-gray disabled:opacity-50">
                Previous
              </button>
              <button
                v-for="page in visiblePages"
                :key="page"
                @click="changePage(page)"
                :class="{
                  'bg-primary text-white': page === currentPage,
                  'border-stroke hover:bg-gray': page !== currentPage,
                }"
                class="rounded-md border px-3 py-1">
                {{ page }}
              </button>
              <button
                @click="changePage(currentPage + 1)"
                :disabled="currentPage === totalPages"
                class="rounded-md border border-stroke px-3 py-1 hover:bg-gray disabled:opacity-50">
                Next
              </button>
            </div>
          </div>
        </div>
      </ComponentCard>
    </div>
  </AdminLayout>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
import AdminLayout from "@/components/layout/AdminLayout.vue";
import ComponentCard from "@/components/common/ComponentCard.vue";
import axios from "axios";

const currentPageTitle = ref("IP Change Logs");
const logs = ref([]);
const proxies = ref([]);
const total = ref(0);
const currentPage = ref(1);

const filters = ref({
  proxyId: "",
  startDate: "",
  endDate: "",
  pageSize: 50,
});

const totalPages = computed(() =>
  Math.ceil(total.value / filters.value.pageSize)
);

const visiblePages = computed(() => {
  const pages = [];
  const maxVisible = 5;
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages.value, start + maxVisible - 1);

  if (end - start < maxVisible - 1) {
    start = Math.max(1, end - maxVisible + 1);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }
  return pages;
});

const formatDate = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleString();
};

const getProxyName = (proxyId) => {
  const proxy = proxies.value.find((p) => p.id === proxyId);
  return proxy ? proxy.name || proxy.ip : proxyId;
};

const filterByProxy = (proxyId) => {
  filters.value.proxyId = proxyId;
  fetchLogs();
};

const fetchProxies = async () => {
  try {
    const response = await axios.get("/api/proxy");
    proxies.value = response.data.data || [];
  } catch (error) {
    console.error("Failed to fetch proxies:", error);
  }
};

const fetchLogs = async () => {
  try {
    const params = {
      page: currentPage.value,
      page_size: filters.value.pageSize,
    };

    if (filters.value.proxyId) params.proxy_id = filters.value.proxyId;
    if (filters.value.startDate) params.start_date = filters.value.startDate;
    if (filters.value.endDate) params.end_date = filters.value.endDate;

    const response = await axios.get("/api/ipLogs", { params });
    logs.value = response.data.data || [];
    total.value = response.data.total || 0;
  } catch (error) {
    console.error("Failed to fetch IP logs:", error);
  }
};

const changePage = (page) => {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  fetchLogs();
};

onMounted(async () => {
  // Check if proxy_id is in URL query params
  const urlParams = new URLSearchParams(window.location.search);
  const proxyId = urlParams.get("proxy_id");
  if (proxyId) {
    filters.value.proxyId = proxyId;
  }

  await fetchProxies();
  await fetchLogs();
});
</script>
