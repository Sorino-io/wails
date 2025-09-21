import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './views/Dashboard.vue'
import Clients from './views/Clients.vue'
import Products from './views/Products.vue'
import Orders from './views/Orders.vue'
import Invoices from './views/Invoices.vue'
import Payments from './views/Payments.vue'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/clients',
    name: 'Clients',
    component: Clients
  },
  {
    path: '/products',
    name: 'Products',
    component: Products
  },
  {
    path: '/orders',
    name: 'Orders',
    component: Orders
  },
  {
    path: '/payments',
    name: 'Payments',
    component: Payments
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
