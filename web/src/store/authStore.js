import { create } from 'zustand';
import { authService } from '../services/auth';

export const useAuthStore = create((set) => ({
    user: null,
    isAuthenticated: false,
    isCheckingAuth: true, // Start as true
    isLoading: false,
    error: null,

    login: async (email, password) => {
        set({ isLoading: true, error: null });
        try {
            await authService.login(email, password);
            const userData = await authService.getCurrentUser();
            set({ user: userData, isAuthenticated: true, isLoading: false });
            return true;
        } catch (error) {
            set({ error: error.response?.data?.error || 'Login failed', isLoading: false });
            return false;
        }
    },

    register: async (name, email, password) => {
        set({ isLoading: true, error: null });
        try {
            await authService.register(name, email, password);
            set({ isLoading: false });
            return true;
        } catch (error) {
            set({ error: error.response?.data?.error || 'Registration failed', isLoading: false });
            return false;
        }
    },

    logout: () => {
        authService.logout();
        set({ user: null, isAuthenticated: false });
    },

    checkAuth: async () => {
        set({ isCheckingAuth: true });
        if (authService.isAuthenticated()) {
            try {
                const userData = await authService.getCurrentUser();
                set({ user: userData, isAuthenticated: true, isCheckingAuth: false });
            } catch (error) {
                set({ user: null, isAuthenticated: false, isCheckingAuth: false });
            }
        } else {
            set({ user: null, isAuthenticated: false, isCheckingAuth: false });
        }
    },

    clearError: () => set({ error: null }),
}));
