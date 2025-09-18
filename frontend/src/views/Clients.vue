<template>
  <div>
    <div class="mb-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 mb-2">
            {{ $t("clients.title") }}
          </h1>
          <p class="text-gray-600">
            {{ $t("clients.list") }}
          </p>
        </div>
        <button @click="showCreateModal = true" class="btn btn-primary">
          {{ $t("clients.create") }}
        </button>
      </div>
    </div>

    <!-- Search Bar -->
    <div class="mb-6">
      <div class="max-w-md">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('clients.search_placeholder')"
          class="form-input"
          @input="debouncedSearch"
        />
      </div>
    </div>

    <!-- Clients Table -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th>{{ $t("fields.name") }}</th>
              <th>{{ $t("fields.phone") }}</th>
              <th>{{ $t("fields.debt") }}</th>
              <th>{{ $t("fields.address") }}</th>
              <th class="text-center">{{ $t("fields.actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="5" class="text-center py-8">
                <div class="spinner mx-auto"></div>
                <p class="mt-2 text-gray-500">{{ $t("messages.loading") }}</p>
              </td>
            </tr>
            <tr v-else-if="clients.length === 0">
              <td colspan="5" class="text-center py-8">
                <p class="text-gray-500">{{ $t("messages.no_data") }}</p>
              </td>
            </tr>
            <tr v-else v-for="client in clients" :key="client.id">
              <td class="font-medium">{{ client.name }}</td>
              <td>{{ client.phone || "---" }}</td>
              <td>{{ formatCurrency(client.debt_cents || 0) }}</td>
              <td>{{ client.address || "---" }}</td>
              <td class="text-center">
                <div class="inline-flex items-center justify-center gap-2">
                  <button
                    @click="editClient(client)"
                    class="text-blue-600 hover:text-blue-900 text-sm"
                  >
                    {{ $t("actions.edit") }}
                  </button>
                  <button
                    @click="openAdjustDebt(client)"
                    class="text-amber-600 hover:text-amber-800 text-sm"
                  >
                    {{ $t("clients.adjust_debt") }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="mt-6 flex items-center justify-between">
      <div class="text-sm text-gray-700">
        {{ $t("pagination.showing") }} {{ (currentPage - 1) * pageSize + 1 }}
        {{ $t("pagination.to") }}
        {{ Math.min(currentPage * pageSize, totalCount) }}
        {{ $t("pagination.of") }} {{ totalCount }}
        {{ $t("pagination.results") }}
      </div>
      <div class="flex space-x-2 space-x-reverse">
        <button
          @click="previousPage"
          :disabled="currentPage === 1"
          class="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ $t("pagination.previous") }}
        </button>
        <button
          @click="nextPage"
          :disabled="currentPage === totalPages"
          class="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ $t("pagination.next") }}
        </button>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div
      v-if="showCreateModal || showEditModal"
      class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50"
      @click="closeModal"
    >
      <div
        class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white"
        @click.stop
      >
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 mb-4">
            {{ showCreateModal ? $t("clients.create") : $t("clients.edit") }}
          </h3>

          <!-- Error Message -->
          <div
            v-if="errorMessage"
            class="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded"
          >
            {{ errorMessage }}
          </div>

          <!-- Success Message -->
          <div
            v-if="successMessage"
            class="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded"
          >
            {{ successMessage }}
          </div>

          <form @submit.prevent="saveClient" class="space-y-4">
            <div>
              <label class="form-label">{{ $t("fields.name") }} *</label>
              <input
                v-model="currentClient.name"
                type="text"
                required
                class="form-input"
              />
            </div>

            <div>
              <label class="form-label">{{ $t("fields.phone") }}</label>
              <input
                v-model="currentClient.phone"
                type="tel"
                class="form-input"
              />
            </div>

            <!-- debt is not editable here; use adjust modal -->

            <div>
              <label class="form-label">{{ $t("fields.address") }}</label>
              <textarea
                v-model="currentClient.address"
                rows="3"
                class="form-input"
              ></textarea>
            </div>

            <div class="flex justify-end space-x-3 space-x-reverse pt-4">
              <button
                type="button"
                @click="closeModal"
                class="btn btn-secondary"
              >
                {{ $t("actions.cancel") }}
              </button>
              <button type="submit" :disabled="loading" class="btn btn-primary">
                {{ loading ? $t("messages.loading") : $t("actions.save") }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
    <!-- Adjust Debt Modal -->
    <div
      v-if="showAdjustModal"
      class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50"
      @click="closeAdjustModal"
    >
      <div
        class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white"
        @click.stop
      >
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 mb-4">
            {{ $t("clients.adjust_debt") }}
          </h3>
          <p class="mb-3 text-sm text-gray-600">
            {{ $t("clients.adjust_debt_hint") }}
          </p>
          <div class="mb-4">
            <label class="form-label">{{ $t("fields.amount") }}</label>
            <input
              v-model.number="adjustAmount"
              type="number"
              step="0.01"
              class="form-input"
            />
          </div>
          <div class="flex justify-end space-x-3 space-x-reverse">
            <button
              type="button"
              @click="closeAdjustModal"
              class="btn btn-secondary"
            >
              {{ $t("actions.cancel") }}
            </button>
            <button
              type="button"
              @click="confirmAdjustDebt"
              class="btn btn-primary"
            >
              {{ $t("actions.save") }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useClientStore } from "../stores/clients";

const { t } = useI18n();
const clientStore = useClientStore();

// Reactive data
const clients = ref<any[]>([]);
const loading = ref(false);
const searchQuery = ref("");
const currentPage = ref(1);
const pageSize = ref(20);
const totalCount = ref(0);
const showCreateModal = ref(false);
const showEditModal = ref(false);
const errorMessage = ref("");
const successMessage = ref("");
const currentClient = ref({
  id: undefined as number | undefined,
  name: "",
  phone: "",
  debt_cents: 0,
  address: "",
});

// Computed
const totalPages = computed(() => Math.ceil(totalCount.value / pageSize.value));

// Methods
async function loadClients() {
  try {
    loading.value = true;
    const offset = (currentPage.value - 1) * pageSize.value;
    const result = await clientStore.fetchClients(
      searchQuery.value,
      pageSize.value,
      offset
    );
    clients.value = result.data;
    totalCount.value = result.total;
  } catch (error) {
    console.error("Failed to load clients:", error);
  } finally {
    loading.value = false;
  }
}

function debouncedSearch() {
  // Simple debounce - reset to first page and reload
  currentPage.value = 1;
  setTimeout(() => {
    loadClients();
  }, 300);
}

function previousPage() {
  if (currentPage.value > 1) {
    currentPage.value--;
    loadClients();
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
    loadClients();
  }
}

function editClient(client: any) {
  currentClient.value = { ...client };
  showEditModal.value = true;
}

function closeModal() {
  showCreateModal.value = false;
  showEditModal.value = false;
  errorMessage.value = "";
  successMessage.value = "";
  currentClient.value = {
    id: undefined,
    name: "",
    phone: "",
    debt_cents: 0,
    address: "",
  };
}

async function saveClient() {
  try {
    loading.value = true;
    errorMessage.value = "";
    successMessage.value = "";

    if (showCreateModal.value) {
      await clientStore.createClient(currentClient.value);
      successMessage.value = "Client created successfully!";
    } else {
      await clientStore.updateClient(currentClient.value);
      successMessage.value = "Client updated successfully!";
    }

    closeModal();
    await loadClients();
  } catch (error) {
    console.error("Failed to save client:", error);
    errorMessage.value =
      error instanceof Error ? error.message : "Failed to save client";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadClients();
});

function formatCurrency(cents: number) {
  const amount = (cents / 100).toFixed(2);
  return `${amount} ${t("currency.symbol")}`;
}

const showAdjustModal = ref(false);
const adjustTarget = ref<any>(null);
const adjustAmount = ref<number>(0);

function openAdjustDebt(client: any) {
  adjustTarget.value = client;
  adjustAmount.value = 0;
  showAdjustModal.value = true;
}

function closeAdjustModal() {
  showAdjustModal.value = false;
  adjustTarget.value = null;
  adjustAmount.value = 0;
}

async function confirmAdjustDebt() {
  if (!adjustTarget.value) return;
  const delta = Math.round(adjustAmount.value * 100);
  if (delta === 0) {
    closeAdjustModal();
    return;
  }
  if (!confirm(t("clients.confirm_adjust"))) return;
  try {
    await clientStore.adjustDebt(adjustTarget.value.id, delta);
    await loadClients();
    closeAdjustModal();
  } catch (e) {
    console.error(e);
  }
}
</script>
