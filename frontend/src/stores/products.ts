import { defineStore } from 'pinia'
import { CreateProduct, GetProducts, UpdateProduct, DeleteProduct } from '../../wailsjs/go/main/App'
import { db } from '../../wailsjs/go/models'

// Use the generated Product type
export type Product = db.Product

export interface ProductSearchResult {
  data: Product[]
  total: number
}

export const useProductStore = defineStore('products', {
  state: () => ({
    products: [] as Product[],
    loading: false,
    error: null as string | null
  }),

  getters: {
    totalProducts: (state) => state.products.length,
    
    activeProducts: (state) => {
      return state.products.filter(product => product.active)
    },
    
    inactiveProducts: (state) => {
      return state.products.filter(product => !product.active)
    }
  },

  actions: {
    async fetchProducts(search = '', limit = 20, offset = 0): Promise<ProductSearchResult> {
      try {
        this.loading = true
        this.error = null
        
        const result = await GetProducts(search, limit, offset)
        
        // Update local state with all products for getters
        if (offset === 0) {
          this.products = result.data || []
        }
        
        return {
          data: result.data || [],
          total: result.total || 0
        }
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch products'
        throw error
      } finally {
        this.loading = false
      }
    },

    async createProduct(product: { name: string; description?: string; price: number; sku?: string }): Promise<Product> {
      try {
        this.loading = true
        this.error = null
        
        const newProduct = await CreateProduct(
          product.name,
          product.description || '',
          product.price * 100,
          product.sku || ''
        )
        
        this.products.push(newProduct)
        return newProduct
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to create product'
        throw error
      } finally {
        this.loading = false
      }
    },

    async updateProduct(product: { id: number; name: string; description?: string; price: number; sku?: string }): Promise<Product> {
      try {
        this.loading = true
        this.error = null
        
        const updatedProduct = await UpdateProduct(
          product.id,
          product.name,
          product.description || '',
          product.price * 100,
          product.sku || ''
        )
        
        const index = this.products.findIndex(p => p.id === product.id)
        if (index !== -1) {
          this.products[index] = updatedProduct
        }
        
        return updatedProduct
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to update product'
        throw error
      } finally {
        this.loading = false
      }
    },

    async deleteProduct(id: number): Promise<void> {
      try {
        this.loading = true
        this.error = null
        await DeleteProduct(id)
        this.products = this.products.filter(p => p.id !== id)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to delete product'
        throw error
      } finally {
        this.loading = false
      }
    },

    clearError() {
      this.error = null
    }
  }
})

export { useProductStore as default }
