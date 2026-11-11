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
  discount_paise?: number;
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
  total_products: number;
  total_categories: number;
}

export interface VendorItem {
  id: string;
  owner_user_id: string;
  business_name: string;
  description?: string;
  phone: string;
  status: 'PENDING' | 'ACTIVE' | 'SUSPENDED' | 'CLOSED';
  availability_status: 'OPEN' | 'CLOSED' | 'TEMPORARILY_UNAVAILABLE';
  address_line1?: string;
  city?: string;
  product_count: number;
  created_at: string;
}

export interface AdminReview {
  id: string;
  customer_id: string;
  vendor_id: string;
  order_id: string;
  rating: number;
  comment?: string;
  is_hidden: boolean;
  created_at: string;
  customer_name?: string;
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

export interface DeliveryAssignment {
  id: string;
  order_id: string;
  agent_user_id: string;
  status: 'ASSIGNED' | 'ACCEPTED' | 'PICKED_UP' | 'OUT_FOR_DELIVERY' | 'DELIVERED';
  assigned_by?: string;
  assigned_at: string;
  accepted_at?: string;
  picked_up_at?: string;
  delivered_at?: string;
  updated_at: string;
  customer_name?: string;
  customer_phone?: string;
  delivery_address_line1?: string;
  delivery_city?: string;
  delivery_lat?: number;
  delivery_lng?: number;
  order_total_paise?: number;
  order_status?: string;
  vendor_name?: string;
  vendor_phone?: string;
  vendor_lat?: number;
  vendor_lng?: number;
}

export interface DeliveryLocation {
  order_id: string;
  agent_user_id: string;
  latitude: number;
  longitude: number;
  updated_at: string;
}

export interface RouteData {
  routing_available: boolean;
  provider: string;
  distance_type?: 'road' | 'straight_line_estimate' | 'none';
  approximate_straight_line_km?: number;
  distance_meters: number;
  duration_seconds: number;
  geometry?: {
    type: string;
    coordinates: [number, number][]; // [longitude, latitude]
  };
  distance_km: number;
  duration_minutes: number;
  attribution: string;
  directions_url: string;
  message?: string;
}

export interface MapConfigData {
  provider: string;
  style_url: string;
  tile_url: string;
  routing_provider: string;
  osrm_base_url: string;
  geocoding_provider: string;
  geocoding_base_url: string;
  attribution: string;
}

// Keyless, MapLibre-compatible fallback. Kept in one place so no component
// hardcodes a tile provider URL.
export const DEFAULT_MAP_STYLE_URL = 'https://tiles.openfreemap.org/styles/liberty';

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

  static async assignUserRole(userId: string, roleCode: string) {
    return this.request(`/users/${userId}/roles`, {
      method: 'POST',
      body: JSON.stringify({ role_code: roleCode }),
    });
  }

  static async revokeUserRole(userId: string, roleCode: string) {
    return this.request(`/users/${userId}/roles`, {
      method: 'DELETE',
      body: JSON.stringify({ role_code: roleCode }),
    });
  }

  // Vendors
  static async listVendors(): Promise<VendorItem[]> {
    return this.request<VendorItem[]>('/vendors');
  }

  static async updateVendorStatus(vendorId: string, status: 'PENDING' | 'ACTIVE' | 'SUSPENDED' | 'CLOSED') {
    return this.request(`/vendors/${vendorId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  }

  // Reviews Moderation
  static async listReviews(): Promise<AdminReview[]> {
    return this.request<AdminReview[]>('/reviews/admin');
  }

  static async moderateReview(reviewId: string, isHidden: boolean) {
    return this.request(`/reviews/${reviewId}/moderate`, {
      method: 'PATCH',
      body: JSON.stringify({ is_hidden: isHidden }),
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

  // Delivery Management
  static async listDeliveryAssignments(): Promise<DeliveryAssignment[]> {
    const res = await this.request<{ data: DeliveryAssignment[] }>('/delivery/admin/assignments');
    return res.data || [];
  }

  static async assignDeliveryAgent(orderId: string, agentUserId: string): Promise<DeliveryAssignment> {
    const res = await this.request<{ data: DeliveryAssignment }>(`/delivery/orders/${orderId}/assign`, {
      method: 'POST',
      body: JSON.stringify({ agent_user_id: agentUserId }),
    });
    return res.data;
  }

  static async getDeliveryLocation(orderId: string): Promise<DeliveryLocation> {
    const res = await this.request<{ data: DeliveryLocation }>(`/delivery/orders/${orderId}/location`);
    return res.data;
  }

  static async getRoute(originLat: number, originLng: number, destLat: number, destLng: number): Promise<RouteData> {
    const res = await this.request<{ data: RouteData }>(`/mobility/route?origin_lat=${originLat}&origin_lng=${originLng}&dest_lat=${destLat}&dest_lng=${destLng}`);
    return res.data;
  }

  static async getMapConfig(): Promise<MapConfigData> {
    return this.request<MapConfigData>('/mobility/config');
  }
}
