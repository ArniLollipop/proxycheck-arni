<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />
    <div class="space-y-5 sm:space-y-6">
      <form @submit.prevent="saveSettings" class="space-y-6">
        <!-- General Settings -->
        <ComponentCard title="General Settings">
          <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
            <div>
              <label
                class="mb-2 block text-sm font-medium text-black"
                >Check URL</label
              >
              <input
                v-model="settings.url"
                type="text"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label
                class="mb-2 block text-sm font-medium text-black"
                >Timeout (seconds)</label
              >
              <input
                v-model.number="settings.timeout"
                type="number"
                min="1"
                max="300"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label
                class="mb-2 block text-sm font-medium text-black"
                >IP Check Interval (minutes)</label
              >
              <input
                v-model.number="settings.checkIPInterval"
                type="number"
                min="1"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
              <p class="mt-1 text-xs text-bodydark">
                How often to check proxy IPs (recommended: 15-30 min)
              </p>
            </div>
            <div>
              <label
                class="mb-2 block text-sm font-medium text-black"
                >Speed Check Interval (minutes)</label
              >
              <input
                v-model.number="settings.speedCheckInterval"
                type="number"
                min="1"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
              <p class="mt-1 text-xs text-bodydark">
                How often to check proxy speeds (recommended: 60-360 min)
              </p>
            </div>
            <div>
              <label
                class="mb-2 block text-sm font-medium text-black"
                >Username</label
              >
              <input
                v-model="settings.username"
                type="text"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div>
              <label
                class="mb-2 block text-sm font-medium text-black"
                >Password</label
              >
              <input
                v-model="settings.password"
                type="password"
                class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
            </div>
            <div class="flex items-center">
              <input
                v-model="settings.skipSSLVerify"
                type="checkbox"
                id="skipSSL"
                class="mr-2 cursor-pointer" />
              <label
                for="skipSSL"
                class="text-sm font-medium text-black cursor-pointer">
                Skip SSL Verification
              </label>
            </div>
          </div>
        </ComponentCard>

        <!-- Telegram Notification Settings -->
        <ComponentCard title="Telegram Notifications">
          <div class="space-y-6">
            <div class="flex items-center justify-between">
              <div>
                <h4 class="font-medium text-black">
                  Enable Telegram Notifications
                </h4>
                <p class="text-sm text-bodydark">
                  Receive alerts about proxy status changes
                </p>
              </div>
              <label class="relative inline-flex cursor-pointer items-center">
                <input
                  v-model="settings.telegramEnabled"
                  type="checkbox"
                  class="peer sr-only" />
                <div
                  class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-primary peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300"></div>
              </label>
            </div>

            <div v-if="settings.telegramEnabled" class="space-y-4">
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-black"
                    >Bot Token</label
                  >
                  <input
                    v-model="settings.telegramToken"
                    type="text"
                    placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                    class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
                  <p class="mt-1 text-xs text-bodydark">
                    Get from @BotFather on Telegram
                  </p>
                </div>
                <div>
                  <label
                    class="mb-2 block text-sm font-medium text-black"
                    >Chat ID</label
                  >
                  <input
                    v-model="settings.telegramChatID"
                    type="text"
                    placeholder="123456789 or -1001234567890"
                    class="w-full rounded-md border border-stroke px-4 py-2 focus:border-primary focus:outline-none" />
                  <p class="mt-1 text-xs text-bodydark">
                    Your chat ID or channel ID
                  </p>
                </div>
              </div>

              <div
                class="rounded-lg border border-stroke p-4">
                <h5 class="mb-3 font-medium text-black">
                  Notification Types
                </h5>
                <div class="space-y-3">
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="settings.notifyOnDown"
                      type="checkbox"
                      class="mr-3 h-5 w-5 cursor-pointer" />
                    <div>
                      <span class="font-medium text-black"
                        >Proxy Down</span
                      >
                      <p class="text-xs text-bodydark">
                        Alert when proxy goes offline
                      </p>
                    </div>
                  </label>
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="settings.notifyOnRecovery"
                      type="checkbox"
                      class="mr-3 h-5 w-5 cursor-pointer" />
                    <div>
                      <span class="font-medium text-black"
                        >Proxy Recovery</span
                      >
                      <p class="text-xs text-bodydark">
                        Alert when offline proxy comes back online
                      </p>
                    </div>
                  </label>
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="settings.notifyOnIPChange"
                      type="checkbox"
                      class="mr-3 h-5 w-5 cursor-pointer" />
                    <div>
                      <span class="font-medium text-black"
                        >IP Change</span
                      >
                      <p class="text-xs text-bodydark">
                        Alert when proxy IP address changes
                      </p>
                    </div>
                  </label>
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="settings.notifyOnIPStuck"
                      type="checkbox"
                      class="mr-3 h-5 w-5 cursor-pointer" />
                    <div>
                      <span class="font-medium text-black"
                        >IP Stuck</span
                      >
                      <p class="text-xs text-bodydark">
                        Alert when IP hasn't changed for >24 hours
                      </p>
                    </div>
                  </label>
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="settings.notifyOnLowSpeed"
                      type="checkbox"
                      class="mr-3 h-5 w-5 cursor-pointer" />
                    <div class="flex-1">
                      <div class="flex items-center justify-between">
                        <div>
                          <span class="font-medium text-black"
                            >Low Speed</span
                          >
                          <p class="text-xs text-bodydark">
                            Alert when speed drops below threshold
                          </p>
                        </div>
                        <input
                          v-model.number="settings.lowSpeedThreshold"
                          :disabled="!settings.notifyOnLowSpeed"
                          type="number"
                          min="1"
                          max="1000"
                          class="ml-4 w-20 rounded-md border border-stroke px-2 py-1 text-sm focus:border-primary focus:outline-none disabled:opacity-50" />
                        <span class="ml-2 text-sm text-bodydark">Mbps</span>
                      </div>
                    </div>
                  </label>
                  <label class="flex items-center cursor-pointer">
                    <input
                      v-model="settings.notifyDailySummary"
                      type="checkbox"
                      class="mr-3 h-5 w-5 cursor-pointer" />
                    <div class="flex-1">
                      <div class="flex items-center justify-between">
                        <div>
                          <span class="font-medium text-black"
                            >Daily Summary</span
                          >
                          <p class="text-xs text-bodydark">
                            Daily report of proxy statistics
                          </p>
                        </div>
                        <input
                          v-model="settings.dailySummaryTime"
                          :disabled="!settings.notifyDailySummary"
                          type="time"
                          class="ml-4 rounded-md border border-stroke px-2 py-1 text-sm focus:border-primary focus:outline-none disabled:opacity-50" />
                      </div>
                    </div>
                  </label>
                </div>
              </div>

              <div class="flex gap-3">
                <button
                  type="button"
                  @click="testNotification"
                  :disabled="
                    !settings.telegramToken ||
                    !settings.telegramChatID ||
                    isTesting
                  "
                  class="inline-flex items-center justify-center rounded-md bg-secondary px-4 py-2 text-center font-medium text-white hover:bg-opacity-90 disabled:opacity-50">
                  <Send
                    class="mr-2 h-4 w-4"
                    :class="{ 'animate-pulse': isTesting }" />
                  {{ isTesting ? "Sending..." : "Test Notification" }}
                </button>
              </div>
            </div>
          </div>
        </ComponentCard>

        <!-- Save button -->
        <div class="flex justify-end">
          <button
            type="submit"
            :disabled="isSaving"
            class="inline-flex items-center justify-center rounded-md bg-primary px-6 py-3 text-center font-medium text-white hover:bg-opacity-90 disabled:opacity-50">
            <Save class="mr-2 h-4 w-4" />
            {{ isSaving ? "Saving..." : "Save Settings" }}
          </button>
        </div>
      </form>
    </div>
  </AdminLayout>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { Save, Send } from "lucide-vue-next";
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
import AdminLayout from "@/components/layout/AdminLayout.vue";
import ComponentCard from "@/components/common/ComponentCard.vue";
import axios from "axios";

