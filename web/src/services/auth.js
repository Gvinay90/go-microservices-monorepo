import api from './api';

export const authService = {
    // Register new user
    register: async (name, email, password) => {
        const response = await api.post('/auth/register', { name, email, password });
        return response.data;
    },

    // Login user - handles Keycloak token response
    login: async (email, password) => {
        const response = await api.post('/auth/login', { email, password });
        const { access_token, refresh_token, expires_in } = response.data;

        if (access_token) {
            // Store tokens
            localStorage.setItem('access_token', access_token);
            if (refresh_token) {
                localStorage.setItem('refresh_token', refresh_token);
            }
            // Store expiration time (current time + expires_in seconds)
            const expiresAt = Date.now() + (expires_in * 1000);
            localStorage.setItem('token_expires_at', expiresAt.toString());
        }

        return response.data;
    },

    // Get current user
    getCurrentUser: async () => {
        const response = await api.get('/auth/me');
        return response.data;
    },

    // Logout
    logout: () => {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('token_expires_at');
        window.location.href = '/login';
    },

    // Check if user is authenticated
    isAuthenticated: () => {
        const token = localStorage.getItem('access_token');
        const expiresAt = localStorage.getItem('token_expires_at');

        if (!token) return false;

        // Check if token is expired
        if (expiresAt && Date.now() > parseInt(expiresAt)) {
            return false;
        }

        return true;
    },

    // Get access token
    getAccessToken: () => {
        return localStorage.getItem('access_token');
    },

    // Get refresh token
    getRefreshToken: () => {
        return localStorage.getItem('refresh_token');
    },
};
