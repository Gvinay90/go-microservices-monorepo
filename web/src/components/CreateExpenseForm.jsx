import { useState } from 'react';

export default function CreateExpenseForm({ friends, currentUser, onSubmit, onCancel }) {
    const [description, setDescription] = useState('');
    const [amount, setAmount] = useState('');
    const [selectedFriends, setSelectedFriends] = useState([]);
    const [validationError, setValidationError] = useState('');

    const participantCount = selectedFriends.length + 1;
    const totalAmount = parseFloat(amount) || 0;
    const splitAmount = totalAmount > 0 ? (totalAmount / participantCount).toFixed(2) : '0.00';

    const handleSubmit = (e) => {
        e.preventDefault();
        setValidationError('');

        if (!totalAmount || totalAmount <= 0) {
            setValidationError('Please enter a valid amount greater than zero.');
            return;
        }

        const participants = [currentUser.user_id, ...selectedFriends];
        const perPerson = totalAmount / participants.length;

        onSubmit({
            description,
            total_amount: totalAmount,
            paid_by: currentUser.user_id,
            splits: participants.map((userId) => ({ user_id: userId, amount: perPerson })),
        });
    };

    const toggleFriend = (friendId) => {
        setSelectedFriends((prev) =>
            prev.includes(friendId) ? prev.filter((id) => id !== friendId) : [...prev, friendId]
        );
    };

    return (
        <div className="border border-indigo-100 bg-indigo-50/40 rounded-xl p-5">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-base font-bold text-slate-900">New Expense</h3>
                <button
                    type="button"
                    onClick={onCancel}
                    className="w-7 h-7 rounded-lg hover:bg-slate-200 flex items-center justify-center text-slate-400 hover:text-slate-600 transition-colors"
                >
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            {validationError && (
                <div className="flex items-center gap-2 bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg mb-4 text-sm">
                    <svg className="w-4 h-4 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                    {validationError}
                </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
                <div className="grid sm:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">
                            Description
                        </label>
                        <input
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="input bg-white"
                            placeholder="e.g. Dinner at Mario's"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-1.5">
                            Total Amount ($)
                        </label>
                        <div className="relative">
                            <span className="absolute inset-y-0 left-3.5 flex items-center text-slate-400 text-sm font-medium pointer-events-none">$</span>
                            <input
                                type="number"
                                step="0.01"
                                min="0.01"
                                value={amount}
                                onChange={(e) => setAmount(e.target.value)}
                                className="input bg-white pl-7"
                                placeholder="0.00"
                                required
                            />
                        </div>
                    </div>
                </div>

                <div>
                    <label className="block text-xs font-semibold text-slate-600 uppercase tracking-wide mb-2">
                        Split with
                    </label>

                    {friends.length === 0 ? (
                        <div className="border border-dashed border-slate-200 rounded-lg px-4 py-3 text-sm text-slate-400 bg-white">
                            No friends yet — add some from the Friends page first.
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                            {friends.map((friend) => {
                                const selected = selectedFriends.includes(friend.id);
                                return (
                                    <label
                                        key={friend.id}
                                        className={`flex items-center gap-3 p-2.5 rounded-lg border cursor-pointer transition-all ${
                                            selected
                                                ? 'border-indigo-400 bg-indigo-50 text-indigo-800'
                                                : 'border-slate-200 bg-white hover:border-slate-300 text-slate-700'
                                        }`}
                                    >
                                        <input
                                            type="checkbox"
                                            checked={selected}
                                            onChange={() => toggleFriend(friend.id)}
                                            className="w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500 flex-shrink-0"
                                        />
                                        <div className="min-w-0">
                                            <p className="text-sm font-medium truncate">{friend.name}</p>
                                            <p className="text-xs opacity-60 truncate">{friend.email}</p>
                                        </div>
                                    </label>
                                );
                            })}
                        </div>
                    )}
                </div>

                {/* Split preview */}
                <div className="flex items-center justify-between bg-white border border-slate-200 rounded-xl px-4 py-3">
                    <div className="text-sm text-slate-600">
                        <span className="font-medium">{participantCount}</span> {participantCount === 1 ? 'person' : 'people'} splitting
                    </div>
                    <div className="text-right">
                        <p className="text-xs text-slate-400">Each person pays</p>
                        <p className="text-lg font-bold text-indigo-600">${splitAmount}</p>
                    </div>
                </div>

                <div className="flex gap-3 pt-1">
                    <button type="submit" className="btn btn-primary flex-1 py-2.5">
                        Save Expense
                    </button>
                    <button
                        type="button"
                        onClick={onCancel}
                        className="btn btn-secondary flex-1 py-2.5"
                    >
                        Cancel
                    </button>
                </div>
            </form>
        </div>
    );
}
