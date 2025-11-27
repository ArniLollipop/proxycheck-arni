<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />
    <div class="space-y-5 sm:space-y-6">
      <ComponentCard title="Proxy Management">
        <div class="space-y-4">
          <!-- Action buttons -->
          <div class="flex flex-wrap gap-3 items-center justify-between">
            <div class="flex gap-2">
              <button
                @click="showAddModal = true"
                class="inline-flex items-center justify-center rounded-md border border-stroke px-4 py-2 text-center font-medium hover:bg-gray">
                <Plus class="w-4 h-4 mr-2" />
                Add Proxy
              </button>
              <button
                @click="showImportModal = true"
                class="inline-flex items-center justify-center rounded-md border border-stroke px-4 py-2 text-center font-medium hover:bg-gray">
                <Upload class="w-4 h-4 mr-2" />
                Import
              </button>
              <button
                @click="exportSelected"
                :disabled="selectedProxies.length === 0"
                class="inline-flex items-center justify-center rounded-md border border-stroke px-4 py-2 text-center font-medium hover:bg-gray">
                <Download class="w-4 h-4 mr-2" />
                Export Selected
              </button>
              <button
                @click="verifySelected"
                :disabled="selectedProxies.length === 0 || isVerifying"
                class="inline-flex items-center justify-center rounded-md border border-stroke px-4 py-2 text-center font-medium hover:bg-gray">
                <RefreshCw
                  class="w-4 h-4 mr-2"
                  :class="{ 'animate-spin': isVerifying }" />
                Verify Selected
              </button>
              <button
                @click="showColumnSettings = true"
                class="inline-flex items-center justify-center rounded-md border border-stroke px-4 py-2 text-center font-medium hover:bg-gray">
                <Settings class="w-4 h-4 mr-2" />
                Columns
              </button>
            </div>
            <div class="flex items-center gap-2">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search proxies..."
                class="rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
          </div>

          <!-- Proxies table -->
          <div class="overflow-x-auto">
            <table class="w-full table-auto text-sm">
              <thead>
                <tr class="bg-gray-2 text-left">
                  <th class="px-2 py-2">
                    <input
                      type="checkbox"
                      @change="toggleSelectAll"
                      :checked="
                        selectedProxies.length === proxies.length &&
                        proxies.length > 0
                      "
                      class="cursor-pointer" />
                  </th>
                  <th
                    v-if="visibleColumns.status"
                    class="px-2 py-2 font-medium text-black">
                    Status
                  </th>
                  <th
                    v-if="visibleColumns.client"
                    class="px-2 py-2 font-medium text-black">
                    Client
                  </th>
                  <th
                    v-if="visibleColumns.modemId"
                    class="px-2 py-2 font-medium text-black">
                    Modem ID
                  </th>
                  <th
                    v-if="visibleColumns.pcId"
                    class="px-2 py-2 font-medium text-black">
                    PC ID
                  </th>
                  <th
                    v-if="visibleColumns.serverIp"
                    class="px-2 py-2 font-medium text-black">
                    Server IP
                  </th>
                  <th
                    v-if="visibleColumns.port"
                    class="px-2 py-2 font-medium text-black">
                    Port
                  </th>
                  <th
                    v-if="visibleColumns.realIp"
                    class="px-2 py-2 font-medium text-black">
                    Real IP
                  </th>
                  <th
                    v-if="visibleColumns.country"
                    class="px-2 py-2 font-medium text-black">
                    Country
                  </th>
                  <th
                    v-if="visibleColumns.operator"
                    class="px-2 py-2 font-medium text-black">
                    Operator
                  </th>
                  <th
                    v-if="visibleColumns.phone"
                    class="px-2 py-2 font-medium text-black">
                    Phone
                  </th>
                  <th
                    v-if="visibleColumns.username"
                    class="px-2 py-2 font-medium text-black">
                    Username
                  </th>
                  <th
                    v-if="visibleColumns.password"
                    class="px-2 py-2 font-medium text-black">
                    Password
                  </th>
                  <th
                    v-if="visibleColumns.uptime"
                    class="px-2 py-2 font-medium text-black">
                    Uptime
                  </th>
                  <th
                    v-if="visibleColumns.latency"
                    class="px-2 py-2 font-medium text-black">
                    Latency
                  </th>
                  <th
                    v-if="visibleColumns.download"
                    class="px-2 py-2 font-medium text-black">
                    Speed ↓
                  </th>
                  <th
                    v-if="visibleColumns.upload"
                    class="px-2 py-2 font-medium text-black">
                    Speed ↑
                  </th>
                  <th
                    v-if="visibleColumns.traffic24h"
                    class="px-2 py-2 font-medium text-black">
                    Traffic 24h
                  </th>
                  <th
                    v-if="visibleColumns.trafficLeft"
                    class="px-2 py-2 font-medium text-black">
                    Traffic Left
                  </th>
                  <th
                    v-if="visibleColumns.logs"
                    class="px-2 py-2 font-medium text-black">
                    Logs
                  </th>
                  <th class="px-2 py-2 font-medium text-black">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="proxy in filteredProxies"
                  :key="proxy.id"
                  :class="{
                    'bg-warning bg-opacity-10': verifyingProxies.has(proxy.id),
                    'animate-pulse': verifyingProxies.has(proxy.id),
                    'bg-danger bg-opacity-5':
                      proxy.stack && !verifyingProxies.has(proxy.id),
                  }"
                  class="border-b border-stroke hover:bg-gray-2">
                  <td class="px-2 py-2">
                    <input
                      type="checkbox"
                      :checked="selectedProxies.includes(proxy.id)"
                      @change="toggleSelect(proxy.id)"
                      :disabled="verifyingProxies.has(proxy.id)"
                      class="cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed" />
                  </td>

                  <!-- Status -->
                  <td v-if="visibleColumns.status" class="px-2 py-2">
                    <div class="flex items-center gap-1">
                      <RefreshCw
                        v-if="verifyingProxies.has(proxy.id)"
                        class="w-4 h-4 animate-spin text-primary" />
                      <span
                        v-else
                        :class="{
                          'text-success': proxy.lastStatus === 1,
                          'text-warning': proxy.lastStatus === 0,
                          'text-danger': proxy.lastStatus === 2,
                        }"
                        class="text-xl">
                        {{
                          proxy.lastStatus === 1
                            ? "🟢"
                            : proxy.lastStatus === 2
                              ? "🔴"
                              : "🟡"
                        }}
                      </span>
                      <span
                        v-if="
                          proxy.failures > 0 && !verifyingProxies.has(proxy.id)
                        "
                        class="text-xs text-bodydark"
                        >({{ proxy.failures }})</span
                      >
                      <span
                        v-if="verifyingProxies.has(proxy.id)"
                        class="text-xs text-primary font-medium"
                        >Verifying...</span
                      >
                    </div>
                  </td>

                  <!-- Client -->
                  <td v-if="visibleColumns.client" class="px-2 py-2 text-black">
                    {{ proxy.name || "-" }}
                  </td>

                  <!-- Modem ID (using contacts field) -->
                  <td
                    v-if="visibleColumns.modemId"
                    class="px-2 py-2 text-black text-xs">
                    {{ proxy.contacts || "-" }}
                  </td>

                  <!-- PC ID (can add new field to backend) -->
                  <td
                    v-if="visibleColumns.pcId"
                    class="px-2 py-2 text-black text-xs">
                    -
                  </td>

                  <!-- Server IP -->
                  <td
                    v-if="visibleColumns.serverIp"
                    class="px-2 py-2 text-black font-mono text-xs">
                    {{ proxy.ip }}
                  </td>

                  <!-- Port -->
                  <td
                    v-if="visibleColumns.port"
                    class="px-2 py-2 text-black font-mono text-xs">
                    {{ proxy.port }}
                  </td>

                  <!-- Real IP with time since change -->
                  <td
                    v-if="visibleColumns.realIp"
                    class="px-2 py-2 font-mono text-xs">
                    <div class="flex flex-col">
                      <span
                        :class="
                          proxy.stack
                            ? 'text-danger font-semibold'
                            : 'text-black'
                        ">
                        {{ proxy.realIP || "-" }}
                      </span>
                      <span
                        v-if="proxy.lastCheck"
                        class="text-xs text-bodydark">
                        ({{ getMinutesSince(proxy.lastCheck) }}m ago)
                      </span>
                      <span
                        v-if="proxy.stack"
                        class="text-xs text-danger font-medium">
                        ⚠️ IP Stuck
                      </span>
                    </div>
                  </td>

                  <!-- Country -->
                  <td
                    v-if="visibleColumns.country"
                    class="px-2 py-2 text-black text-xs">
                    {{ proxy.realCountry || "-" }}
                  </td>

                  <!-- Operator -->
                  <td
                    v-if="visibleColumns.operator"
                    class="px-2 py-2 text-black text-xs">
                    {{ proxy.operator || "-" }}
                  </td>

                  <!-- Phone -->
                  <td
                    v-if="visibleColumns.phone"
                    class="px-2 py-2 text-black text-xs">
                    {{ proxy.phone || "-" }}
                  </td>

                  <!-- Username with copy -->
                  <td v-if="visibleColumns.username" class="px-2 py-2">
                    <div class="flex items-center gap-1">
                      <span class="text-black font-mono text-xs">{{
                        proxy.username
                      }}</span>
                      <button
                        @click="copyToClipboard(proxy.username)"
                        class="text-primary hover:text-opacity-80"
                        title="Copy username">
                        <Key class="w-3 h-3" />
                      </button>
                    </div>
                  </td>

                  <!-- Password with show/hide -->
                  <td v-if="visibleColumns.password" class="px-2 py-2">
                    <div class="flex items-center gap-1">
                      <span class="text-black font-mono text-xs">
                        {{ showPasswords[proxy.id] ? proxy.password : "***" }}
                      </span>
                      <button
                        @click="togglePassword(proxy.id)"
                        class="text-bodydark hover:text-black"
                        :title="showPasswords[proxy.id] ? 'Hide' : 'Show'">
                        <Eye v-if="showPasswords[proxy.id]" class="w-3 h-3" />
                        <EyeOff v-else class="w-3 h-3" />
                      </button>
                    </div>
                  </td>

                  <!-- Uptime as percentage with color -->
                  <td v-if="visibleColumns.uptime" class="px-2 py-2">
                    <span
                      :class="{
                        'text-success': getUptimePercent(proxy) >= 95,
                        'text-warning':
                          getUptimePercent(proxy) >= 70 &&
                          getUptimePercent(proxy) < 95,
                        'text-danger': getUptimePercent(proxy) < 70,
                      }"
                      class="font-medium text-xs">
                      {{ getUptimePercent(proxy).toFixed(1) }}%
                    </span>
                  </td>

                  <!-- Latency with trend -->
                  <td v-if="visibleColumns.latency" class="px-2 py-2">
                    <div class="flex items-center gap-1">
                      <span class="text-black text-xs"
                        >{{ proxy.lastLatency || 0 }}ms</span
                      >
                      <span class="text-bodydark text-xs">→</span>
                    </div>
                  </td>

                  <!-- Download Speed -->
                  <td
                    v-if="visibleColumns.download"
                    class="px-2 py-2 text-black text-xs">
                    {{ proxy.speed || 0 }} Mbps
                  </td>

                  <!-- Upload Speed -->
                  <td
                    v-if="visibleColumns.upload"
                    class="px-2 py-2 text-black text-xs">
                    {{ proxy.upload || 0 }} Mbps
                  </td>

                  <!-- Traffic 24h (placeholder) -->
                  <td
                    v-if="visibleColumns.traffic24h"
                    class="px-2 py-2 text-black text-xs">
                    - GB
                  </td>

                  <!-- Traffic Left (placeholder) -->
                  <td
                    v-if="visibleColumns.trafficLeft"
                    class="px-2 py-2 text-black text-xs">
                    - GB
                  </td>

                  <!-- Logs Analysis (placeholder top-5) -->
                  <td v-if="visibleColumns.logs" class="px-2 py-2">
                    <div class="flex flex-col gap-1">
                      <button
                        @click="viewLogs(proxy.id)"
                        class="text-primary hover:underline text-xs">
                        IP Logs
                      </button>
                      <button
                        @click="viewSpeedLogs(proxy.id)"
                        class="text-success hover:underline text-xs">
                        Speed Logs
                      </button>
                    </div>
                  </td>

                  <!-- Actions -->
                  <td class="px-2 py-2">
                    <div class="flex gap-1">
                      <button
                        @click="verifySingle(proxy.id)"
                        class="text-primary hover:text-opacity-80"
                        title="Refresh/Verify">
                        <RefreshCw class="w-4 h-4" />
                      </button>
                      <button
                        @click="editProxy(proxy)"
                        class="text-secondary hover:text-opacity-80"
                        title="Settings">
                        <Settings class="w-4 h-4" />
                      </button>
                      <button
                        @click="viewDetails(proxy)"
                        class="text-success hover:text-opacity-80"
                        title="View Details">
                        <Search class="w-4 h-4" />
                      </button>
                      <button
                        @click="deleteProxy(proxy.id)"
                        class="text-danger hover:text-opacity-80"
                        title="Delete">
                        <X class="w-4 h-4" />
                      </button>
                      <button
                        @click="copyProxyString(proxy)"
                        class="text-warning hover:text-opacity-80"
                        title="Copy proxy string">
                        <Clipboard class="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="proxies.length === 0">
                  <td
                    :colspan="getVisibleColumnCount()"
                    class="px-4 py-8 text-center text-bodydark">
                    No proxies found. Add a proxy to get started.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </ComponentCard>
    </div>

    <!-- Column Settings Modal -->
    <div
      v-if="showColumnSettings"
      class="fixed inset-0 z-999999 flex items-center justify-center bg-black bg-opacity-50"
      @click.self="showColumnSettings = false">
      <div
        class="w-full max-w-2xl rounded-lg bg-white p-6 max-h-[80vh] overflow-y-auto">
        <h3 class="mb-4 text-xl font-semibold text-black">Column Settings</h3>
        <div class="grid grid-cols-2 gap-3">
          <label
            v-for="(column, key) in columnLabels"
            :key="key"
            class="flex items-center cursor-pointer">
            <input
              v-model="visibleColumns[key]"
              type="checkbox"
              class="mr-2 cursor-pointer" />
            <span class="text-sm text-black">{{ column }}</span>
          </label>
        </div>
        <div class="mt-6 flex justify-end gap-3">
          <button
            @click="resetColumns"
            class="rounded-md border border-stroke px-4 py-2 hover:bg-gray">
            Reset to Default
          </button>
          <button
            @click="saveColumnSettings"
            class="rounded-md border border-stroke px-4 py-2 hover:bg-gray">
            Save
          </button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div
      v-if="showAddModal || showEditModal"
      class="fixed inset-0 z-999999 flex items-center justify-center bg-black bg-opacity-50"
      @click.self="closeModals">
      <div
        class="w-full max-w-2xl rounded-lg bg-white p-6 max-h-[80vh] overflow-y-auto">
        <h3 class="mb-4 text-xl font-semibold text-black">
          {{ showEditModal ? "Edit Proxy" : "Add Proxy" }}
        </h3>
        <form
          @submit.prevent="showEditModal ? updateProxy() : createProxy()"
          class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Client/Name</label
              >
              <input
                v-model="proxyForm.name"
                type="text"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >IP Address *</label
              >
              <input
                v-model="proxyForm.ip"
                type="text"
                required
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Port *</label
              >
              <input
                v-model="proxyForm.port"
                type="text"
                required
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Username *</label
              >
              <input
                v-model="proxyForm.username"
                type="text"
                required
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Password *</label
              >
              <input
                v-model="proxyForm.password"
                type="password"
                required
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-black"
                >Phone</label
              >
              <input
                v-model="proxyForm.phone"
                type="text"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div class="col-span-2">
              <label class="mb-2 block text-sm font-medium text-black"
                >Modem ID / Contacts</label
              >
              <input
                v-model="proxyForm.contacts"
                type="text"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
          </div>
          <div class="flex justify-end gap-3">
            <button
              type="button"
              @click="closeModals"
              class="rounded-md border border-stroke px-4 py-2 hover:bg-gray">
              Cancel
            </button>
            <button
              type="submit"
              class="rounded-md border border-stroke px-4 py-2 hover:bg-gray">
              {{ showEditModal ? "Update" : "Create" }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Import Modal -->
    <div
      v-if="showImportModal"
      class="fixed inset-0 z-999999 flex items-center justify-center bg-black bg-opacity-50"
      @click.self="showImportModal = false">
      <div class="w-full max-w-lg rounded-lg bg-white p-6">
        <h3 class="mb-4 text-xl font-semibold text-black">Import Proxies</h3>
        <form @submit.prevent="importProxies" class="space-y-4">
          <div>
            <label class="mb-2 block text-sm font-medium text-black">
              Select file (format: ip:port:username:password|name|contacts)
            </label>
            <input
              ref="fileInput"
              type="file"
              accept=".txt"
              @change="handleFileSelect"
              class="w-full" />
          </div>
          <div class="flex justify-end gap-3">
            <button
              type="button"
              @click="showImportModal = false"
              class="rounded-md border border-stroke px-4 py-2 hover:bg-gray">
              Cancel
            </button>
            <button
              type="submit"
              :disabled="!selectedFile"
              class="rounded-md border border-stroke px-4 py-2 hover:bg-gray">
              Import
            </button>
          </div>
        </form>
      </div>
    </div>
  </AdminLayout>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import {
  Plus,
  Upload,
  Download,
  RefreshCw,
  Settings,
  Search,
  X,
  Clipboard,
  Key,
  Eye,
  EyeOff,
} from "lucide-vue-next";
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
import AdminLayout from "@/components/layout/AdminLayout.vue";
import ComponentCard from "@/components/common/ComponentCard.vue";
import axios from "axios";

const currentPageTitle = ref("Proxies");
const proxies = ref([]);
const selectedProxies = ref([]);
const searchQuery = ref("");
const showAddModal = ref(false);
const showEditModal = ref(false);
const showImportModal = ref(false);
const showColumnSettings = ref(false);
const isVerifying = ref(false);
const verifyingProxies = ref(new Set()); // Track which proxies are being verified
const selectedFile = ref(null);
const fileInput = ref(null);
const showPasswords = ref({});

const columnLabels = {
  status: "Status",
  client: "Client/Tenant",
  modemId: "Modem ID",
  pcId: "PC ID/Address",
  serverIp: "Server IP",
  port: "Port",
  realIp: "Real IP",
  country: "Country",
  operator: "Operator",
  phone: "Phone",
  username: "Username",
  password: "Password",
  uptime: "Uptime %",
  latency: "Latency",
  download: "Speed ↓",
  upload: "Speed ↑",
  traffic24h: "Traffic 24h",
  trafficLeft: "Traffic Left",
  logs: "Logs Analysis",
};

const visibleColumns = ref({
  status: true,
  client: true,
  modemId: true,
  pcId: false,
  serverIp: true,
  port: true,
  realIp: true,
  country: true,
  operator: true,
  phone: true,
  username: true,
  password: true,
  uptime: true,
  latency: true,
  download: true,
  upload: true,
  traffic24h: false,
  trafficLeft: false,
  logs: true,
});

const proxyForm = ref({
  id: "",
  name: "",
  ip: "",
  port: "",
  username: "",
  password: "",
  phone: "",
  contacts: "",
});

const filteredProxies = computed(() => {
  if (!searchQuery.value) return proxies.value;
  const query = searchQuery.value.toLowerCase();
  return proxies.value.filter(
    (proxy) =>
      proxy.name?.toLowerCase().includes(query) ||
      proxy.ip?.toLowerCase().includes(query) ||
      proxy.username?.toLowerCase().includes(query) ||
      proxy.realIP?.toLowerCase().includes(query) ||
      proxy.realCountry?.toLowerCase().includes(query) ||
      proxy.phone?.toLowerCase().includes(query) ||
      proxy.contacts?.toLowerCase().includes(query)
  );
});

const getVisibleColumnCount = () => {
  return Object.values(visibleColumns.value).filter((v) => v).length + 2; // +2 for checkbox and actions
};

const getUptimePercent = (proxy) => {
  const totalTime = proxy.uptime + proxy.failures * 15; // Assume 15 min per failure
  if (totalTime === 0) return 100;
  return (proxy.uptime / totalTime) * 100;
};

const getMinutesSince = (lastCheck) => {
  if (!lastCheck) return 0;
  const diff = Date.now() - new Date(lastCheck).getTime();
  return Math.floor(diff / 60000);
};

const togglePassword = (proxyId) => {
  showPasswords.value[proxyId] = !showPasswords.value[proxyId];
};

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text);
    alert("Copied to clipboard!");
  } catch (err) {
    console.error("Failed to copy:", err);
  }
};

