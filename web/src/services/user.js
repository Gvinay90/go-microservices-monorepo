import api from './api';

export const userService = {
    // Get current user's friends
    getFriends: async (userId) => {
        const response = await api.get(`/users/${userId}/friends`);
        return response.data.friends || [];
    },

    // Add a friend by email (or neighbor id)
    addFriend: async (userId, email) => {
        console.log('UserService.addFriend called with:', { userId, email });
        if (!userId) console.error('UserService: userId is missing!');
        try {
            const response = await api.post(`/users/${userId}/friends`, { email });
            console.log('UserService.addFriend response:', response);
            return response.data;
        } catch (error) {
            console.error('UserService.addFriend error:', error);
            throw error;
        }
    },

    // List all users (deprecated but kept for now)
    getAllUsers: async () => {
        const response = await api.get('/users');
        return response.data.users;
    }
};