const currentPageTitle = ref("Settings");
const isSaving = ref(false);
const isTesting = ref(false);

const settings = ref({
  url: "",
  timeout: 5,
  checkIPInterval: 15,
  speedCheckInterval: 360,
  username: "",
  password: "",
  skipSSLVerify: true,
  telegramEnabled: false,
  telegramToken: "",
  telegramChatID: "",
  notifyOnDown: true,
  notifyOnRecovery: true,
  notifyOnIPChange: false,
  notifyOnIPStuck: true,
  notifyOnLowSpeed: false,
  lowSpeedThreshold: 10,
  notifyDailySummary: false,
  dailySummaryTime: "09:00",
});

const fetchSettings = async () => {
  try {
    const response = await axios.get("/api/settings");
    settings.value = { ...settings.value, ...response.data.data };
  } catch (error) {
    console.error("Failed to fetch settings:", error);
  }
};

const saveSettings = async () => {
  isSaving.value = true;
  try {
    await axios.put("/api/settings", settings.value);
    alert("Settings saved successfully!");
  } catch (error) {
    console.error("Failed to save settings:", error);
    alert("Failed to save settings");
  } finally {
    isSaving.value = false;
  }
};

const testNotification = async () => {
  isTesting.value = true;
  try {
    const response = await axios.post("/api/testNotification", {
      message: "This is a test notification from Proxy Checker!",
    });
    alert(response.data.message || "Test notification sent successfully!");
  } catch (error) {
    console.error("Failed to send test notification:", error);
    alert(error.response?.data?.error || "Failed to send test notification");
  } finally {
    isTesting.value = false;
  }
};

onMounted(() => {
  fetchSettings();
});
</script>
