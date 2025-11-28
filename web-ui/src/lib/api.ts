// API client for Mynt NAS
const API_BASE = '/api/v1';

interface User {
    id: number;
    username: string;
    full_name?: string;
    email?: string;
    is_admin: boolean;
}

interface Pool {
    name: string;
    health: string;
    size: number;
    allocated: number;
    free: number;
    capacity: number;
}

interface Disk {
    name: string;
    path: string;
    model?: string;
    serial: string;
    size: number;
    type: string;
}

interface Share {
    id: number;
    name: string;
    path: string;
    protocol: string;
    read_only: boolean;
    browseable: boolean;
    guest_ok: boolean;
    valid_users: string;
    comment: string;
    share_type: 'normal' | 'public' | 'restricted';
}

interface Notification {
    id: number;
    type: string;
    data: string;
    status: string;
    created_at: string;
    read_at?: string;
    acked_at?: string;
}

class ApiClient {
    private token: string | null = null;

    constructor() {
        if (typeof window !== 'undefined') {
            this.token = localStorage.getItem('auth_token');
        }
    }

    private async request<T>(
        endpoint: string,
        options: RequestInit = {}
    ): Promise<T> {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
            ...(options.headers as Record<string, string>),
        };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const response = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || response.statusText);
        }

        return response.json();
    }

    // Auth
    async login(username: string, password: string): Promise<{ token: string; user: User }> {
        const data = await this.request<{ token: string; user: User }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });

        this.token = data.token;
        if (typeof window !== 'undefined') {
            localStorage.setItem('auth_token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
        }
        return data;
    }

    async setup(username: string, password: string, full_name?: string, email?: string) {
        const data = await this.request<{ token: string; user: User }>('/setup', {
            method: 'POST',
            body: JSON.stringify({ username, password, full_name, email }),
        });

        this.token = data.token;
        if (typeof window !== 'undefined') {
            localStorage.setItem('auth_token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
        }
        return data;
    }

    logout() {
        this.token = null;
        if (typeof window !== 'undefined') {
            localStorage.removeItem('auth_token');
            localStorage.removeItem('user');
        }
    }

    getCurrentUser(): User | null {
        if (typeof window !== 'undefined') {
            const userStr = localStorage.getItem('user');
            return userStr ? JSON.parse(userStr) : null;
        }
        return null;
    }

    isAuthenticated(): boolean {
        return !!this.token;
    }

    // Disks
    async listDisks(): Promise<Disk[]> {
        return this.request('/disks');
    }

    // Pools
    async listPools(): Promise<Pool[]> {
        return this.request('/pools');
    }

    async createPool(name: string, devices: string[], type: string) {
        return this.request('/pools', {
            method: 'POST',
            body: JSON.stringify({ name, devices, type }),
        });
    }

    // Shares
    async listShares(): Promise<Share[]> {
        return this.request('/shares');
    }

    async createShare(share: Partial<Share>) {
        return this.request('/shares', {
            method: 'POST',
            body: JSON.stringify(share),
        });
    }

    async deleteShare(id: number) {
        return this.request(`/shares/${id}`, {
            method: 'DELETE',
        });
    }

    // Notifications
    async listNotifications(status = '', limit = 20, offset = 0): Promise<Notification[]> {
        const params = new URLSearchParams({
            limit: limit.toString(),
            offset: offset.toString(),
        });
        if (status) params.append('status', status);

        return this.request(`/notifications?${params}`);
    }

    async getNotificationCount(): Promise<{ unread: number; total: number }> {
        return this.request('/notifications/count');
    }

    async markNotificationRead(id: number) {
        return this.request(`/notifications/${id}/read`, {
            method: 'POST',
        });
    }
}

export const api = new ApiClient();
export type { User, Pool, Disk, Share, Notification };
