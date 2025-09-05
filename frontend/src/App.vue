<template>
  <div id="app" class="min-h-screen bg-gray-50 rtl" dir="rtl">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4 space-x-reverse">
            <h1 class="text-xl font-bold text-gray-900">
              {{ $t('app.title') }}
            </h1>
          </div>
          <div class="flex items-center space-x-4 space-x-reverse">
            <!-- Language Toggle -->
            <button
              @click="toggleLanguage"
              class="p-2 text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-md"
            >
              <span class="text-sm font-medium">
                {{ currentLocale === 'ar' ? 'EN' : 'ع' }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </header>

    <div class="flex">
      <!-- Right Sidebar Navigation -->
      <nav class="w-64 bg-white shadow-sm border-l border-gray-200 min-h-[92vh]">
        <div class="py-6">
          <div class="px-3 space-y-1">
            <router-link
              v-for="item in navigation"
              :key="item.name"
              :to="item.href"
              :class="[
                $route.path === item.href
                  ? 'bg-blue-50 border-blue-500 text-blue-700 border-l-4'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
                'group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-colors duration-150'
              ]"
            >
              <component
                :is="item.icon"
                :class="[
                  $route.path === item.href
                    ? 'text-blue-500'
                    : 'text-gray-400 group-hover:text-gray-500',
                  'ml-3 h-6 w-6 transition-colors duration-150'
                ]"
              />
              {{ $t(item.name) }}
            </router-link>
          </div>
        </div>
      </nav>

      <!-- Main Content -->
      <main class="flex-1 overflow-auto">
        <div class="py-6">
          <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <router-view />
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  HomeIcon,
  UsersIcon,
  CubeIcon,
  DocumentIcon,
  DocumentTextIcon,
  CreditCardIcon,
  CogIcon
} from '@heroicons/vue/24/outline'

const { locale } = useI18n()

const currentLocale = computed(() => locale.value)

const navigation = [
  { name: 'nav.dashboard', href: '/dashboard', icon: HomeIcon },
  { name: 'nav.clients', href: '/clients', icon: UsersIcon },
  { name: 'nav.products', href: '/products', icon: CubeIcon },
  { name: 'nav.orders', href: '/orders', icon: DocumentIcon },
  { name: 'nav.invoices', href: '/invoices', icon: DocumentTextIcon },
  { name: 'nav.payments', href: '/payments', icon: CreditCardIcon },
  { name: 'nav.settings', href: '/settings', icon: CogIcon },
]

function toggleLanguage() {
  locale.value = locale.value === 'ar' ? 'en' : 'ar'
  
  // Update document direction
  const html = document.documentElement
  if (locale.value === 'ar') {
    html.setAttribute('dir', 'rtl')
    html.classList.add('rtl')
  } else {
    html.setAttribute('dir', 'ltr')
    html.classList.remove('rtl')
  }
}
</script>

<style scoped>
/* RTL adjustments for router-link active state */
.rtl .border-l-4 {
  border-left-width: 0;
  border-right-width: 4px;
}

/* RTL spacing adjustments */
.rtl .space-x-reverse > :not([hidden]) ~ :not([hidden]) {
  --tw-space-x-reverse: 1;
  margin-right: calc(1rem * var(--tw-space-x-reverse));
  margin-left: calc(1rem * calc(1 - var(--tw-space-x-reverse)));
}
</style>