const copyProxyString = (proxy) => {
  const proxyString = `${proxy.ip}:${proxy.port}:${proxy.username}:${proxy.password}`;
  copyToClipboard(proxyString);
};

const saveColumnSettings = () => {
  localStorage.setItem("proxyColumns", JSON.stringify(visibleColumns.value));
  showColumnSettings.value = false;
};

const resetColumns = () => {
  visibleColumns.value = {
    status: true,
    client: true,
    modemId: true,
    pcId: false,
    serverIp: true,
    port: true,
    realIp: true,
    country: true,
    operator: true,
    phone: true,
    username: true,
    password: true,
    uptime: true,
    latency: true,
    download: true,
    upload: true,
    traffic24h: false,
    trafficLeft: false,
    logs: true,
  };
};

const loadColumnSettings = () => {
  const saved = localStorage.getItem("proxyColumns");
  if (saved) {
    visibleColumns.value = JSON.parse(saved);
  }
};

const fetchProxies = async () => {
  try {
    const response = await axios.get("/api/proxy");
    proxies.value = response.data.data || [];
  } catch (error) {
    console.error("Failed to fetch proxies:", error);
  }
};

const createProxy = async () => {
  try {
    await axios.post("/api/proxy", proxyForm.value);
    await fetchProxies();
    closeModals();
    resetForm();
  } catch (error) {
    console.error("Failed to create proxy:", error);
    alert("Failed to create proxy");
  }
};

