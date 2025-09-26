<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">
          {{ $t("nav.payments") }}
        </h1>
        <p class="text-gray-600">
          {{ $t("payments.debt_history_subtitle") }}
        </p>
      </div>

      <!-- Debt Payments Table -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("fields.client") }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("payments.previous_debt") }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("payments.new_debt") }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("payments.adjustment") }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("payments.type") }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("fields.date") }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                >
                  {{ $t("fields.notes") }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="loading" v-for="n in 5" :key="n">
                <td v-for="m in 7" :key="m" class="px-6 py-4">
                  <div class="animate-pulse bg-gray-200 h-4 rounded"></div>
                </td>
              </tr>
              <tr v-else-if="debtPayments.length === 0">
                <td colspan="7" class="px-6 py-12 text-center text-gray-500">
                  <CreditCardIcon
                    class="mx-auto h-12 w-12 text-gray-400 mb-4"
                  />
                  <p class="text-lg font-medium mb-2">
                    {{ $t("payments.no_debt_payments") }}
                  </p>
                  <p class="text-sm">
                    {{ $t("payments.no_debt_payments_subtitle") }}
                  </p>
                </td>
              </tr>
              <tr
                v-else
                v-for="payment in debtPayments"
                :key="payment.debt_payment.id"
                class="hover:bg-gray-50"
              >
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">
                    {{ payment.client.name }}
                  </div>
                  <div
                    v-if="payment.client.phone"
                    class="text-sm text-gray-500"
                  >
                    {{ payment.client.phone }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{
                      formatCurrency(
                        payment.debt_payment.previous_debt_cents || 0
                      )
                    }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{
                      formatCurrency(payment.debt_payment.new_debt_cents || 0)
                    }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div
                    class="text-sm font-medium"
                    :class="
                      payment.debt_payment.adjustment_cents >= 0
                        ? 'text-red-600'
                        : 'text-green-600'
                    "
                  >
                    {{ payment.debt_payment.adjustment_cents >= 0 ? "+" : ""
                    }}{{
                      formatCurrency(payment.debt_payment.adjustment_cents || 0)
                    }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span
                    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                    :class="
                      payment.debt_payment.type === 'INCREASE'
                        ? 'bg-red-100 text-red-800'
                        : 'bg-green-100 text-green-800'
                    "
                  >
                    {{
                      payment.debt_payment.type === "INCREASE"
                        ? $t("payments.debt_increase")
                        : $t("payments.debt_decrease")
                    }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500">
                    {{ formatDate(payment.debt_payment.created_at) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500 max-w-xs truncate">
                    {{ payment.debt_payment.notes || "-" }}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div
          v-if="!loading && debtPayments.length > 0"
          class="bg-white px-4 py-3 border-t border-gray-200 sm:px-6"
        >
          <div class="flex items-center justify-between">
            <div class="text-sm text-gray-700">
              {{ $t("pagination.showing") }} {{ offset + 1 }}
              {{ $t("pagination.to") }} {{ Math.min(offset + limit, total) }}
              {{ $t("pagination.of") }} {{ total }}
              {{ $t("pagination.results") }}
            </div>
            <div class="flex space-x-2">
              <button
                @click="previousPage"
                :disabled="offset === 0"
                class="px-3 py-1 text-sm bg-white border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ $t("pagination.previous") }}
              </button>
              <button
                @click="nextPage"
                :disabled="offset + limit >= total"
                class="px-3 py-1 text-sm bg-white border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ $t("pagination.next") }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useClientStore } from "../stores/clients";
import { CreditCardIcon } from "@heroicons/vue/24/outline";

const { t } = useI18n();
const clientStore = useClientStore();

// State
const debtPayments = ref<any[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const total = ref(0);
const limit = ref(20);
const offset = ref(0);

// Methods
const fetchDebtPayments = async () => {
  try {
    loading.value = true;
    error.value = null;

    const result = await clientStore.fetchDebtPayments(
      limit.value,
      offset.value
    );
    debtPayments.value = result.data || [];
    total.value = result.total || 0;
  } catch (err) {
    error.value =
      err instanceof Error ? err.message : "Failed to fetch debt payments";
    console.error("Error fetching debt payments:", err);
  } finally {
    loading.value = false;
  }
};

const formatCurrency = (cents: number) => {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(cents / 100);
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString();
};

const previousPage = () => {
  if (offset.value > 0) {
    offset.value = Math.max(0, offset.value - limit.value);
    fetchDebtPayments();
  }
};

const nextPage = () => {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value;
    fetchDebtPayments();
  }
};

// Lifecycle
onMounted(() => {
  fetchDebtPayments();
});
</script>
