<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">
          {{ $t('nav.payments') }}
        </h1>
        <p class="text-gray-600">
          {{ $t('payments.subtitle') }}
        </p>
      </div>

      <!-- Action Bar -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
        <div class="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
          <!-- Search -->
          <div class="flex-1 max-w-md">
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="$t('payments.search_placeholder')"
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="handleSearch"
              >
            </div>
          </div>
          
          <!-- Add Button -->
          <button
            @click="openCreateModal"
            class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors"
          >
            <PlusIcon class="h-5 w-5" />
            <span>{{ $t('payments.add_payment') }}</span>
          </button>
        </div>
      </div>

      <!-- Payments Table -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('payments.payment_reference') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('payments.invoice') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('payments.client') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('payments.method') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('payments.amount') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('payments.date') }}
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('common.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="loading" v-for="n in 5" :key="n">
                <td v-for="m in 7" :key="m" class="px-6 py-4">
                  <div class="animate-pulse bg-gray-200 h-4 rounded"></div>
                </td>
              </tr>
              <tr v-else-if="payments.length === 0">
                <td colspan="7" class="px-6 py-12 text-center text-gray-500">
                  <CreditCardIcon class="mx-auto h-12 w-12 text-gray-400 mb-4" />
                  <p class="text-lg font-medium mb-2">{{ $t('payments.no_payments') }}</p>
                  <p class="text-sm">{{ $t('payments.no_payments_subtitle') }}</p>
                </td>
              </tr>
              <tr v-else v-for="payment in payments" :key="payment.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">
                    {{ payment.reference || 'N/A' }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ getInvoiceNumber(payment.invoice_id) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ getClientName(payment.client_id) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900 capitalize">
                    {{ payment.method }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ formatCurrency(payment.amount_cents || 0) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500">
                    {{ formatDate(payment.payment_date) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div class="flex items-center justify-end space-x-2">
                    <button
                      @click="viewPayment(payment)"
                      class="text-blue-600 hover:text-blue-900 p-1 rounded transition-colors"
                      :title="$t('common.view')"
                    >
                      <EyeIcon class="h-4 w-4" />
                    </button>
                    <button
                      @click="editPayment(payment)"
                      class="text-green-600 hover:text-green-900 p-1 rounded transition-colors"
                      :title="$t('common.edit')"
                    >
                      <PencilIcon class="h-4 w-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  MagnifyingGlassIcon,
  PlusIcon,
  EyeIcon,
  PencilIcon,
  CreditCardIcon
} from '@heroicons/vue/24/outline'

const { t } = useI18n()

// State
const payments = ref<any[]>([])
const invoices = ref<any[]>([])
const clients = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref('')

// Methods
const fetchPayments = async () => {
  try {
    loading.value = true
    error.value = null
    
    // TODO: Implement GetPayments API call
    // const result = await GetPayments(searchQuery.value, 20, 0)
    // payments.value = result.data || []
    
    // Mock data for now
    payments.value = []
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch payments'
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchPayments()
}

const getInvoiceNumber = (invoiceId: number) => {
  const invoice = invoices.value.find((i: any) => i.id === invoiceId)
  return invoice ? invoice.invoice_number : 'N/A'
}

const getClientName = (clientId: number) => {
  const client = clients.value.find((c: any) => c.id === clientId)
  return client ? client.name : 'Unknown Client'
}

const formatCurrency = (cents: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(cents / 100)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const openCreateModal = () => {
  // TODO: Implement create payment modal
  console.log('Create payment')
}

const viewPayment = (payment: any) => {
  // TODO: Implement view payment
  console.log('View payment:', payment)
}

const editPayment = (payment: any) => {
  // TODO: Implement edit payment
  console.log('Edit payment:', payment)
}

// Lifecycle
onMounted(() => {
  fetchPayments()
})
</script>
