// Wails Runtime Types
declare global {
  interface Window {
    go: {
      main: {
        App: {
          // Client operations
          CreateClient(name: string, phone: string, email: string, address: string): Promise<any>
          GetClients(query: string, limit: number, offset: number): Promise<any>
          GetClient(id: number): Promise<any>
          UpdateClient(id: number, name: string, phone: string, email: string, address: string): Promise<any>
          
          // Product operations
          CreateProduct(name: string, description: string, price: number, sku: string): Promise<any>
          GetProducts(query: string, limit: number, offset: number): Promise<any>
          GetProduct(id: number): Promise<any>
          UpdateProduct(id: number, name: string, description: string, price: number, sku: string): Promise<any>
          
          // Dashboard operations
          GetDashboardMetrics(timeRange: string): Promise<any>
          
          // General
          Greet(name: string): Promise<string>
        }
      }
    }
  }
}

export {}
