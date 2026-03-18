import { useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../../services/api';

function BrandPanel() {
    return (
        <div className="hidden lg:flex w-[46%] flex-shrink-0 bg-gradient-to-br from-indigo-600 via-indigo-700 to-violet-700 flex-col justify-between p-12 text-white">
            <div className="flex items-center gap-2.5">
                <div className="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center flex-shrink-0">
                    <svg viewBox="0 0 24 24" className="w-5 h-5 fill-none stroke-white stroke-2" strokeLinecap="round" strokeLinejoin="round">
                        <rect x="5" y="2" width="14" height="20" rx="2" />
                        <line x1="9" y1="7" x2="15" y2="7" />
                        <line x1="9" y1="11" x2="15" y2="11" />
                        <line x1="9" y1="15" x2="12" y2="15" />
                    </svg>
                </div>
                <span className="text-xl font-bold tracking-tight">SplitEase</span>
            </div>
            <div>
                <h2 className="text-4xl font-bold leading-tight mb-4">
                    We&rsquo;ve got<br />you covered.
                </h2>
                <p className="text-indigo-200 text-base leading-relaxed">
                    Password resets are quick and secure. You&apos;ll be back in under a minute.
                </p>
            </div>
            <p className="text-indigo-300 text-xs">© {new Date().getFullYear()} SplitEase</p>
        </div>
    );
}

export default function ForgotPasswordForm() {
    const [email, setEmail] = useState('');
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        try {
            await api.post('/auth/forgot-password', { email });
            setSuccess(true);
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to send reset email. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    if (success) {
        return (
            <div className="min-h-screen flex">
                <BrandPanel />
                <div className="flex-1 flex items-center justify-center bg-white p-8 lg:p-12">
                    <div className="w-full max-w-sm text-center animate-fade-in">
                        <div className="w-16 h-16 bg-emerald-100 rounded-full flex items-center justify-center mx-auto mb-5">
                            <svg className="w-8 h-8 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                            </svg>
                        </div>
                        <h2 className="text-2xl font-bold text-slate-900 mb-2">Check your inbox</h2>
                        <p className="text-slate-500 text-sm mb-1">
                            We&apos;ve sent reset instructions to
                        </p>
                        <p className="font-semibold text-slate-800 text-sm mb-7">{email}</p>
                        <Link to="/login" className="btn btn-primary w-full py-2.5">
                            Back to sign in
                        </Link>
                        <p className="mt-4 text-xs text-slate-400">
                            Didn&apos;t receive it? Check your spam folder or{' '}
                            <button onClick={() => setSuccess(false)} className="text-indigo-600 hover:underline">
                                try again
                            </button>
                        </p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex">
            <BrandPanel />

            {/* Form side */}
            <div className="flex-1 flex items-center justify-center bg-white p-8 lg:p-12">
                <div className="w-full max-w-sm animate-fade-in">
                    {/* Mobile logo */}
                    <div className="flex items-center gap-2 mb-8 lg:hidden">
                        <div className="w-8 h-8 bg-indigo-600 rounded-xl flex items-center justify-center">
                            <svg viewBox="0 0 24 24" className="w-4 h-4 fill-none stroke-white stroke-2" strokeLinecap="round" strokeLinejoin="round">
                                <rect x="5" y="2" width="14" height="20" rx="2" />
                                <line x1="9" y1="7" x2="15" y2="7" />
                                <line x1="9" y1="11" x2="15" y2="11" />
                                <line x1="9" y1="15" x2="12" y2="15" />
                            </svg>
                        </div>
                        <span className="text-lg font-bold text-slate-900">SplitEase</span>
                    </div>

                    <div className="w-12 h-12 bg-indigo-50 rounded-2xl flex items-center justify-center mb-6">
                        <svg className="w-6 h-6 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                        </svg>
                    </div>

                    <h1 className="text-2xl font-bold text-slate-900 mb-1">Forgot your password?</h1>
                    <p className="text-slate-500 text-sm mb-8">
                        No worries — enter your email and we&apos;ll send you reset instructions.
                    </p>

                    {error && (
                        <div className="flex items-start gap-2.5 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-5 text-sm">
                            <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                            </svg>
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1.5">
                                Email address
                            </label>
                            <input
                                type="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="input"
                                placeholder="you@example.com"
                                autoComplete="email"
                                required
                            />
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="btn btn-primary w-full py-2.5"
                        >
                            {loading ? <span className="spinner" /> : 'Send reset link'}
                        </button>
                    </form>

                    <p className="mt-6 text-center text-sm text-slate-500">
                        Remembered it?{' '}
                        <Link to="/login" className="text-indigo-600 hover:text-indigo-700 font-semibold">
                            Sign in
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}
