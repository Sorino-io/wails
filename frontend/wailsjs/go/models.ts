export namespace db {
	
	export class Client {
	    id: number;
	    name: string;
	    phone?: string;
	    email?: string;
	    address?: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Client(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	        this.address = source["address"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TopClient {
	    id: number;
	    name: string;
	    order_count: number;
	    total_paid_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new TopClient(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.order_count = source["order_count"];
	        this.total_paid_cents = source["total_paid_cents"];
	    }
	}
	export class RevenueByMonth {
	    month: string;
	    revenue_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new RevenueByMonth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.month = source["month"];
	        this.revenue_cents = source["revenue_cents"];
	    }
	}
	export class DashboardData {
	    total_orders_month: number;
	    total_invoices_month: number;
	    payments_collected_month_cents: number;
	    outstanding_invoices_count: number;
	    revenue_by_month: RevenueByMonth[];
	    top_clients: TopClient[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_orders_month = source["total_orders_month"];
	        this.total_invoices_month = source["total_invoices_month"];
	        this.payments_collected_month_cents = source["payments_collected_month_cents"];
	        this.outstanding_invoices_count = source["outstanding_invoices_count"];
	        this.revenue_by_month = this.convertValues(source["revenue_by_month"], RevenueByMonth);
	        this.top_clients = this.convertValues(source["top_clients"], TopClient);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Invoice {
	    id: number;
	    invoice_number: string;
	    order_id?: number;
	    client_id: number;
	    status: string;
	    // Go type: time
	    issue_date: any;
	    // Go type: time
	    due_date?: any;
	    notes?: string;
	    subtotal_cents: number;
	    discount_percent: number;
	    tax_percent: number;
	    total_cents: number;
	    currency: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Invoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoice_number = source["invoice_number"];
	        this.order_id = source["order_id"];
	        this.client_id = source["client_id"];
	        this.status = source["status"];
	        this.issue_date = this.convertValues(source["issue_date"], null);
	        this.due_date = this.convertValues(source["due_date"], null);
	        this.notes = source["notes"];
	        this.subtotal_cents = source["subtotal_cents"];
	        this.discount_percent = source["discount_percent"];
	        this.tax_percent = source["tax_percent"];
	        this.total_cents = source["total_cents"];
	        this.currency = source["currency"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Payment {
	    id: number;
	    invoice_id: number;
	    amount_cents: number;
	    method: string;
	    reference?: string;
	    // Go type: time
	    paid_at: any;
	    notes?: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Payment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoice_id = source["invoice_id"];
	        this.amount_cents = source["amount_cents"];
	        this.method = source["method"];
	        this.reference = source["reference"];
	        this.paid_at = this.convertValues(source["paid_at"], null);
	        this.notes = source["notes"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InvoiceItem {
	    id: number;
	    invoice_id: number;
	    product_id?: number;
	    name_snapshot: string;
	    sku_snapshot?: string;
	    qty: number;
	    unit_price_cents: number;
	    currency: string;
	    total_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new InvoiceItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoice_id = source["invoice_id"];
	        this.product_id = source["product_id"];
	        this.name_snapshot = source["name_snapshot"];
	        this.sku_snapshot = source["sku_snapshot"];
	        this.qty = source["qty"];
	        this.unit_price_cents = source["unit_price_cents"];
	        this.currency = source["currency"];
	        this.total_cents = source["total_cents"];
	    }
	}
	export class InvoiceDetail {
	    invoice: Invoice;
	    client: Client;
	    items: InvoiceItem[];
	    payments: Payment[];
	    paid_cents: number;
	    balance_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new InvoiceDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoice = this.convertValues(source["invoice"], Invoice);
	        this.client = this.convertValues(source["client"], Client);
	        this.items = this.convertValues(source["items"], InvoiceItem);
	        this.payments = this.convertValues(source["payments"], Payment);
	        this.paid_cents = source["paid_cents"];
	        this.balance_cents = source["balance_cents"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Order {
	    id: number;
	    order_number: string;
	    client_id: number;
	    status: string;
	    notes?: string;
	    discount_percent: number;
	    tax_percent: number;
	    // Go type: time
	    issue_date: any;
	    // Go type: time
	    due_date?: any;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Order(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.order_number = source["order_number"];
	        this.client_id = source["client_id"];
	        this.status = source["status"];
	        this.notes = source["notes"];
	        this.discount_percent = source["discount_percent"];
	        this.tax_percent = source["tax_percent"];
	        this.issue_date = this.convertValues(source["issue_date"], null);
	        this.due_date = this.convertValues(source["due_date"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OrderItem {
	    id: number;
	    order_id: number;
	    product_id?: number;
	    name_snapshot: string;
	    sku_snapshot?: string;
	    qty: number;
	    unit_price_cents: number;
	    currency: string;
	    total_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new OrderItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.order_id = source["order_id"];
	        this.product_id = source["product_id"];
	        this.name_snapshot = source["name_snapshot"];
	        this.sku_snapshot = source["sku_snapshot"];
	        this.qty = source["qty"];
	        this.unit_price_cents = source["unit_price_cents"];
	        this.currency = source["currency"];
	        this.total_cents = source["total_cents"];
	    }
	}
	export class OrderDetail {
	    order: Order;
	    client: Client;
	    items: OrderItem[];
	    subtotal_cents: number;
	    discount_cents: number;
	    tax_cents: number;
	    total_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new OrderDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.order = this.convertValues(source["order"], Order);
	        this.client = this.convertValues(source["client"], Client);
	        this.items = this.convertValues(source["items"], OrderItem);
	        this.subtotal_cents = source["subtotal_cents"];
	        this.discount_cents = source["discount_cents"];
	        this.tax_cents = source["tax_cents"];
	        this.total_cents = source["total_cents"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class PaginatedResult_myproject_backend_db_Client_ {
	    data: Client[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResult_myproject_backend_db_Client_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], Client);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaginatedResult_myproject_backend_db_InvoiceDetail_ {
	    data: InvoiceDetail[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResult_myproject_backend_db_InvoiceDetail_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], InvoiceDetail);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaginatedResult_myproject_backend_db_OrderDetail_ {
	    data: OrderDetail[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResult_myproject_backend_db_OrderDetail_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], OrderDetail);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Product {
	    id: number;
	    sku?: string;
	    name: string;
	    description?: string;
	    unit_price_cents: number;
	    currency: string;
	    active: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Product(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sku = source["sku"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.unit_price_cents = source["unit_price_cents"];
	        this.currency = source["currency"];
	        this.active = source["active"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaginatedResult_myproject_backend_db_Product_ {
	    data: Product[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResult_myproject_backend_db_Product_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], Product);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	

}

