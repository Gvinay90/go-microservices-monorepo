import api from './api';

export const expenseService = {
    // Get all expenses
    getExpenses: async () => {
        const response = await api.get('/expenses');
        return response.data.expenses || [];
    },

    // Get expense by ID
    getExpense: async (id) => {
        const response = await api.get(`/expenses/${id}`);
        return response.data.expense;
    },

    // Create new expense
    createExpense: async (expenseData) => {
        const response = await api.post('/expenses', expenseData);
        return response.data.expense;
    },

    // Update expense
    updateExpense: async (id, expenseData) => {
        const response = await api.put(`/expenses/${id}`, expenseData);
        return response.data.expense;
    },

    // Delete expense
    deleteExpense: async (id) => {
        const response = await api.delete(`/expenses/${id}`);
        return response.data;
    },

    // Settle balance (from authenticated user -> toUserId)
    settleBalance: async (toUserId, amount) => {
        const response = await api.post('/expenses/settle', {
            to_user_id: toUserId,
            amount,
        });
        return response.data;
    },

    // Get authenticated user's net balance
    getNetBalance: async () => {
        const response = await api.get('/expenses/net-balance');
        return response.data.net_balance;
    },
};

export const userService = {
    // Get all users
    getUsers: async () => {
        const response = await api.get('/users');
        return response.data.users || [];
    },

    // Get user by ID
    getUser: async (id) => {
        const response = await api.get(`/users/${id}`);
        return response.data.user;
    },

    // Get user's friends
    getFriends: async (id) => {
        const response = await api.get(`/users/${id}/friends`);
        return response.data.friends || [];
    },

    // Add friend
    addFriend: async (userId, friendId) => {
        const response = await api.post(`/users/${userId}/friends`, { friend_id: friendId });
        return response.data;
    },
};
