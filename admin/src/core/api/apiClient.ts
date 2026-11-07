// Typed API Client for VendorSaathi Admin Dashboard

export interface User {
  id: string;
  phone: string;
  name?: string;
  status: 'ACTIVE' | 'SUSPENDED' | 'CLOSED';
  roles: string[];
  created_at: string;
}

export interface VendorProfile {
  id: string;
  owner_user_id: string;
  business_name: string;
  description?: string;
  phone: string;
  status: 'PENDING' | 'ACTIVE' | 'SUSPENDED' | 'CLOSED';
  availability_status: 'OPEN' | 'CLOSED' | 'TEMPORARILY_UNAVAILABLE';
  created_at: string;
}

export interface OrderItem {
  id: string;
  product_name_snapshot: string;
  unit_price_paise: number;
  quantity: number;
  line_total_paise: number;
}

export interface Order {
  id: string;
  customer_id: string;
  vendor_id: string;
  fulfilment_type: 'PICKUP' | 'DELIVERY';
  status: 'PENDING' | 'ACCEPTED' | 'PREPARING' | 'READY_FOR_PICKUP' | 'OUT_FOR_DELIVERY' | 'COMPLETED' | 'REJECTED' | 'CANCELLED';
  subtotal_paise: number;
  total_paise: number;
  items: OrderItem[];
  created_at: string;
}

export interface PlatformOverview {
  total_users: number;
  total_vendors: number;
  active_vendors: number;
  total_customers: number;
  total_orders: number;
  completed_orders: number;
  total_gross_sales_paise: number;
  total_collections_paise: number;
}

export interface AuditLog {
  id: string;
  actor_user_id?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  reason?: string;
  created_at: string;
}

const BASE_URL = '/api/v1';

export class ApiClient {
  private static token: string | null = localStorage.getItem('access_token');

  static setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('access_token', token);
    } else {
      localStorage.removeItem('access_token');
    }
  }

  static getToken(): string | null {
    return this.token;
  }

  private static async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const res = await fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    if (res.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new Event('auth:unauthorized'));
      throw new Error('Session expired. Please log in again.');
    }

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({}));
      const msg = errBody?.error?.message || errBody?.message || `Request failed (${res.status})`;
      throw new Error(msg);
    }

    if (res.status === 204) {
      return {} as T;
    }

    return res.json();
  }

  // Auth
  static async login(phone: string, password: string) {
    const data = await this.request<{ access_token: string; refresh_token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ phone, password }),
    });
    this.setToken(data.access_token);
    return data;
  }

  static async me() {
    return this.request<{ user: User }>('/me');
  }

  static async logout() {
    try {
      await this.request('/auth/logout', { method: 'POST' });
    } finally {
      this.setToken(null);
    }
  }

  // Dashboard & Analytics
  static async getOverview(): Promise<PlatformOverview> {
    return this.request<PlatformOverview>('/analytics/platform/overview');
  }

  // Users
  static async listUsers(limit = 20, cursor = '') {
    return this.request<{ data: User[]; next_cursor: string; has_more: boolean }>(`/users?limit=${limit}&cursor=${cursor}`);
  }

  static async updateUserStatus(userId: string, status: 'ACTIVE' | 'SUSPENDED' | 'CLOSED') {
    return this.request(`/users/${userId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  }

  // Orders
  static async listOrders(limit = 20, cursor = '') {
    return this.request<{ orders: Order[]; next_cursor: string; has_more: boolean }>(`/orders?limit=${limit}&cursor=${cursor}`);
  }

  // Audit Logs
  static async listAuditLogs(limit = 20, cursor = '') {
    return this.request<{ data: AuditLog[]; next_cursor: string; has_more: boolean }>(`/audit-logs?limit=${limit}&cursor=${cursor}`);
  }
}
