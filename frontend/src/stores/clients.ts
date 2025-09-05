import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { CreateClient, GetClients, GetClient, UpdateClient } from '../../wailsjs/go/main/App'

export interface Client {
  id?: number
  name: string
  phone?: string
  email?: string
  address?: string
  created_at?: string
  updated_at?: string
}

export const useClientStore = defineStore('clients', () => {
  const clients = ref<Client[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const getClients = computed(() => clients.value)
  const isLoading = computed(() => loading.value)
  const getError = computed(() => error.value)

  async function fetchClients(query = '', limit = 20, offset = 0) {
    loading.value = true
    error.value = null
    
    try {
      // Call Wails backend method using generated bindings
      const result = await GetClients(query, limit, offset)
      clients.value = result.data || []
      return {
        data: result.data || [],
        total: result.total || 0
      }
    } catch (err) {
      console.error('Error fetching clients:', err)
      error.value = err instanceof Error ? err.message : 'حدث خطأ غير متوقع'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createClient(client: Client) {
    loading.value = true
    error.value = null
    
    try {
      console.log('Creating client:', client)
      const result = await CreateClient(
        client.name,
        client.phone || '',
        client.email || '',
        client.address || ''
      )
      console.log('Client created:', result)
      clients.value.push(result)
      return result
    } catch (err) {
      console.error('Error creating client:', err)
      error.value = err instanceof Error ? err.message : 'فشل في إنشاء العميل'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateClient(client: Client) {
    loading.value = true
    error.value = null
    
    try {
      const result = await UpdateClient(
        client.id!,
        client.name,
        client.phone || '',
        client.email || '',
        client.address || ''
      )
      const index = clients.value.findIndex(c => c.id === client.id)
      if (index !== -1) {
        clients.value[index] = result
      }
      return result
    } catch (err) {
      console.error('Error updating client:', err)
      error.value = err instanceof Error ? err.message : 'فشل في تحديث العميل'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getClient(id: number) {
    loading.value = true
    error.value = null
    
    try {
      const result = await GetClient(id)
      return result
    } catch (err) {
      console.error('Error getting client:', err)
      error.value = err instanceof Error ? err.message : 'العميل غير موجود'
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    clients,
    loading,
    error,
    getClients,
    isLoading,
    getError,
    fetchClients,
    createClient,
    updateClient,
    getClient,
    clearError
  }
})
