import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { userService } from '../services/user';

function Avatar({ name }) {
    const initials = name
        ? name.split(' ').map((n) => n[0]).slice(0, 2).join('').toUpperCase()
        : '?';
    return (
        <div className="w-10 h-10 rounded-full bg-indigo-100 text-indigo-700 font-bold text-sm flex items-center justify-center flex-shrink-0">
            {initials}
        </div>
    );
}

export default function Friends() {
    const { user } = useAuthStore();
    const [friends, setFriends] = useState([]);
    const [loading, setLoading] = useState(true);
    const [email, setEmail] = useState('');
    const [submitting, setSubmitting] = useState(false);
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
        setSubmitting(true);

        try {
            if (!user?.user_id) {
                throw new Error('User ID is missing. Please re-login.');
            }
            const response = await userService.addFriend(user.user_id, email);
            setMessage(response.message || 'Friend added successfully!');
            setEmail('');
            loadFriends();
        } catch (err) {
            setError(err.message || err.response?.data?.error || 'Failed to add friend.');
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className="min-h-screen bg-slate-50">
            {/* Top Navigation */}
            <header className="bg-white border-b border-slate-200 sticky top-0 z-20">
                <div className="max-w-3xl mx-auto px-4 sm:px-6 h-16 flex items-center gap-4">
                    <Link
                        to="/dashboard"
                        className="flex items-center gap-1.5 text-slate-500 hover:text-slate-800 transition-colors text-sm font-medium"
                    >
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                        </svg>
                        Dashboard
                    </Link>
                    <span className="text-slate-300">/</span>
                    <div className="flex items-center gap-2">
                        <div className="w-6 h-6 bg-indigo-600 rounded-md flex items-center justify-center">
                            <svg viewBox="0 0 24 24" className="w-3.5 h-3.5 fill-none stroke-white stroke-2" strokeLinecap="round" strokeLinejoin="round">
                                <rect x="5" y="2" width="14" height="20" rx="2" />
                                <line x1="9" y1="7" x2="15" y2="7" />
                                <line x1="9" y1="11" x2="15" y2="11" />
                                <line x1="9" y1="15" x2="12" y2="15" />
                            </svg>
                        </div>
                        <span className="text-sm font-bold text-slate-900">SplitEase</span>
                    </div>
                </div>
            </header>

            <main className="max-w-3xl mx-auto px-4 sm:px-6 py-8">
                <div className="mb-6">
                    <h1 className="text-2xl font-bold text-slate-900">Friends</h1>
                    <p className="text-slate-500 text-sm mt-0.5">Manage the people you split expenses with.</p>
                </div>

                {/* Add friend card */}
                <div className="card mb-6">
                    <h2 className="text-base font-bold text-slate-900 mb-4">Add a Friend</h2>
                    <form onSubmit={handleAddFriend}>
                        <div className="flex gap-3">
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder="Enter friend's email address"
                                className="input flex-1"
                                required
                            />
                            <button
                                type="submit"
                                disabled={submitting}
                                className="btn btn-primary px-5 flex-shrink-0"
                            >
                                {submitting ? <span className="spinner" /> : 'Add'}
                            </button>
                        </div>

                        {message && (
                            <div className="flex items-center gap-2 mt-3 text-sm text-emerald-700 bg-emerald-50 border border-emerald-200 px-3 py-2 rounded-lg">
                                <svg className="w-4 h-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                                </svg>
                                {message}
                            </div>
                        )}
                        {error && (
                            <div className="flex items-center gap-2 mt-3 text-sm text-red-700 bg-red-50 border border-red-200 px-3 py-2 rounded-lg">
                                <svg className="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                                </svg>
                                {error}
                            </div>
                        )}
                    </form>
                </div>

                {/* Friends list */}
                <div className="card">
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-base font-bold text-slate-900">Your Friends</h2>
                        {friends.length > 0 && (
                            <span className="text-xs font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded-full">
                                {friends.length} {friends.length === 1 ? 'friend' : 'friends'}
                            </span>
                        )}
                    </div>

                    {loading ? (
                        <div className="flex items-center justify-center py-10">
                            <span className="spinner-dark" />
                        </div>
                    ) : friends.length === 0 ? (
                        <div className="text-center py-12">
                            <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                <svg className="w-7 h-7 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                            </div>
                            <p className="font-semibold text-slate-700 mb-1">No friends yet</p>
                            <p className="text-sm text-slate-400">Add a friend above to start splitting expenses together.</p>
                        </div>
                    ) : (
                        <div className="divide-y divide-slate-100">
                            {friends.map((friend) => (
                                <div key={friend.id} className="flex items-center gap-4 py-3 first:pt-0 last:pb-0">
                                    <Avatar name={friend.name} />
                                    <div className="flex-1 min-w-0">
                                        <p className="font-semibold text-slate-900 text-sm">{friend.name}</p>
                                        <p className="text-xs text-slate-500 truncate">{friend.email}</p>
                                    </div>
                                    <span className="px-2.5 py-1 bg-emerald-50 text-emerald-700 rounded-full text-xs font-medium border border-emerald-100">
                                        Connected
                                    </span>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
}
