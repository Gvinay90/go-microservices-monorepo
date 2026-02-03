import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { expenseService } from '../services/expenses';
import { userService } from '../services/user';
import CreateExpenseForm from '../components/CreateExpenseForm';

export default function Dashboard() {
    const { user, logout } = useAuthStore();
    const [expenses, setExpenses] = useState([]);
    const [friends, setFriends] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showCreateForm, setShowCreateForm] = useState(false);

    useEffect(() => {
        if (user?.user_id) {
            loadData();
        }
    }, [user]);

    const loadData = async () => {
        try {
            // Fetch expenses (now filtering by user in backend)
            // Fetch friends (filtered by user)
            const [expensesData, friendsData] = await Promise.all([
                expenseService.getExpenses(),
                userService.getFriends(user.user_id),
            ]);
            setExpenses(expensesData || []);
            setFriends(friendsData || []);
        } catch (error) {
            console.error('Failed to load data:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateExpense = async (expenseData) => {
        try {
            await expenseService.createExpense(expenseData);
            setShowCreateForm(false);
            loadData();
        } catch (error) {
            alert('Failed to create expense');
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="spinner"></div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 p-4">
            {/* Header */}
            <div className="max-w-6xl mx-auto mb-8">
                <div className="glass rounded-2xl p-6 flex justify-between items-center text-white">
                    <div>
                        <h1 className="text-3xl font-bold mb-1">Expense Sharing</h1>
                        <p className="opacity-90">Welcome, {user?.name || user?.email}!</p>
                    </div>
                    <div className="flex gap-3">
                        <Link to="/friends" className="btn bg-white/20 text-white hover:bg-white/30">
                            Manage Friends
                        </Link>
                        <button onClick={logout} className="btn bg-white/10 text-white hover:bg-white/20 border-white/20">
                            Logout
                        </button>
                    </div>
                </div>
            </div>

            <div className="max-w-6xl mx-auto grid lg:grid-cols-3 gap-6">
                {/* Main Content (Expenses) */}
                <div className="lg:col-span-2">
                    <div className="card">
                        <div className="flex justify-between items-center mb-6">
                            <h2 className="text-2xl font-bold text-gray-900">Recent Expenses</h2>
                            <button
                                onClick={() => setShowCreateForm(!showCreateForm)}
                                className="btn btn-primary"
                            >
                                + Add Expense
                            </button>
                        </div>

                        {showCreateForm && (
                            <CreateExpenseForm
                                friends={friends}
                                currentUser={user}
                                onSubmit={handleCreateExpense}
                                onCancel={() => setShowCreateForm(false)}
                            />
                        )}

                        <div className="space-y-4">
                            {expenses.length === 0 ? (
                                <div className="text-center py-12 bg-gray-50 rounded-lg border border-dashed border-gray-200">
                                    <p className="text-gray-500 text-lg mb-2">No expenses yet</p>
                                    <p className="text-sm text-gray-400">Add friends and start splitting bills!</p>
                                </div>
                            ) : (
                                expenses.map((expense) => {
                                    const isPayer = expense.paid_by === user.user_id;
                                    const mySplit = expense.splits?.find(s => s.user_id === user.user_id);

                                    return (
                                        <div
                                            key={expense.id}
                                            className="p-5 bg-white rounded-xl border border-gray-100 shadow-sm hover:shadow-md transition-all flex justify-between items-center group"
                                        >
                                            <div>
                                                <div className="flex items-baseline gap-2 mb-1">
                                                    <h3 className="text-lg font-semibold text-gray-900 group-hover:text-indigo-600 transition-colors">
                                                        {expense.description}
                                                    </h3>
                                                    <span className="text-xs text-gray-400">
                                                        {new Date(expense.created_at * 1000).toLocaleDateString()}
                                                    </span>
                                                </div>
                                                <div className="text-sm text-gray-600">
                                                    {isPayer ? (
                                                        <span className="text-green-600 font-medium">You paid total</span>
                                                    ) : (
                                                        <span>
                                                            <span className="font-medium text-gray-900">Someone</span> paid
                                                        </span>
                                                    )}
                                                </div>
                                            </div>

                                            <div className="text-right">
                                                <div className="font-bold text-xl text-gray-900">
                                                    ${expense.total_amount?.toFixed(2)}
                                                </div>
                                                <div className="text-sm mt-1">
                                                    {isPayer ? (
                                                        <span className="text-green-600 font-medium">
                                                            you lent ${(expense.total_amount - (mySplit?.amount || 0)).toFixed(2)}
                                                        </span>
                                                    ) : (
                                                        <span className="text-red-500 font-medium">
                                                            you owe ${(mySplit?.amount || 0).toFixed(2)}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                    );
                                })
                            )}
                        </div>
                    </div>
                </div>

                {/* Sidebar (Balances & Friends) */}
                <div className="space-y-6">
                    {/* Friends Quick View */}
                    <div className="card">
                        <div className="flex justify-between items-center mb-4">
                            <h2 className="text-lg font-bold text-gray-900">My Friends</h2>
                            <Link to="/friends" className="text-indigo-600 text-sm hover:underline">View All</Link>
                        </div>
                        <div className="space-y-3">
                            {friends.length === 0 ? (
                                <p className="text-sm text-gray-500 italic">No friends added.</p>
                            ) : (
                                friends.slice(0, 5).map((f) => (
                                    <div key={f.id} className="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg transition-colors">
                                        <div className="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 font-bold text-xs">
                                            {f.name?.[0]?.toUpperCase() || '?'}
                                        </div>
                                        <div className="overflow-hidden">
                                            <p className="font-medium text-gray-900 truncate">{f.name}</p>
                                            <p className="text-xs text-gray-500 truncate">{f.email}</p>
                                        </div>
                                    </div>
                                ))
                            )}
                            <Link to="/friends" className="block w-full py-2 text-center text-sm text-gray-500 border border-dashed border-gray-200 rounded-lg hover:bg-gray-50 hover:border-gray-300 transition-all">
                                + Add Friend
                            </Link>
                        </div>
                    </div>

                    {/* Balance Summary (Placeholder - Future Feature) */}
                    <div className="card bg-gradient-to-br from-gray-900 to-gray-800 text-white">
                        <h2 className="text-lg font-bold mb-4">Your Balance</h2>
                        <div className="text-3xl font-bold mb-2">
                            {/* Calculated balance would go here */}
                            $0.00
                        </div>
                        <p className="text-sm text-gray-400">Total owed to you</p>
                    </div>
                </div>
            </div>
        </div>
    );
}