const updateProxy = async () => {
  try {
    await axios.put(`/api/proxy/${proxyForm.value.id}`, proxyForm.value);
    await fetchProxies();
    closeModals();
    resetForm();
  } catch (error) {
    console.error("Failed to update proxy:", error);
    alert("Failed to update proxy");
  }
};

const deleteProxy = async (id) => {
  if (!confirm("Are you sure you want to delete this proxy?")) return;
  try {
    await axios.delete(`/api/proxy/${id}`);
    await fetchProxies();
  } catch (error) {
    console.error("Failed to delete proxy:", error);
    alert("Failed to delete proxy");
  }
};

const editProxy = (proxy) => {
  proxyForm.value = { ...proxy };
  showEditModal.value = true;
};

const viewDetails = (proxy) => {
  // Could open a detail modal or navigate to detail page
  alert(`Viewing details for ${proxy.name || proxy.ip}`);
};

const verifySingle = async (id) => {
  selectedProxies.value = [id];
  await verifySelected();
};

const toggleSelect = (id) => {
  const index = selectedProxies.value.indexOf(id);
  if (index > -1) {
    selectedProxies.value.splice(index, 1);
  } else {
    selectedProxies.value.push(id);
  }
};

const toggleSelectAll = () => {
  if (selectedProxies.value.length === proxies.value.length) {
    selectedProxies.value = [];
  } else {
    selectedProxies.value = proxies.value.map((p) => p.id);
  }
};

