<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">
          {{ $t('nav.invoices') }}
        </h1>
        <p class="text-gray-600">
          {{ $t('invoices.subtitle') }}
        </p>
      </div>

      <!-- Action Bar -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
        <div class="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
          <!-- Search -->
          <div class="flex-1 max-w-md">
            <div class="relative">
              <MagnifyingGlassIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
              <input v-model="searchQuery" type="text" :placeholder="$t('invoices.search_placeholder')"
                class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="handleSearch">
            </div>
          </div>

          <!-- Add Button -->
          <button @click="openCreateModal"
            class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors">
            <PlusIcon class="h-5 w-5" />
            <span>{{ $t('invoices.add_invoice') }}</span>
          </button>
        </div>
      </div>

      <!-- Invoices Table -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('invoices.invoice_number') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('invoices.client') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('fields.status') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('invoices.total') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('invoices.due_date') }}
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('invoices.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="loading" v-for="n in 5" :key="n">
                <td v-for="m in 6" :key="m" class="px-6 py-4">
                  <div class="animate-pulse bg-gray-200 h-4 rounded"></div>
                </td>
              </tr>
              <tr v-else-if="invoices.length === 0">
                <td colspan="6" class="px-6 py-12 text-center text-gray-500">
                  <DocumentTextIcon class="mx-auto h-12 w-12 text-gray-400 mb-4" />
                  <p class="text-lg font-medium mb-2">{{ $t('invoices.no_invoices') }}</p>
                  <p class="text-sm">{{ $t('invoices.no_invoices_subtitle') }}</p>
                </td>
              </tr>
              <tr v-else v-for="invoice in invoices" :key="invoice.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">
                    {{ invoice.invoice.invoice_number }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ invoice.client.name }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusColor(invoice.invoice.status)"
                    class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ $t(`invoices.status.${invoice.invoice.status}`) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ formatPrice(invoice.total_amount_cents || 0) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500">
                    {{ formatDate(invoice.invoice.issue_date) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div class="flex items-center justify-end space-x-2">
                    <button @click="viewInvoice(invoice)"
                      class="text-blue-600 hover:text-blue-900 p-1 rounded transition-colors"
                      :title="$t('common.view')">
                      <EyeIcon class="h-4 w-4" />
                    </button>
                    <button @click="editInvoice(invoice)"
                      class="text-green-600 hover:text-green-900 p-1 rounded transition-colors"
                      :title="$t('common.edit')">
                      <PencilIcon class="h-4 w-4" />
                    </button>
                    <button @click="exportInvoicePDF(invoice)"
                      class="text-purple-600 hover:text-purple-900 p-1 rounded transition-colors"
                      :title="$t('invoices.export_pdf')">
                      <DocumentArrowDownIcon class="h-4 w-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <!-- Create Invoice Modal -->
      <div v-if="showCreateModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50"
        @click="closeCreateModal">
        <div class="relative top-10 mx-auto p-5 border w-full max-w-4xl shadow-lg rounded-md bg-white" @click.stop>
          <h3 class="text-lg font-medium mb-4">{{ $t('invoices.create') }}</h3>

          <form @submit.prevent="saveInvoice" class="space-y-4">
            <div v-if="error" class="bg-red-50 border border-red-200 text-red-800 px-3 py-2 rounded">
              {{ error }}
            </div>
            <!-- Client search + Notes (two columns like create order) -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="form-label">{{ $t('fields.client') }}</label>
                <div class="relative">
                  <input v-model="clientSearch" type="text" class="form-input"
                    :placeholder="$t('orders.search_client_placeholder')" @input="filterClients"
                    @focus="showClientDropdown = true" @blur="hideClientDropdown" />
                  <div v-if="showClientDropdown && filteredClients.length > 0"
                    class="absolute z-20 w-full bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto">
                    <div v-for="c in filteredClients" :key="c.id" @mousedown.prevent="selectClient(c)"
                      class="px-3 py-2 hover:bg-gray-100 cursor-pointer">
                      <div class="font-medium">{{ c.name }}</div>
                      <div class="text-sm text-gray-500" v-if="c.phone">{{ c.phone }}</div>
                    </div>
                  </div>
                </div>
              </div>

              <div>
                <label class="form-label">{{ $t('fields.notes') }}</label>
                <input v-model="newInvoice.notes" type="text" class="form-input"
                  :placeholder="$t('orders.notes_placeholder')" />
              </div>
            </div>

            <!-- Items -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="form-label">{{ $t('orders.items') }}</label>
                <button type="button" @click="addInvoiceItem"
                  class="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm">{{ $t('actions.add')
                  }}</button>
              </div>

              <div class="space-y-3">
                <div v-for="(item, idx) in newInvoice.items" :key="idx" class="border rounded-lg p-3 bg-gray-50">
                  <div class="grid grid-cols-1 md:grid-cols-6 gap-3 items-end">
                    <div>
                      <label class="form-label text-xs">{{ $t('fields.product') }}</label>
                      <div class="relative">
                        <input v-model="item.productSearch" type="text" class="form-input text-sm"
                          :placeholder="$t('orders.search_product_placeholder')" @input="filterProducts(idx, $event)"
                          @focus="onFocusProduct(idx)" @blur="hideProductDropdown(idx)" />
                        <div
                          v-if="showProductDropdown[idx] && filteredProducts[idx] && filteredProducts[idx].length > 0"
                          class="absolute z-30 w-full bg-white border border-gray-200 rounded-md shadow-md max-h-56 overflow-y-auto">
                          <div v-for="p in filteredProducts[idx]" :key="p.id"
                            @mousedown.prevent="selectProductForItem(idx, p)"
                            class="px-3 py-3 hover:bg-gray-50 cursor-pointer border-b last:border-b-0">
                            <div class="flex items-center justify-between">
                              <div class="font-medium text-gray-800">{{ p.name }}</div>
                              <div class="text-sm text-gray-600">{{ formatCurrency(p.unit_price_cents) }}</div>
                            </div>
                            <div class="text-sm text-gray-500 mt-1"> <span v-if="p.sku">SKU: {{ p.sku }}</span></div>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div>
                      <label class="form-label text-xs">{{ $t('fields.quantity') }}</label>
                      <input v-model.number="item.qty" type="number" min="1" class="form-input text-sm"
                        @input="recalcItem(idx)" />
                    </div>

                    <div>
                      <label class="form-label text-xs">{{ $t('fields.unit_price') }}</label>
                      <div class="relative">
                        <input v-model.number="item.unit_price" type="number" step="0.01" min="0"
                          class="form-input text-sm pr-14" @input="recalcItem(idx)" />
                        <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-600">DZD</span>
                      </div>
                    </div>

                    <div>
                      <label class="form-label text-xs">{{ $t('fields.total') }}</label>
                      <div class="text-sm font-medium">{{ formatCurrency((item.qty || 0) * (item.unit_price || 0) * 100)
                        }}</div>
                    </div>

                    <div class="flex items-center">
                      <button type="button" @click="removeInvoiceItem(idx)"
                        class="text-red-600 hover:text-red-800 p-1 bg-white rounded-full border border-gray-100 shadow-sm">
                        <TrashIcon class="h-5 w-5" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Totals and notes -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label class="form-label">{{ $t('orders.discount_percent') }} %</label>
                <input v-model.number="newInvoice.discount_percent" type="number" min="0" max="100"
                  class="form-input" />
              </div>
              <div>
                <label class="form-label">{{ $t('orders.tax_percent') }} %</label>
                <input v-model.number="newInvoice.tax_percent" type="number" min="0" max="100" class="form-input" />
              </div>
              <div class="flex flex-col justify-end">
                <div class="text-lg font-bold">{{ $t('fields.total') }}: {{ formatCurrency(calculateInvoiceTotal()) }}
                </div>
              </div>
            </div>

            <!-- notes moved to top next to client -->

            <div class="flex justify-end space-x-3">
              <button type="button" class="btn btn-secondary" @click="closeCreateModal">{{ $t('actions.cancel')
              }}</button>
              <button type="submit" :disabled="loading || !newInvoice.client_id || newInvoice.items.length === 0"
                class="btn btn-primary">{{ loading ? $t('messages.loading') : $t('actions.create') }}</button>
            </div>
          </form>
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
  DocumentTextIcon,
  DocumentArrowDownIcon,
  TrashIcon
} from '@heroicons/vue/24/outline'

const { t } = useI18n()

// State
const invoices = ref<any[]>([])
const clients = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref('')

const showCreateModal = ref(false)
const newInvoice = ref({
  client_id: '',
  notes: '',
  discount_percent: 0,
  tax_percent: 0,
  items: [] as Array<any>
})
const clientSearch = ref('')
const showClientDropdown = ref(false)
const filteredClients = ref<any[]>([])
const showProductDropdown = ref<Record<number, boolean>>({})
const filteredProducts = ref<Record<number, any[]>>({})
const productsCache = ref<any[]>([])

const addInvoiceItem = () => {
  const idx = newInvoice.value.items.length
  newInvoice.value.items.push({ product_id: '', name_snapshot: '', productSearch: '', qty: 1, unit_price: 0, currency: 'DZD' })
  showProductDropdown.value[idx] = false
  filteredProducts.value[idx] = []
}

const removeInvoiceItem = (idx: number) => {
  newInvoice.value.items.splice(idx, 1)
  delete showProductDropdown.value[idx]
  delete filteredProducts.value[idx]
}

const filterClients = () => {
  const q = clientSearch.value.toLowerCase()
  if (!q) {
    filteredClients.value = clients.value.slice(0, 20)
    return
  }
  filteredClients.value = clients.value.filter((c: any) => c.name.toLowerCase().includes(q) || (c.phone && c.phone.includes(q))).slice(0, 20)
}

const selectClient = (c: any) => {
  newInvoice.value.client_id = c.id.toString()
  clientSearch.value = c.name
  showClientDropdown.value = false
}

const hideClientDropdown = () => { setTimeout(() => { showClientDropdown.value = false }, 150) }

const filterProducts = (index: number, event: Event) => {
  const target = event.target as HTMLInputElement
  const q = target.value.toLowerCase()
  if (!q) {
    // show cached products when query is empty
    filteredProducts.value[index] = productsCache.value.slice(0, 20)
    return
  }
  // fetch products for the query
  ;(async () => {
    try {
      const res = await GetProducts(q, 20, 0)
      // @ts-ignore
      filteredProducts.value[index] = res.data || []
    } catch (e) {
      console.error(e)
    }
  })()
}

const onFocusProduct = async (index: number) => {
  showProductDropdown.value[index] = true
  // if we have a cache, show it immediately
  if (productsCache.value && productsCache.value.length > 0) {
    filteredProducts.value[index] = productsCache.value.slice(0, 20)
    return
  }
  // otherwise fetch a small list
  try {
    const res = await GetProducts('', 20, 0)
    // @ts-ignore
    productsCache.value = res.data || []
    filteredProducts.value[index] = productsCache.value.slice(0, 20)
  } catch (e) {
    console.error(e)
  }
}

const selectProductForItem = (index: number, p: any) => {
  const it = newInvoice.value.items[index]
  it.product_id = p.id.toString()
  it.name_snapshot = p.name
  it.productSearch = p.name
  it.unit_price = p.unit_price_cents / 100
  it.currency = p.currency || 'DZD'
  showProductDropdown.value[index] = false
  filteredProducts.value[index] = []
}

const recalcItem = (index: number) => { /* UI only, backend will recalc */ }

const calculateInvoiceTotal = () => {
  const subtotal = newInvoice.value.items.reduce((s, it) => s + ((it.qty || 0) * (it.unit_price || 0) * 100), 0)
  const discount = subtotal * (newInvoice.value.discount_percent / 100)
  const afterDiscount = subtotal - discount
  const tax = afterDiscount * (newInvoice.value.tax_percent / 100)
  return afterDiscount + tax
}

const closeCreateModal = () => {
  showCreateModal.value = false
  newInvoice.value = { client_id: '', notes: '', discount_percent: 0, tax_percent: 0, items: [] }
  clientSearch.value = ''
  showClientDropdown.value = false
  filteredClients.value = []
  showProductDropdown.value = {}
  filteredProducts.value = {}
}

const hideProductDropdown = (index: number) => { setTimeout(() => { showProductDropdown.value[index] = false }, 150) }

// Import Wails bindings
import { CreateInvoice, GetInvoices, GetInvoice, ExportInvoicePDF, GetClients, GetProducts } from '../../wailsjs/go/main/App'

// Methods
const fetchInvoices = async () => {
  try {
    loading.value = true
    error.value = null

    const result = await GetInvoices(50, 0)
    invoices.value = result.data || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch invoices'
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  fetchInvoices()
}

function formatPrice(priceCents: number): string {
  return new Intl.NumberFormat('ar-DZ', {
    style: 'currency',
    currency: 'DZD'
  }).format(priceCents / 100) // Convert cents to dinars
}

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    DRAFT: 'bg-gray-100 text-gray-800',
    ISSUED: 'bg-blue-100 text-blue-800',
    VIEWED: 'bg-yellow-100 text-yellow-800',
    PAID: 'bg-green-100 text-green-800',
    OVERDUE: 'bg-red-100 text-red-800',
    CANCELLED: 'bg-red-100 text-red-800'
  }
  return colors[status] || 'bg-gray-100 text-gray-800'
}

const formatCurrency = (cents: number) => {
  return new Intl.NumberFormat('en-DZ', {
    style: 'currency',
    currency: 'DZD'
  }).format(cents / 100)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const openCreateModal = async () => {
  showCreateModal.value = true
  // fetch clients for selection
  try {
    const clientsRes = await GetClients('', 100, 0)
    clients.value = clientsRes.data || []
  } catch (e) {
    console.error(e)
  }
}

const viewInvoice = (invoice: any) => {
  // TODO: Implement view invoice
  console.log('View invoice:', invoice)
}

const editInvoice = (invoice: any) => {
  // TODO: Implement edit invoice
  console.log('Edit invoice:', invoice)
}

const saveInvoice = async () => {
  try {
    loading.value = true
    error.value = null

    const clientIdNum = parseInt(newInvoice.value.client_id)
    if (!clientIdNum || isNaN(clientIdNum)) {
      error.value = 'Please select a client'
      return
    }

    if (!newInvoice.value.items || newInvoice.value.items.length === 0) {
      error.value = 'Add at least one item'
      return
    }

    // Build items and validate
    const items = newInvoice.value.items.map((i, idx) => {
      const qty = i.qty || 0
      const unitPrice = i.unit_price || 0
      if (qty <= 0) throw new Error(`Quantity must be > 0 for item ${idx + 1}`)
      if (unitPrice <= 0) throw new Error(`Unit price must be > 0 for item ${idx + 1}`)
      return {
        product_id: i.product_id ? parseInt(i.product_id) : null,
        name_snapshot: i.name_snapshot || '',
        qty: qty,
        unit_price_cents: Math.round(unitPrice * 100),
        currency: i.currency || 'DZD'
      }
    })

    await CreateInvoice(clientIdNum, newInvoice.value.notes, newInvoice.value.discount_percent, newInvoice.value.tax_percent, items)
    showCreateModal.value = false
    await fetchInvoices()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

const exportInvoicePDF = async (invoice: any) => {
  try {
    const bytes = await ExportInvoicePDF(invoice.id)
    const u8 = Uint8Array.from(bytes as any)
    const blob = new Blob([u8], { type: 'application/pdf' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `invoice-${invoice.invoice_number}.pdf`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to export PDF'
  }
}

// Lifecycle
onMounted(() => {
  fetchInvoices()
})
</script>
