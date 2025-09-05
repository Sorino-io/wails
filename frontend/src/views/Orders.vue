<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-7xl mx-auto">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-2">
          {{ $t('nav.orders') }}
        </h1>
        <p class="text-gray-600">
          {{ $t('orders.subtitle') }}
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
                :placeholder="$t('orders.search_placeholder')"
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
            <span>{{ $t('orders.add_order') }}</span>
          </button>
        </div>
      </div>

      <!-- Orders Table -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('orders.order_number') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('orders.client') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('fields.status') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('orders.total') }}
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('orders.date') }}
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {{ $t('orders.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="loading" v-for="n in 5" :key="n">
                <td v-for="m in 6" :key="m" class="px-6 py-4">
                  <div class="animate-pulse bg-gray-200 h-4 rounded"></div>
                </td>
              </tr>
              <tr v-else-if="orders.length === 0">
                <td colspan="6" class="px-6 py-12 text-center text-gray-500">
                  <DocumentIcon class="mx-auto h-12 w-12 text-gray-400 mb-4" />
                  <p class="text-lg font-medium mb-2">{{ $t('orders.no_orders') }}</p>
                  <p class="text-sm">{{ $t('orders.no_orders_subtitle') }}</p>
                </td>
              </tr>
              <tr v-else v-for="order in orders" :key="order.order.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">
                    {{ order.order.order_number }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ order.client.name }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusColor(order.order.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ $t(`orders.order_status.${order.order.status.toLowerCase()}`) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">
                    {{ formatPrice(order.total_cents || 0) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-500">
                    {{ formatDate(order.order.created_at) }}
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div class="flex items-center justify-end space-x-2">
                    <button
                      @click="viewOrder(order)"
                      class="text-blue-600 hover:text-blue-900 p-1 rounded transition-colors"
                      :title="$t('common.view')"
                    >
                      <EyeIcon class="h-4 w-4" />
                    </button>
                    <button
                      @click="editOrder(order)"
                      class="text-green-600 hover:text-green-900 p-1 rounded transition-colors"
                      :title="$t('common.edit')"
                    >
                      <PencilIcon class="h-4 w-4" />
                    </button>
                    <button
                      @click="exportOrderPDF(order)"
                      class="text-purple-600 hover:text-purple-900 p-1 rounded transition-colors"
                      :title="$t('orders.export_pdf')"
                    >
                      <DocumentArrowDownIcon class="h-4 w-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalOrders > 0" class="bg-white px-4 py-3 border-t border-gray-200 sm:px-6">
          <div class="flex items-center justify-between">
            <div class="flex-1 flex justify-between sm:hidden">
              <button
                @click="prevPage"
                :disabled="currentPage === 1"
                class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                {{ $t('common.previous') }}
              </button>
              <button
                @click="nextPage"
                :disabled="currentPage >= totalPages"
                class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                {{ $t('common.next') }}
              </button>
            </div>
            <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p class="text-sm text-gray-700">
                  {{ $t('common.showing') }} {{ ((currentPage - 1) * pageSize) + 1 }} {{ $t('common.to') }} 
                  {{ Math.min(currentPage * pageSize, totalOrders) }} {{ $t('common.of') }} {{ totalOrders }} {{ $t('common.results') }}
                </p>
              </div>
              <div>
                <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                  <button
                    @click="prevPage"
                    :disabled="currentPage === 1"
                    class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                  >
                    <ChevronLeftIcon class="h-5 w-5" />
                  </button>
                  <button
                    @click="nextPage"
                    :disabled="currentPage >= totalPages"
                    class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                  >
                    <ChevronRightIcon class="h-5 w-5" />
                  </button>
                </nav>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Create Order Modal -->
      <div
        v-if="showCreateModal"
        class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50"
        @click="closeCreateModal"
      >
        <div class="relative top-10 mx-auto p-5 border w-full max-w-4xl shadow-lg rounded-md bg-white" @click.stop>
          <div class="mt-3">
            <h3 class="text-lg font-medium text-gray-900 mb-4">
              {{ $t('orders.create_order') }}
            </h3>
            
            <form @submit.prevent="saveOrder" class="space-y-6">
              <!-- Client Selection -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="form-label">{{ $t('orders.client') }} *</label>
                  <div class="relative">
                    <input
                      v-model="clientSearch"
                      type="text"
                      required
                      class="form-input"
                      :placeholder="$t('orders.search_client_placeholder')"
                      @input="filterClients"
                      @focus="showClientDropdown = true"
                      @blur="hideClientDropdown"
                    />
                    <div
                      v-if="showClientDropdown && filteredClients.length > 0"
                      class="absolute z-10 w-full bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
                    >
                      <div
                        v-for="client in filteredClients"
                        :key="client.id"
                        @mousedown="selectClient(client)"
                        class="px-3 py-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
                      >
                        <div class="font-medium">{{ client.name }}</div>
                        <div class="text-sm text-gray-500" v-if="client.phone">{{ client.phone }}</div>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div>
                  <label class="form-label">{{ $t('orders.notes') }}</label>
                  <input
                    v-model="newOrder.notes"
                    type="text"
                    class="form-input"
                    :placeholder="$t('orders.notes_placeholder')"
                  />
                </div>
              </div>

              <!-- Order Items -->
              <div>
                <div class="flex items-center justify-between mb-4">
                  <label class="form-label">{{ $t('orders.items') }} *</label>
                  <button
                    type="button"
                    @click="addOrderItem"
                    class="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm"
                  >
                    {{ $t('orders.add_item') }}
                  </button>
                </div>
                
                <div class="space-y-3">
                  <div 
                    v-for="(item, index) in newOrder.items" 
                    :key="index"
                    class="border rounded-lg p-4 bg-gray-50"
                  >
                    <div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">
                      <div>
                        <label class="form-label text-xs">{{ $t('orders.product') }}</label>
                        <div class="relative">
                          <input
                            v-model="item.productSearch"
                            type="text"
                            class="form-input text-sm"
                            :placeholder="$t('orders.search_product_placeholder')"
                            @input="filterProducts(index, $event)"
                            @focus="showProductDropdown[index] = true"
                            @blur="hideProductDropdown(index)"
                          />
                          <div
                            v-if="showProductDropdown[index] && filteredProducts[index] && filteredProducts[index].length > 0"
                            class="absolute z-10 w-full bg-white border border-gray-300 rounded-md shadow-lg max-h-48 overflow-y-auto"
                          >
                            <div
                              v-for="product in filteredProducts[index]"
                              :key="product.id"
                              @mousedown="selectProductForItem(index, product)"
                              class="px-3 py-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
                            >
                              <div class="font-medium">{{ product.name }}</div>
                              <div class="text-sm text-gray-500">
                                {{ formatPrice(product.unit_price_cents) }}
                                <span v-if="product.sku" class="ml-2">SKU: {{ product.sku }}</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div>
                        <label class="form-label text-xs">{{ $t('orders.quantity') }} *</label>
                        <input
                          v-model.number="item.qty"
                          type="number"
                          min="1"
                          required
                          class="form-input text-sm"
                          @input="calculateItemTotal(index)"
                        />
                      </div>
                      
                      <div>
                        <label class="form-label text-xs">{{ $t('orders.unit_price') }} *</label>
                        <input
                          v-model.number="item.unit_price"
                          type="number"
                          step="0.01"
                          min="0"
                          required
                          class="form-input text-sm"
                          @input="calculateItemTotal(index)"
                        />
                      </div>
                      
                      <div class="flex items-center space-x-2">
                        <div class="text-sm font-medium">
                          {{ formatCurrency((item.qty || 0) * (item.unit_price || 0) * 100) }}
                        </div>
                        <button
                          type="button"
                          @click="removeOrderItem(index)"
                          class="text-red-600 hover:text-red-800 p-1"
                        >
                          <TrashIcon class="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Order Totals -->
              <div class="border-t pt-4">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <label class="form-label">{{ $t('orders.discount_percent') }}</label>
                    <input
                      v-model.number="newOrder.discount_percent"
                      type="number"
                      min="0"
                      max="100"
                      class="form-input"
                    />
                  </div>
                  
                  <div>
                    <label class="form-label">{{ $t('orders.tax_percent') }}</label>
                    <input
                      v-model.number="newOrder.tax_percent"
                      type="number"
                      min="0"
                      max="100"
                      class="form-input"
                    />
                  </div>
                  
                  <div class="flex flex-col justify-end">
                    <div class="text-lg font-bold">
                      {{ $t('orders.total') }}: {{ formatCurrency(calculateOrderTotal()) }}
                    </div>
                  </div>
                </div>
              </div>
            
              <div class="flex justify-end space-x-3 space-x-reverse pt-4">
                <button
                  type="button"
                  @click="closeCreateModal"
                  class="btn btn-secondary"
                >
                  {{ $t('actions.cancel') }}
                </button>
                <button
                  type="submit"
                  :disabled="loading || newOrder.items.length === 0"
                  class="btn btn-primary disabled:opacity-50"
                >
                  {{ loading ? $t('messages.loading') : $t('orders.create_order') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
      
      <!-- Order Detail / Edit Modal -->
      <div
        v-if="showDetailModal"
        class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50"
        @click="() => { showDetailModal = false }"
      >
        <div class="relative top-10 mx-auto p-5 border w-full max-w-3xl shadow-lg rounded-md bg-white" @click.stop>
          <div class="mt-3">
            <h3 class="text-lg font-medium text-gray-900 mb-4">
              {{ isEditing ? $t('orders.edit_order') : $t('orders.order_details') }}
            </h3>

            <div v-if="detailOrder">
              <div class="mb-4">
                <div class="text-sm font-semibold">{{ $t('orders.order_number') }}: {{ detailOrder.order.order_number }}</div>
                <div class="text-sm">{{ $t('orders.client') }}: {{ detailOrder.client.name }}</div>
                <div class="text-sm">{{ $t('orders.status') }}: {{ detailOrder.order.status }}</div>
              </div>

              <div class="mb-3">
                <h4 class="font-semibold">{{ $t('orders.items') }}</h4>
                <ul class="list-disc pl-5">
                  <li v-for="item in detailOrder.items" :key="item.id">
                    {{ item.name_snapshot }} — {{ item.qty }} × {{ formatCurrency(item.unit_price_cents) }} = {{ formatCurrency(item.total_cents) }}
                  </li>
                </ul>
              </div>

              <div class="flex justify-end space-x-3 space-x-reverse pt-4">
                <button v-if="!isEditing" @click="showDetailModal = false" class="btn btn-secondary">{{ $t('actions.close') }}</button>
                <button v-if="!isEditing" @click="beginEdit(detailOrder)" class="btn btn-primary">{{ $t('common.edit') }}</button>
                <button v-if="isEditing" @click="cancelEdit" class="btn btn-secondary">{{ $t('actions.cancel') }}</button>
                <button v-if="isEditing" @click="saveOrder" :disabled="loading" class="btn btn-primary">{{ loading ? $t('messages.loading') : $t('actions.save') }}</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  MagnifyingGlassIcon,
  PlusIcon,
  EyeIcon,
  PencilIcon,
  DocumentIcon,
  DocumentArrowDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  TrashIcon
} from '@heroicons/vue/24/outline'

// Import Wails functions
import { GetOrders, CreateOrder, GetClients, GetProducts, ExportOrderPDF, GetOrder, UpdateOrder } from '../../wailsjs/go/main/App'

const { t } = useI18n()

// State
const orders = ref<any[]>([])
const allClients = ref<any[]>([])
const allProducts = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref('')
const currentPage = ref(1)
const totalOrders = ref(0)
const pageSize = 20
const showCreateModal = ref(false)

// Search and dropdown states
const clientSearch = ref('')
const showClientDropdown = ref(false)
const filteredClients = ref<any[]>([])
const showProductDropdown = ref<Record<number, boolean>>({})
const filteredProducts = ref<Record<number, any[]>>({})

// New order form
const newOrder = ref({
  client_id: '',
  notes: '',
  discount_percent: 0,
  tax_percent: 0,
  items: [] as Array<{
    product_id: string
    name_snapshot: string
    sku_snapshot: string
    qty: number
    unit_price: number
    currency: string
    productSearch: string
  }>
})

// Detail and edit state
const showDetailModal = ref(false)
const detailOrder = ref<any | null>(null)
const isEditing = ref(false)
const editingOrderId = ref<number | null>(null)

// Computed
const totalPages = computed(() => Math.ceil(totalOrders.value / pageSize))

// Methods
const fetchOrders = async () => {
  try {
    loading.value = true
    error.value = null
    
    const result = await GetOrders(searchQuery.value, 0, '', pageSize, (currentPage.value - 1) * pageSize)
    orders.value = result.data || []
    totalOrders.value = result.total || 0
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch orders'
  } finally {
    loading.value = false
  }
}

const fetchClients = async () => {
  try {
    const result = await GetClients('', 100, 0)
    allClients.value = result.data || []
  } catch (err) {
    console.error('Failed to fetch clients:', err)
  }
}

const fetchProducts = async () => {
  try {
    const result = await GetProducts('', 100, 0)
    allProducts.value = result.data || []
  } catch (err) {
    console.error('Failed to fetch products:', err)
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchOrders()
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchOrders()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    fetchOrders()
  }
}

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    confirmed: 'bg-blue-100 text-blue-800',
    completed: 'bg-green-100 text-green-800',
    canceled: 'bg-red-100 text-red-800'
  }
  return colors[status.toLowerCase()] || 'bg-gray-100 text-gray-800'
}

const formatCurrency = (cents: number) => {
  return new Intl.NumberFormat('en-DZ', {
    style: 'currency',
    currency: 'DZD'
  }).format(cents / 100)
}

function formatPrice(priceCents: number): string {
  return new Intl.NumberFormat('ar-DZ', {
    style: 'currency',
    currency: 'DZD'
  }).format(priceCents / 100) // Convert cents to dinars
}

// const formatPrice = (cents: number) => {
//   return formatCurrency(cents)
// }

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

const openCreateModal = () => {
  showCreateModal.value = true
  fetchClients()
  fetchProducts()
  // Initialize filtered clients
  filteredClients.value = allClients.value
}

const closeCreateModal = () => {
  showCreateModal.value = false
  newOrder.value = {
    client_id: '',
    notes: '',
    discount_percent: 0,
    tax_percent: 0,
    items: []
  }
  // Reset search states
  clientSearch.value = ''
  showClientDropdown.value = false
  filteredClients.value = []
  showProductDropdown.value = {}
  filteredProducts.value = {}
}

const addOrderItem = () => {
  const newIndex = newOrder.value.items.length
  newOrder.value.items.push({
    product_id: '',
    name_snapshot: '',
    sku_snapshot: '',
    qty: 1,
    unit_price: 0,
    currency: 'USD',
    productSearch: ''
  })
  // Initialize dropdown state for new item
  showProductDropdown.value[newIndex] = false
  filteredProducts.value[newIndex] = []
}

const removeOrderItem = (index: number) => {
  newOrder.value.items.splice(index, 1)
  // Clean up dropdown states
  delete showProductDropdown.value[index]
  delete filteredProducts.value[index]
}

// Client search and selection functions
const filterClients = () => {
  const query = clientSearch.value.toLowerCase()
  if (query.length === 0) {
    filteredClients.value = allClients.value
  } else {
    filteredClients.value = allClients.value.filter(client => 
      client.name.toLowerCase().includes(query) || 
      (client.phone && client.phone.includes(query))
    )
  }
}

const selectClient = (client: any) => {
  newOrder.value.client_id = client.id.toString()
  clientSearch.value = client.name
  showClientDropdown.value = false
  filteredClients.value = []
}

const hideClientDropdown = () => {
  setTimeout(() => {
    showClientDropdown.value = false
  }, 200) // Delay to allow click on dropdown item
}

// Product search and selection functions
const filterProducts = (index: number, event: Event) => {
  const target = event.target as HTMLInputElement
  const query = target.value.toLowerCase()
  
  if (query.length === 0) {
    filteredProducts.value[index] = allProducts.value.slice(0, 10) // Limit to 10 items
  } else {
    filteredProducts.value[index] = allProducts.value.filter(product => 
      product.name.toLowerCase().includes(query) || 
      (product.sku && product.sku.toLowerCase().includes(query))
    ).slice(0, 10) // Limit to 10 items
  }
}

const selectProductForItem = (index: number, product: any) => {
  newOrder.value.items[index].product_id = product.id.toString()
  newOrder.value.items[index].name_snapshot = product.name
  newOrder.value.items[index].sku_snapshot = product.sku || ''
  newOrder.value.items[index].unit_price = product.unit_price_cents / 100
  newOrder.value.items[index].productSearch = product.name
  showProductDropdown.value[index] = false
  filteredProducts.value[index] = []
}

const hideProductDropdown = (index: number) => {
  setTimeout(() => {
    showProductDropdown.value[index] = false
  }, 200) // Delay to allow click on dropdown item
}

const calculateItemTotal = (index: number) => {
  // This is just for UI display, actual calculation happens on backend
}

const calculateOrderTotal = () => {
  const subtotal = newOrder.value.items.reduce((sum, item) => {
    return sum + (item.qty * item.unit_price * 100)
  }, 0)
  
  const discount = subtotal * (newOrder.value.discount_percent / 100)
  const afterDiscount = subtotal - discount
  const tax = afterDiscount * (newOrder.value.tax_percent / 100)
  
  return afterDiscount + tax
}

const saveOrder = async () => {
  try {
    loading.value = true

    // Prepare items for API
    const items = newOrder.value.items.map(item => ({
      product_id: item.product_id ? parseInt(item.product_id) : null,
      name_snapshot: item.name_snapshot,
      sku_snapshot: item.sku_snapshot || null,
      qty: item.qty,
      unit_price_cents: Math.round(item.unit_price * 100),
      currency: item.currency
    }))

    if (isEditing.value && editingOrderId.value) {
      // Use UpdateOrder: (id, status, notes, discountPercent, taxPercent, items)
      await UpdateOrder(editingOrderId.value, '', newOrder.value.notes, newOrder.value.discount_percent, newOrder.value.tax_percent, items)
      isEditing.value = false
      editingOrderId.value = null
      showCreateModal.value = false
      showDetailModal.value = false
    } else {
      await CreateOrder(
        parseInt(newOrder.value.client_id),
        newOrder.value.notes,
        newOrder.value.discount_percent,
        newOrder.value.tax_percent,
        items
      )
      closeCreateModal()
    }

    await fetchOrders()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to save order'
  } finally {
    loading.value = false
  }
}

const viewOrder = async (order: any) => {
  try {
    console.log('Viewing order:', order.order.id)
    const orderDetail = await GetOrder(order.order.id)
    detailOrder.value = orderDetail
    showDetailModal.value = true
    isEditing.value = false
  } catch (err) {
    console.error('Error fetching order details:', err)
    error.value = err instanceof Error ? err.message : 'Failed to load order details'
  }
}

const editOrder = async (order: any) => {
  try {
    const orderDetail = await GetOrder(order.order.id)
    // Populate edit form with data
    isEditing.value = true
    editingOrderId.value = order.order.id
    newOrder.value.client_id = orderDetail.order.client_id.toString()
    newOrder.value.notes = orderDetail.order.notes || ''
    newOrder.value.discount_percent = orderDetail.order.discount_percent
    newOrder.value.tax_percent = orderDetail.order.tax_percent
    newOrder.value.items = orderDetail.items.map((it: any) => ({
      product_id: it.product_id ? it.product_id.toString() : '',
      name_snapshot: it.name_snapshot,
      sku_snapshot: it.sku_snapshot || '',
      qty: it.qty,
      unit_price: it.unit_price_cents / 100,
      currency: it.currency || 'USD',
      productSearch: it.name_snapshot
    }))

    detailOrder.value = orderDetail
    showDetailModal.value = true
    showCreateModal.value = true
  } catch (err) {
    console.error('Error preparing edit:', err)
    error.value = err instanceof Error ? err.message : 'Failed to prepare edit'
  }
}

const beginEdit = (orderDetailLocal: any) => {
  isEditing.value = true
  showCreateModal.value = true
  showDetailModal.value = false
}

const cancelEdit = () => {
  isEditing.value = false
  editingOrderId.value = null
  showCreateModal.value = false
  showDetailModal.value = false
  // Reset newOrder if needed
  newOrder.value = {
    client_id: '',
    notes: '',
    discount_percent: 0,
    tax_percent: 0,
    items: []
  }
}

const exportOrderPDF = async (order: any) => {
  try {
    console.log('Exporting PDF for order:', order.order.id)
    const pdfBytes = await ExportOrderPDF(order.order.id)
  // Convert number array to Uint8Array (use from for safety)
  const uint8Array = Uint8Array.from(pdfBytes as any)
    
    // Create blob and download
    const blob = new Blob([uint8Array], { type: 'application/pdf' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `order-${order.order.order_number}.pdf`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    console.log('PDF exported successfully')
  } catch (err) {
    console.error('PDF export error:', err)
    error.value = err instanceof Error ? err.message : 'Failed to export PDF'
  }
}

// Lifecycle
onMounted(() => {
  fetchOrders()
})
</script>
