// API client for Mynt NAS
const API_BASE = '/api/v1';

interface User {
    id: number;
    username: string;
    full_name?: string;
    email?: string;
    account_type: 'system' | 'virtual';
    is_admin: boolean;
    is_active: boolean;
    home_dir?: string;
    shell?: string;
    uid?: number;
    gid?: number;
    created_at: string;
    last_login?: string;
}

interface Pool {
    name: string;
    guid?: string;
    health: string;
    size: number;
    allocated: number;
    free: number;
    frag?: number;
    disk_count?: number;
    redundancy?: number;
    vdevs?: VDevDetail[];
    scrub_status?: ScrubStatus;
    resilver_status?: ResilverStatus;
}

interface ScrubStatus {
    in_progress: boolean;
    end_time?: string;
    errors: number;
    data_scanned: number;
    data_to_scan: number;
    scan_rate: number;
}

interface VDevDetail {
    name: string;
    type: string;
    status: string;
    children: DiskDetail[];
}

interface DiskDetail {
    name: string;
    path: string;
    status: string;
    slot?: string;
    read: number;
    write: number;
    checksum: number;
    replacing: boolean;
}

interface ResilverStatus {
    in_progress: boolean;
    percent_done: number;
    start_time: number;    // Unix timestamp for frontend to calculate remaining time
    scanned_bytes: number;
    issued_bytes: number;  // bytes processed (for rate calculation)
    total_bytes: number;
    rate: number;
}

interface PoolHealth {
    status: string;
    can_lose_more: number;
    risk_level: string;
    risk_description: string;
    recommendation: string;
}

interface StorageSpace {
    name: string;
    type: 'filesystem' | 'volume';
    pool: string;
    used: number;
    available: number;
    referenced: number;
    quota?: number;
    reservation?: number;
    mountpoint?: string;
    compression?: string;
}

interface Snapshot {
    name: string;
    dataset: string;
    created_at: string;
    used: number;
    referenced: number;
    source: string; // "manual", "policy:daily", etc.
}

interface SnapshotPolicy {
    id: number;
    name: string;
    schedule: string;
    retention: string;
    datasets: string[];
    enabled: boolean;
    created_at: string;
    updated_at: string;
}

interface CreateDatasetRequest {
    name: string;
    type?: string;
    use_case?: string;
    quota_mode?: string;
    quota?: number;  // size/quota in bytes (required for volumes, optional for filesystems)
    properties?: Record<string, string>;
}

interface UsageInfo {
    type: string;
    params?: Record<string, string>;
}

interface Disk {
    name: string;
    path: string;
    model?: string;
    serial: string;
    size: number;
    type: string;
    in_use: boolean;
    usage?: UsageInfo;
    slot?: string;
    pool?: string;
    status: string;          // "healthy", "warning", "failed", "unknown"
    smart_health: string;    // "good", "warning", "failed", "unknown"
    temperature?: number;
}

interface SmartAttribute {
    id: number;
    name: string;
    value: number;
    worst: number;
    thresh: number;
    raw: string;
    status: string;
}

interface DetailedSmartReport {
    disk: string;
    passed: boolean;
    attributes: SmartAttribute[];
    power_on_hours: number;
    power_cycle_count: number;
    reallocated_sectors: number;
    pending_sectors: number;
    uncorrectable_errors: number;
    temperature: number;
    checked_at: string;
}

interface SmartTestStatus {
    running: boolean;
    type?: string;
    progress?: number;
    remaining_mins?: number;
    last_result?: string;
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

export interface SystemStats {
    cpu: {
        total_usage: number;
        cores: number[];
        model: string;
    };
    memory: {
        total: number;
        available: number;
        used: number;
        used_percent: number;
    };
    swap: {
        total: number;
        used: number;
        free: number;
        used_percent: number;
    };
    network?: {
        bytes_sent: number;
        bytes_recv: number;
        packets_sent: number;
        packets_recv: number;
    };
}

export interface ProcessInfo {
    pid: number;
    name: string;
    username: string;
    status: string;
    cpu_percent: number;
    memory_percent: number;
    create_time: number;
    cmdline: string;
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

        // Check Content-Type to determine if response is JSON
        const contentType = response.headers.get('Content-Type');
        if (contentType?.includes('application/json')) {
            return response.json();
        }

        return undefined as T;
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

    async getDiskSmartDetails(name: string): Promise<DetailedSmartReport> {
        return this.request(`/disks/${encodeURIComponent(name)}/smart`);
    }

    async refreshSmartData(name: string): Promise<DetailedSmartReport> {
        return this.request(`/disks/${encodeURIComponent(name)}/smart/refresh`, {
            method: 'POST',
        });
    }

