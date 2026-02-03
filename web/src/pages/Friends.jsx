import { useState, useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import { userService } from '../services/user';

export default function Friends() {
    const { user } = useAuthStore();
    const [friends, setFriends] = useState([]);
    const [loading, setLoading] = useState(true);
    const [email, setEmail] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');

    useEffect(() => {
        if (user?.user_id) {
            loadFriends();
        }
    }, [user]);

    const loadFriends = async () => {
        try {
            const data = await userService.getFriends(user.user_id);
            setFriends(data);
        } catch (err) {
            console.error('Failed to load friends', err);
        } finally {
            setLoading(false);
        }
    };

    const handleAddFriend = async (e) => {
        e.preventDefault();
        setMessage('');
        setError('');

        try {
            console.log('Submitting friend request:', { userId: user?.user_id, email });
            if (!user?.user_id) {
                throw new Error("User ID is missing. Please re-login.");
            }
            const response = await userService.addFriend(user.user_id, email);
            setMessage(response.message || 'Friend request sent!');
            setEmail('');
            loadFriends(); // Refresh list
        } catch (err) {
            console.error('Friends page error:', err);
            setError(err.message || err.response?.data?.error || 'Failed to add friend');
        }
    };

    if (loading) return <div className="p-4 text-center">Loading friends...</div>;

    return (
        <div className="card">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">My Friends</h2>

            {/* Add Friend Form */}
            <form onSubmit={handleAddFriend} className="mb-8 p-4 bg-gray-50 rounded-lg">
                <h3 className="text-lg font-semibold mb-3">Add Friend</h3>
                <div className="flex gap-2">
                    <input
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="Enter friend's email"
                        className="input flex-1"
                        required
                    />
                    <button type="submit" className="btn btn-primary">
                        Add / Invite
                    </button>
                </div>
                {message && <p className="text-green-600 mt-2 text-sm">{message}</p>}
                {error && <p className="text-red-600 mt-2 text-sm">{error}</p>}
            </form>

            {/* Friends List */}
            <div className="space-y-3">
                {friends.length === 0 ? (
                    <p className="text-gray-500 text-center py-4">No friends yet. Invite someone!</p>
                ) : (
                    friends.map((friend) => (
                        <div
                            key={friend.id}
                            className="p-4 bg-white border border-gray-100 rounded-lg shadow-sm flex justify-between items-center"
                        >
                            <div>
                                <h3 className="font-semibold text-gray-900">{friend.name}</h3>
                                <p className="text-sm text-gray-600">{friend.email}</p>
                            </div>
                            <span className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-xs">
                                Friend
                            </span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
