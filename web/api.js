// API Client for Mynt NAS
class API {
    constructor(baseURL = '/api/v1') {
        this.baseURL = baseURL;
        this.token = localStorage.getItem('token');
    }

    async request(method, path, body = null) {
        const headers = { 'Content-Type': 'application/json' };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const options = { method, headers };
        if (body) {
            options.body = JSON.stringify(body);
        }

        const response = await fetch(this.baseURL + path, options);

        if (response.status === 401) {
            // Unauthorized - redirect to login
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            if (window.location.pathname !== '/login.html') {
                window.location = '/login.html';
            }
            throw new Error('Unauthorized');
        }

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || response.statusText);
        }

        if (response.status === 204) {
            return null; // No content
        }

        return response.json();
    }

    // Auth
    async checkSetup() {
        return this.request('GET', '/setup/status');
    }

    async setup(data) {
        const result = await this.request('POST', '/setup', data);
        if (result.token) {
            this.token = result.token;
            localStorage.setItem('token', result.token);
            localStorage.setItem('user', JSON.stringify(result.user));
        }
        return result;
    }

    async login(username, password) {
        const result = await this.request('POST', '/auth/login', { username, password });
        if (result.token) {
            this.token = result.token;
            localStorage.setItem('token', result.token);
            localStorage.setItem('user', JSON.stringify(result.user));
        }
        return result;
    }

    logout() {
        this.token = null;
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location = '/login.html';
    }

    getCurrentUser() {
        const user = localStorage.getItem('user');
        return user ? JSON.parse(user) : null;
    }

    // Disks
    async listDisks() {
        return this.request('GET', '/disks');
    }

    async getDiskSmart(name) {
        return this.request('GET', `/disks/smart?name=${encodeURIComponent(name)}`);
    }

    // Pools
    async listPools() {
        return this.request('GET', '/pools');
    }

    async createPool(data) {
        return this.request('POST', '/pools', data);
    }

    // Datasets
    async listDatasets() {
        return this.request('GET', '/datasets');
    }

    async createDataset(data) {
        return this.request('POST', '/datasets', data);
    }

    async deleteDataset(name) {
        return this.request('DELETE', `/datasets/${encodeURIComponent(name)}`);
    }

    // Shares
    async listShares() {
        return this.request('GET', '/shares');
    }

    async createShare(data) {
        return this.request('POST', '/shares', data);
    }

    async deleteShare(id) {
        return this.request('DELETE', `/shares/${id}`);
    }

    // Users
    async listUsers() {
        return this.request('GET', '/users');
    }

    async createUser(data) {
        return this.request('POST', '/users', data);
    }

    async deleteUser(username) {
        return this.request('DELETE', `/users/${encodeURIComponent(username)}`);
    }

    // Notifications
    async listNotifications(status = '', limit = 50, offset = 0) {
        let url = `/notifications?limit=${limit}&offset=${offset}`;
        if (status) url += `&status=${encodeURIComponent(status)}`;
        return this.request('GET', url);
    }

    async getNotificationCount() {
        return this.request('GET', '/notifications/count');
    }

    async markNotificationRead(id) {
        return this.request('POST', `/notifications/${id}/read`);
    }

    async markNotificationAck(id) {
        return this.request('POST', `/notifications/${id}/ack`);
    }

    async deleteNotification(id) {
        return this.request('DELETE', `/notifications/${id}`);
    }
}

// Global API instance
const api = new API();