const exportSelected = () => {
  const ids = selectedProxies.value.join(",");
  window.location.href = `/api/export/selected?ids=${ids}`;
};

const verifySelected = async () => {
  if (selectedProxies.value.length === 0) return;

  isVerifying.value = true;
  verifyingProxies.value = new Set(selectedProxies.value);
  const ids = selectedProxies.value.join(",");

  const eventSource = new EventSource(`/api/verify-batch?ids=${ids}`);

  eventSource.addEventListener("start", (event) => {
    const data = JSON.parse(event.data);
    console.log(
      `Starting verification ${data.current}/${data.total} for proxy ${data.id}`
    );
  });

  eventSource.addEventListener("progress", (event) => {
    const data = JSON.parse(event.data);
    const index = proxies.value.findIndex((p) => p.id === data.id);
    if (index > -1) {
      proxies.value[index] = data;
    }
    // Remove from verifying set
    verifyingProxies.value.delete(data.id);
  });

  eventSource.addEventListener("complete", () => {
    isVerifying.value = false;
    verifyingProxies.value.clear();
    eventSource.close();
    selectedProxies.value = [];
  });

  eventSource.onerror = (error) => {
    console.error("SSE error:", error);
    isVerifying.value = false;
    verifyingProxies.value.clear();
    eventSource.close();
  };
};

const handleFileSelect = (event) => {
  const target = event.target;
  selectedFile.value = target.files?.[0] || null;
};

const importProxies = async () => {
  if (!selectedFile.value) return;

  const formData = new FormData();
  formData.append("file", selectedFile.value);

  try {
    const response = await axios.post("/api/import", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    alert(response.data.message);
    await fetchProxies();
    showImportModal.value = false;
    selectedFile.value = null;
    if (fileInput.value) fileInput.value.value = "";
  } catch (error) {
    console.error("Failed to import proxies:", error);
    alert("Failed to import proxies");
  }
};

const viewLogs = (proxyId) => {
  window.location.href = `/ip-logs?proxy_id=${proxyId}`;
};

const viewSpeedLogs = (proxyId) => {
  window.location.href = `/speed-logs?proxy_id=${proxyId}`;
};

const closeModals = () => {
  showAddModal.value = false;
  showEditModal.value = false;
  resetForm();
};

const resetForm = () => {
  proxyForm.value = {
    id: "",
    name: "",
    ip: "",
    port: "",
    username: "",
    password: "",
    phone: "",
    contacts: "",
  };
};

onMounted(() => {
  loadColumnSettings();
  fetchProxies();
});
</script>
