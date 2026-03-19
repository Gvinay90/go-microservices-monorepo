import { useState, useEffect, useMemo, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { expenseService } from '../services/expenses';
import { userService } from '../services/user';
import CreateExpenseForm from '../components/CreateExpenseForm';

function Avatar({ name, size = 'md' }) {
    const initials = name
        ? name.split(' ').map((n) => n[0]).slice(0, 2).join('').toUpperCase()
        : '?';
    const sz = size === 'lg' ? 'w-10 h-10 text-sm' : size === 'sm' ? 'w-7 h-7 text-xs' : 'w-8 h-8 text-xs';
    return (
        <div className={`${sz} rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center flex-shrink-0`}>
            {initials}
        </div>
    );
}

function StatCard({ label, value, color = 'default', icon }) {
    const colorMap = {
        default: 'bg-white',
        green: 'bg-emerald-50 border-emerald-100',
        red: 'bg-red-50 border-red-100',
        indigo: 'bg-indigo-50 border-indigo-100',
    };
    const textMap = {
        default: 'text-slate-900',
        green: 'text-emerald-700',
        red: 'text-red-600',
        indigo: 'text-indigo-700',
    };
    return (
        <div className={`card ${colorMap[color]} flex items-center gap-4 py-4`}>
            <div className="text-2xl">{icon}</div>
            <div>
                <p className="text-xs text-slate-500 font-medium uppercase tracking-wide">{label}</p>
                <p className={`text-xl font-bold mt-0.5 ${textMap[color]}`}>{value}</p>
            </div>
        </div>
    );
}

export default function Dashboard() {
    const { user, logout } = useAuthStore();
    const [expenses, setExpenses] = useState([]);
    const [friends, setFriends] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showCreateForm, setShowCreateForm] = useState(false);
    const [createError, setCreateError] = useState('');
    const [netBalance, setNetBalance] = useState(null);

    const [editExpenseId, setEditExpenseId] = useState(null);
    const [editDescription, setEditDescription] = useState('');
    const [editTotalAmount, setEditTotalAmount] = useState('');
    const [editError, setEditError] = useState('');

    const [deletingExpenseId, setDeletingExpenseId] = useState(null);

    const [settleToUserId, setSettleToUserId] = useState('');
    const [settleAmount, setSettleAmount] = useState('');
    const [settleError, setSettleError] = useState('');
    const [settleSuccess, setSettleSuccess] = useState('');
    const [settling, setSettling] = useState(false);

    const loadData = useCallback(async () => {
        try {
            // Load friends/expenses first. Net balance is optional; if it fails
            // we still want the user list + expense form to render.
            const [expensesData, friendsData] = await Promise.all([
                expenseService.getExpenses(),
                userService.getFriends(user.user_id),
            ]);

            setExpenses(expensesData || []);
            setFriends(friendsData || []);

            try {
                const netBalanceData = await expenseService.getNetBalance();
                setNetBalance(typeof netBalanceData === 'number' ? netBalanceData : null);
            } catch (err) {
                console.warn('Failed to load net balance:', err);
                setNetBalance(null);
            }
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
        } finally {
            setLoading(false);
        }
    }, [user?.user_id]);

    useEffect(() => {
        if (user?.user_id) {
            loadData();
        }
    }, [user?.user_id, loadData]);

    const handleCreateExpense = async (expenseData) => {
        setCreateError('');
        try {
            await expenseService.createExpense(expenseData);
            setShowCreateForm(false);
            loadData();
        } catch {
            setCreateError('Failed to create expense. Please try again.');
        }
    };

    const stats = useMemo(() => {
        let totalLent = 0;
        let totalOwed = 0;
        expenses.forEach((expense) => {
            const mySplit = expense.splits?.find((s) => s.user_id === user?.user_id);
            if (expense.paid_by === user?.user_id) {
                totalLent += expense.total_amount - (mySplit?.amount || 0);
            } else {
                totalOwed += mySplit?.amount || 0;
            }
        });
        const derivedNet = totalLent - totalOwed;
        return { totalLent, totalOwed, net: netBalance ?? derivedNet };
    }, [expenses, user, netBalance]);

    if (loading) {
        return (
            <div className="min-h-screen bg-slate-50 flex items-center justify-center">
                <div className="text-center">
                    <div className="spinner-dark spinner-lg mx-auto mb-3" />
                    <p className="text-slate-500 text-sm">Loading your dashboard…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-slate-50">
            {/* Top Navigation */}
            <header className="bg-white border-b border-slate-200 sticky top-0 z-20">
                <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center">
                            <svg viewBox="0 0 24 24" className="w-4 h-4 fill-none stroke-white stroke-2" strokeLinecap="round" strokeLinejoin="round">
                                <rect x="5" y="2" width="14" height="20" rx="2" />
                                <line x1="9" y1="7" x2="15" y2="7" />
                                <line x1="9" y1="11" x2="15" y2="11" />
                                <line x1="9" y1="15" x2="12" y2="15" />
                            </svg>
                        </div>
                        <span className="text-base font-bold text-slate-900 tracking-tight">SplitEase</span>
                    </div>

                    <div className="flex items-center gap-3">
                        <Link
                            to="/friends"
                            className="btn btn-secondary text-xs px-3 py-1.5 gap-1.5"
                        >
                            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                            </svg>
                            Friends
                        </Link>

                        <div className="flex items-center gap-2 pl-3 border-l border-slate-200">
                            <Avatar name={user?.name} size="sm" />
                            <span className="text-sm font-medium text-slate-700 hidden sm:block">
                                {user?.name || user?.email}
                            </span>
                        </div>

                        <button
                            onClick={logout}
                            className="btn btn-ghost text-xs px-3 py-1.5 text-slate-500"
                            title="Sign out"
                        >
                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                            </svg>
                        </button>
                    </div>
                </div>
            </header>

            <main className="max-w-6xl mx-auto px-4 sm:px-6 py-8">
                {/* Page heading */}
                <div className="mb-6">
                    <h1 className="text-2xl font-bold text-slate-900">
                        Hello, {user?.name?.split(' ')[0] || 'there'} 👋
                    </h1>
                    <p className="text-slate-500 text-sm mt-0.5">Here&apos;s a summary of your shared expenses.</p>
                </div>

                {/* Stats row */}
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
                    <StatCard
                        label="Total Expenses"
                        value={expenses.length}
                        icon={
                            <svg className="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                            </svg>
                        }
                    />
                    <StatCard
                        label="You Lent"
                        value={`$${stats.totalLent.toFixed(2)}`}
                        color="green"
                        icon={
                            <svg className="w-5 h-5 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M5 10l7-7m0 0l7 7m-7-7v18" />
                            </svg>
                        }
                    />
                    <StatCard
                        label="You Owe"
                        value={`$${stats.totalOwed.toFixed(2)}`}
                        color="red"
                        icon={
                            <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
                            </svg>
                        }
                    />
                    <StatCard
                        label="Net Balance"
                        value={`${stats.net >= 0 ? '+' : ''}$${stats.net.toFixed(2)}`}
                        color={stats.net >= 0 ? 'indigo' : 'red'}
                        icon={
                            <svg className="w-5 h-5 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
                            </svg>
                        }
                    />
                </div>

                {/* Main two-column layout */}
                <div className="grid lg:grid-cols-3 gap-6">
                    {/* Expenses column */}
                    <div className="lg:col-span-2">
                        <div className="card">
                            <div className="flex items-center justify-between mb-5">
                                <h2 className="text-lg font-bold text-slate-900">Recent Expenses</h2>
                                <button
                                    onClick={() => {
                                        setCreateError('');
                                        setShowCreateForm(!showCreateForm);
                                    }}
                                    className="btn btn-primary text-sm px-4 py-2 gap-1.5"
                                >
                                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                                    </svg>
                                    Add Expense
                                </button>
                            </div>

                            {showCreateForm && (
                                <div className="mb-5">
                                    {createError && (
                                        <div className="flex items-center gap-2 bg-red-50 border border-red-200 text-red-700 px-3 py-2.5 rounded-lg mb-3 text-sm">
                                            <svg className="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                                            </svg>
                                            {createError}
                                        </div>
                                    )}
                                    <CreateExpenseForm
                                        friends={friends}
                                        currentUser={user}
                                        onSubmit={handleCreateExpense}
                                        onCancel={() => setShowCreateForm(false)}
                                    />
                                </div>
                            )}

                            <div className="space-y-2">
                                {expenses.length === 0 ? (
                                    <div className="text-center py-14">
                                        <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
                                            <svg className="w-7 h-7 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                                            </svg>
                                        </div>
                                        <p className="font-semibold text-slate-700 mb-1">No expenses yet</p>
                                        <p className="text-sm text-slate-400">Add friends and start splitting bills!</p>
                                    </div>
                                ) : (
                                    expenses.map((expense) => {
                                        const isPayer = expense.paid_by === user.user_id;
                                        const mySplit = expense.splits?.find((s) => s.user_id === user.user_id);
                                        const netAmount = isPayer
                                            ? expense.total_amount - (mySplit?.amount || 0)
                                            : mySplit?.amount || 0;

                                        return (
                                            <div
                                                key={expense.id}
                                                className="flex items-center gap-4 p-4 rounded-xl border border-slate-100 hover:border-slate-200 hover:bg-slate-50 transition-all group"
                                            >
                                                <div className="w-10 h-10 rounded-xl bg-indigo-50 flex items-center justify-center flex-shrink-0">
                                                    <svg className="w-5 h-5 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                                        <path strokeLinecap="round" strokeLinejoin="round" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2z" />
                                                    </svg>
                                                </div>

                                                <div className="flex-1 min-w-0">
                                                    <p className="font-semibold text-slate-900 text-sm truncate group-hover:text-indigo-600 transition-colors">
                                                        {expense.description}
                                                    </p>
                                                    <p className="text-xs text-slate-400 mt-0.5">
                                                        {isPayer ? 'You paid' : 'Someone paid'} •{' '}
                                                        {new Date(expense.created_at * 1000).toLocaleDateString('en-US', {
                                                            month: 'short',
                                                            day: 'numeric',
                                                            year: 'numeric',
                                                        })}
                                                    </p>
                                                </div>

                                                <div className="text-right flex-shrink-0 min-w-[210px]">
                                                    <p className="font-bold text-slate-900 text-sm">
                                                        ${expense.total_amount?.toFixed(2)}
                                                    </p>
                                                    <p className={`text-xs font-semibold mt-0.5 ${isPayer ? 'text-emerald-600' : 'text-red-500'}`}>
                                                        {isPayer
                                                            ? `+$${netAmount.toFixed(2)} lent`
                                                            : `-$${netAmount.toFixed(2)} owed`}
                                                    </p>

                                                    <div className="mt-2 flex justify-end gap-2">
                                                        <button
                                                            type="button"
                                                            className="btn btn-ghost text-xs px-3 py-1.5"
                                                            onClick={() => {
                                                                setEditError('');
                                                                setEditExpenseId(expense.id);
                                                                setEditDescription(expense.description || '');
                                                                setEditTotalAmount(String(expense.total_amount ?? ''));
                                                            }}
                                                        >
                                                            Edit
                                                        </button>
                                                        <button
                                                            type="button"
                                                            className="btn btn-danger text-xs px-3 py-1.5"
                                                            disabled={deletingExpenseId === expense.id}
                                                            onClick={async () => {
                                                                if (!window.confirm('Delete this expense?')) return;
                                                                setDeletingExpenseId(expense.id);
                                                                try {
                                                                    await expenseService.deleteExpense(expense.id);
                                                                    await loadData();
                                                                } catch (err) {
                                                                    console.error('Failed to delete expense:', err);
                                                                    alert('Failed to delete expense');
                                                                } finally {
                                                                    setDeletingExpenseId(null);
                                                                }
                                                            }}
                                                        >
                                                            {deletingExpenseId === expense.id ? (
                                                                <span className="spinner-dark" />
                                                            ) : (
                                                                'Delete'
                                                            )}
                                                        </button>
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    })
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Sidebar */}
                    <div className="space-y-5">
                        {/* Friends quick view */}
                        <div className="card">
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-base font-bold text-slate-900">My Friends</h2>
                                <Link to="/friends" className="text-xs text-indigo-600 hover:text-indigo-700 font-semibold">
                                    View all
                                </Link>
                            </div>

                            {friends.length === 0 ? (
                                <p className="text-sm text-slate-400 italic text-center py-2">No friends added yet.</p>
                            ) : (
                                <div className="space-y-2">
                                    {friends.slice(0, 5).map((f) => (
                                        <div key={f.id} className="flex items-center gap-3 py-1.5">
                                            <Avatar name={f.name} size="sm" />
                                            <div className="min-w-0">
                                                <p className="text-sm font-medium text-slate-800 truncate">{f.name}</p>
                                                <p className="text-xs text-slate-400 truncate">{f.email}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}

                            <Link
                                to="/friends"
                                className="mt-4 flex items-center justify-center gap-1.5 w-full py-2 text-xs text-slate-500 border border-dashed border-slate-200 rounded-lg hover:bg-slate-50 hover:border-slate-300 hover:text-slate-700 transition-all font-medium"
                            >
                                <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                                </svg>
                                Add a friend
                            </Link>
                        </div>

                        {/* Balance card */}
                        <div className="rounded-2xl p-6 bg-gradient-to-br from-indigo-600 to-violet-700 text-white">
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-sm font-semibold opacity-90">Your Net Balance</h2>
                                <div className="w-7 h-7 bg-white/15 rounded-lg flex items-center justify-center">
                                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                    </svg>
                                </div>
                            </div>
                            <p className={`text-3xl font-bold mb-1 ${stats.net < 0 ? 'text-red-300' : ''}`}>
                                {stats.net >= 0 ? '+' : ''}${stats.net.toFixed(2)}
                            </p>
                            <p className="text-xs opacity-70">
                                {stats.net > 0
                                    ? 'Others owe you money'
                                    : stats.net < 0
                                    ? 'You owe others money'
                                    : 'All settled up!'}
                            </p>
                            <div className="mt-4 pt-4 border-t border-white/20 grid grid-cols-2 gap-3 text-xs">
                                <div>
                                    <p className="opacity-60 mb-0.5">Lent</p>
                                    <p className="font-semibold text-emerald-300">+${stats.totalLent.toFixed(2)}</p>
                                </div>
                                <div>
                                    <p className="opacity-60 mb-0.5">Owe</p>
                                    <p className="font-semibold text-red-300">-${stats.totalOwed.toFixed(2)}</p>
                                </div>
                            </div>
                        </div>

                        {/* Settle up card */}
                        <div className="card">
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-base font-bold text-slate-900">Settle up</h2>
                                <span className="text-xs text-slate-500">{friends.length} friends</span>
                            </div>

                            {friends.length === 0 ? (
                                <p className="text-sm text-slate-400 italic">
                                    Add friends first to settle balances.
                                </p>
                            ) : (
                                <div className="space-y-3">
                                    <div>
                                        <label className="block text-xs font-medium text-slate-600 mb-1.5">
                                            Pay to
                                        </label>
                                        <select
                                            className="input"
                                            value={settleToUserId}
                                            onChange={(e) => {
                                                setSettleToUserId(e.target.value);
                                                setSettleError('');
                                                setSettleSuccess('');
                                            }}
                                        >
                                            <option value="">Select a friend</option>
                                            {friends.map((f) => (
                                                <option key={f.id} value={f.id}>
                                                    {f.name}
                                                </option>
                                            ))}
                                        </select>
                                    </div>

                                    <div>
                                        <label className="block text-xs font-medium text-slate-600 mb-1.5">
                                            Amount ($)
                                        </label>
                                        <input
                                            type="number"
                                            min="0.01"
                                            step="0.01"
                                            className="input"
                                            value={settleAmount}
                                            onChange={(e) => {
                                                setSettleAmount(e.target.value);
                                                setSettleError('');
                                                setSettleSuccess('');
                                            }}
                                            placeholder="0.00"
                                        />
                                    </div>

                                    {settleError && (
                                        <div className="flex items-start gap-2 text-sm bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg">
                                            <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                                            </svg>
                                            <div>{settleError}</div>
                                        </div>
                                    )}

                                    {settleSuccess && (
                                        <div className="flex items-start gap-2 text-sm bg-emerald-50 border border-emerald-200 text-emerald-800 px-3 py-2 rounded-lg">
                                            <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                                            </svg>
                                            <div>{settleSuccess}</div>
                                        </div>
                                    )}

                                    <button
                                        type="button"
                                        disabled={settling || !settleToUserId || !settleAmount || parseFloat(settleAmount) <= 0}
                                        className="btn btn-primary w-full py-2.5"
                                        onClick={async () => {
                                            setSettling(true);
                                            setSettleError('');
                                            setSettleSuccess('');
                                            try {
                                                await expenseService.settleBalance(
                                                    settleToUserId,
                                                    parseFloat(settleAmount)
                                                );
                                                setSettleSuccess('Settlement recorded successfully.');
                                                setSettleAmount('');
                                                setSettleToUserId('');
                                                await loadData();
                                            } catch (err) {
                                                const msg = err?.response?.data?.error || 'Failed to settle balance';
                                                setSettleError(msg);
                                            } finally {
                                                setSettling(false);
                                            }
                                        }}
                                    >
                                        {settling ? <span className="spinner-dark" /> : 'Settle'}
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </main>

            {/* Edit Expense Modal */}
            {editExpenseId && (
                <div className="fixed inset-0 z-50 bg-slate-900/40 flex items-center justify-center p-4">
                    <div className="w-full max-w-md bg-white rounded-2xl shadow-xl border border-slate-200 p-6 animate-fade-in">
                        <div className="flex items-center justify-between mb-4">
                            <h3 className="text-lg font-bold text-slate-900">Edit expense</h3>
                            <button
                                type="button"
                                className="btn btn-ghost text-xs px-3 py-1.5"
                                onClick={() => {
                                    setEditExpenseId(null);
                                    setEditError('');
                                }}
                            >
                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                                </svg>
                                Close
                            </button>
                        </div>

                        {editError && (
                            <div className="flex items-start gap-2 text-sm bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg mb-3">
                                <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                                </svg>
                                <div>{editError}</div>
                            </div>
                        )}

                        <form
                            onSubmit={async (e) => {
                                e.preventDefault();
                                setEditError('');
                                try {
                                    await expenseService.updateExpense(editExpenseId, {
                                        description: editDescription,
                                        total_amount: parseFloat(editTotalAmount),
                                    });
                                    setEditExpenseId(null);
                                    setEditDescription('');
                                    setEditTotalAmount('');
                                    await loadData();
                                } catch (err) {
                                    const msg = err?.response?.data?.error || 'Failed to update expense';
                                    setEditError(msg);
                                }
                            }}
                            className="space-y-4"
                        >
                            <div>
                                <label className="block text-xs font-medium text-slate-600 mb-1.5">
                                    Description
                                </label>
                                <input
                                    className="input"
                                    value={editDescription}
                                    onChange={(e) => setEditDescription(e.target.value)}
                                    required
                                />
                            </div>

                            <div>
                                <label className="block text-xs font-medium text-slate-600 mb-1.5">
                                    Total Amount ($)
                                </label>
                                <input
                                    type="number"
                                    min="0.01"
                                    step="0.01"
                                    className="input"
                                    value={editTotalAmount}
                                    onChange={(e) => setEditTotalAmount(e.target.value)}
                                    required
                                />
                            </div>

                            <div className="flex gap-3">
                                <button type="submit" className="btn btn-primary flex-1 py-2.5">
                                    Save
                                </button>
                                <button
                                    type="button"
                                    className="btn btn-secondary flex-1 py-2.5"
                                    onClick={() => {
                                        setEditExpenseId(null);
                                        setEditError('');
                                    }}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}
