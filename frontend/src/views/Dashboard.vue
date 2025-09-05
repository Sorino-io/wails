<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 mb-2">
        {{ $t('dashboard.title') }}
      </h1>
      <p class="text-gray-600">
        {{ $t('app.subtitle') }}
      </p>
    </div>

    <!-- Metrics Cards -->
    <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
      <div v-for="metric in metrics" :key="metric.name" class="card">
        <div class="card-body">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <component
                :is="metric.icon"
                class="h-8 w-8 text-blue-600"
              />
            </div>
            <div class="mr-5 w-0 flex-1">
              <dt class="text-sm font-medium text-gray-500 truncate">
                {{ $t(metric.name) }}
              </dt>
              <dd class="text-lg font-medium text-gray-900">
                {{ metric.value }}
              </dd>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
      <!-- Revenue Chart -->
      <div class="card">
        <div class="card-header">
          <h3 class="text-lg font-medium text-gray-900">
            {{ $t('dashboard.charts.revenue_by_month') }}
          </h3>
        </div>
        <div class="card-body">
          <div class="h-64 flex items-center justify-center text-gray-500">
            <ChartBarIcon class="h-16 w-16 mb-4" />
            <p>{{ $t('messages.loading') }}</p>
          </div>
        </div>
      </div>

      <!-- Top Clients -->
      <div class="card">
        <div class="card-header">
          <h3 class="text-lg font-medium text-gray-900">
            {{ $t('dashboard.charts.top_clients') }}
          </h3>
        </div>
        <div class="card-body">
          <div class="space-y-4">
            <div v-for="client in topClients" :key="client.id" class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-900">{{ client.name }}</p>
                <p class="text-sm text-gray-500">{{ client.order_count }} {{ $t('nav.orders') }}</p>
              </div>
              <div class="text-left">
                <p class="text-sm font-medium text-gray-900">
                  {{ formatCurrency(client.total_paid_cents) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Orders -->
      <div class="card">
        <div class="card-header">
          <h3 class="text-lg font-medium text-gray-900">
            {{ $t('orders.title') }} - الأحدث
          </h3>
        </div>
        <div class="card-body">
          <div class="text-center text-gray-500 py-8">
            <DocumentIcon class="h-12 w-12 mx-auto mb-4" />
            <p>{{ $t('messages.loading') }}</p>
          </div>
        </div>
      </div>

      <!-- Recent Invoices -->
      <div class="card">
        <div class="card-header">
          <h3 class="text-lg font-medium text-gray-900">
            {{ $t('invoices.title') }} - الأحدث
          </h3>
        </div>
        <div class="card-body">
          <div class="text-center text-gray-500 py-8">
            <DocumentTextIcon class="h-12 w-12 mx-auto mb-4" />
            <p>{{ $t('messages.loading') }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetDashboardMetrics } from '../../wailsjs/go/main/App'
import { db } from '../../wailsjs/go/models'
import {
  DocumentIcon,
  DocumentTextIcon,
  CreditCardIcon,
  ExclamationTriangleIcon,
  ChartBarIcon
} from '@heroicons/vue/24/outline'

const { t } = useI18n()

const dashboardData = ref<db.DashboardData | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const metrics = ref([
  {
    name: 'dashboard.metrics.orders_this_month',
    icon: DocumentIcon,
    value: '---'
  },
  {
    name: 'dashboard.metrics.invoices_this_month', 
    icon: DocumentTextIcon,
    value: '---'
  },
  {
    name: 'dashboard.metrics.payments_collected',
    icon: CreditCardIcon,
    value: '---'
  },
  {
    name: 'dashboard.metrics.outstanding_invoices',
    icon: ExclamationTriangleIcon,
    value: '---'
  }
])

const topClients = ref<db.TopClient[]>([])

async function loadDashboardData() {
  try {
    loading.value = true
    error.value = null
    
    const data = await GetDashboardMetrics('month')
    dashboardData.value = data
    
    // Update metrics
    metrics.value[0].value = data.total_orders_month.toString()
    metrics.value[1].value = data.total_invoices_month.toString()
    metrics.value[2].value = formatCurrency(data.payments_collected_month_cents)
    metrics.value[3].value = data.outstanding_invoices_count.toString()
    
    // Update top clients
    topClients.value = data.top_clients || []
    
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'حدث خطأ في تحميل البيانات'
    console.error('Failed to load dashboard data:', err)
  } finally {
    loading.value = false
  }
}

function formatCurrency(cents: number): string {
  const amount = (cents / 100).toFixed(2)
  return `${amount} د.ج`
}

onMounted(() => {
  loadDashboardData()
})
</script>