    async runSmartTest(name: string, type: 'short' | 'long'): Promise<void> {
        return this.request(`/disks/${encodeURIComponent(name)}/smart/test`, {
            method: 'POST',
            body: JSON.stringify({ type }),
        });
    }

    async getSmartTestStatus(name: string): Promise<SmartTestStatus> {
        return this.request(`/disks/${encodeURIComponent(name)}/smart/test/status`);
    }

    async locateDisk(name: string, action: 'on' | 'off'): Promise<void> {
        return this.request(`/disks/${encodeURIComponent(name)}/locate`, {
            method: 'POST',
            body: JSON.stringify({ action }),
        });
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

    // Datasets
    async listDatasets(): Promise<StorageSpace[]> {
        return this.request('/datasets');
    }

    async createDataset(req: CreateDatasetRequest): Promise<void> {
        return this.request('/datasets', {
            method: 'POST',
            body: JSON.stringify(req),
        });
    }

    async deleteDataset(name: string): Promise<void> {
        return this.request(`/datasets/${encodeURIComponent(name)}`, {
            method: 'DELETE',
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

    // Users
    async listUsers(): Promise<User[]> {
        return this.request('/users');
    }

    async createUser(data: {
        username: string;
        password: string;
        full_name?: string;
        email?: string;
        account_type?: 'system' | 'virtual';
        is_admin?: boolean;
    }): Promise<User> {
        return this.request('/users', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async deleteUser(username: string): Promise<void> {
        return this.request(`/users/${username}`, {
            method: 'DELETE',
        });
    }

    // Snapshots
    async listSnapshots(dataset: string): Promise<Snapshot[]> {
        return this.request(`/snapshots?dataset=${encodeURIComponent(dataset)}`);
    }

    async createSnapshot(dataset: string, name: string): Promise<Snapshot> {
        return this.request('/snapshots', {
            method: 'POST',
            body: JSON.stringify({ dataset, name }),
        });
    }

    async deleteSnapshot(snapshotName: string): Promise<void> {
        return this.request(`/snapshots/${encodeURIComponent(snapshotName)}`, {
            method: 'DELETE',
        });
    }

    async rollbackSnapshot(snapshotName: string): Promise<void> {
        return this.request(`/snapshots/rollback?name=${encodeURIComponent(snapshotName)}`, {
            method: 'POST',
        });
    }

    // Snapshot Policies
    async listSnapshotPolicies(): Promise<SnapshotPolicy[]> {
        return this.request('/snapshot-policies');
    }

    async createSnapshotPolicy(policy: Partial<SnapshotPolicy>): Promise<SnapshotPolicy> {
        return this.request('/snapshot-policies', {
            method: 'POST',
            body: JSON.stringify(policy),
        });
    }

    async updateSnapshotPolicy(id: number, policy: Partial<SnapshotPolicy>): Promise<SnapshotPolicy> {
        return this.request(`/snapshot-policies/${id}`, {
            method: 'PUT',
            body: JSON.stringify(policy),
        });
    }

    async deleteSnapshotPolicy(id: number): Promise<void> {
        return this.request(`/snapshot-policies/${id}`, {
            method: 'DELETE',
        });
    }

    // Dataset quota management
    async setDatasetQuota(datasetName: string, quota: number): Promise<void> {
        return this.request(`/datasets/quota?name=${encodeURIComponent(datasetName)}`, {
            method: 'PUT',
            body: JSON.stringify({ quota }),
        });
    }

    // Pool management
    async scrubPool(poolName: string): Promise<void> {
        return this.request(`/pools/${poolName}/scrub`, {
            method: 'POST',
            body: JSON.stringify({ action: 'start' }),
        });
    }

    // Pool detail operations
    async getPool(poolName: string): Promise<Pool> {
        return this.request(`/pools/${poolName}`);
    }

    async replaceDisk(poolName: string, oldDisk: string, newDisk: string): Promise<void> {
        return this.request(`/pools/${poolName}/replace`, {
            method: 'POST',
            body: JSON.stringify({ old_disk: oldDisk, new_disk: newDisk }),
        });
    }

    // System Monitoring
    async getSystemStats(): Promise<SystemStats> {
        return this.request<SystemStats>("/system/stats");
    }

    async getProcesses(): Promise<ProcessInfo[]> {
        return this.request<ProcessInfo[]>("/system/processes");
    }

    async killProcess(pid: number): Promise<void> {
        return this.request<void>(`/system/processes/${pid}`, {
            method: "DELETE",
        });
    }
}

export const api = new ApiClient();
export type { User, Pool, VDevDetail, DiskDetail, ResilverStatus, ScrubStatus, PoolHealth, Disk, Share, Notification, Snapshot, StorageSpace, CreateDatasetRequest, SnapshotPolicy, SmartAttribute, DetailedSmartReport, SmartTestStatus };

