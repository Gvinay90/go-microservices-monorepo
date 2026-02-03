import { useState, useEffect } from 'react';

export default function CreateExpenseForm({ friends, currentUser, onSubmit, onCancel }) {
    const [description, setDescription] = useState('');
    const [amount, setAmount] = useState('');
    const [selectedFriends, setSelectedFriends] = useState([]);

    const handleSubmit = (e) => {
        e.preventDefault();

        const totalAmount = parseFloat(amount);
        if (!totalAmount || totalAmount <= 0) {
            alert("Please enter a valid amount");
            return;
        }

        // Participants include the current user + selected friends
        // Wait, does the user want to split with themselves? Usually yes.
        // "I select who I want to share with"

        const participants = [currentUser.user_id, ...selectedFriends];
        const splitAmount = totalAmount / participants.length;

        const splits = participants.map(userId => ({
            user_id: userId,
            amount: splitAmount // Equal split
        }));

        onSubmit({
            description,
            total_amount: totalAmount,
            paid_by: currentUser.user_id, // User paid
            splits
        });
    };

    const toggleFriend = (friendId) => {
        if (selectedFriends.includes(friendId)) {
            setSelectedFriends(selectedFriends.filter(id => id !== friendId));
        } else {
            setSelectedFriends([...selectedFriends, friendId]);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="mb-6 p-6 bg-white rounded-lg shadow-lg border border-indigo-100">
            <h3 className="text-xl font-bold mb-4 text-indigo-700">Add New Expense</h3>

            <div className="space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <input
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        className="input w-full"
                        placeholder="e.g. Dinner at Mario's"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Total Amount ($)</label>
                    <input
                        type="number"
                        step="0.01"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        className="input w-full"
                        placeholder="0.00"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">Split with:</label>
                    <div className="max-h-40 overflow-y-auto border rounded-md p-2 bg-gray-50">
                        {friends.length === 0 ? (
                            <p className="text-sm text-gray-500 italic">No friends found. Add some friends first!</p>
                        ) : (
                            friends.map(friend => (
                                <div key={friend.id} className="flex items-center space-x-2 py-1">
                                    <input
                                        type="checkbox"
                                        id={`friend-${friend.id}`}
                                        checked={selectedFriends.includes(friend.id)}
                                        onChange={() => toggleFriend(friend.id)}
                                        className="rounded text-indigo-600 focus:ring-indigo-500"
                                    />
                                    <label htmlFor={`friend-${friend.id}`} className="text-sm text-gray-700 cursor-pointer select-none">
                                        {friend.name} ({friend.email})
                                    </label>
                                </div>
                            ))
                        )}
                    </div>
                </div>

                <div className="pt-2">
                    <p className="text-sm text-gray-600 mb-2">
                        You and {selectedFriends.length} others will owe
                        <span className="font-bold text-indigo-600 ml-1">
                            ${(amount && (parseFloat(amount) / (selectedFriends.length + 1)).toFixed(2)) || '0.00'}
                        </span> each.
                    </p>
                </div>

                <div className="flex gap-3 pt-2">
                    <button type="submit" className="btn btn-primary flex-1">
                        Save Expense
                    </button>
                    <button
                        type="button"
                        onClick={onCancel}
                        className="btn btn-secondary flex-1"
                    >
                        Cancel
                    </button>
                </div>
            </div>
        </form>
    );
}
