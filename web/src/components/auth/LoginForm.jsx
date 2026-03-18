import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';

function AppLogo({ size = 'md' }) {
    const sz = size === 'sm' ? 'w-8 h-8' : 'w-10 h-10';
    const txt = size === 'sm' ? 'text-base' : 'text-xl';
    return (
        <div className="flex items-center gap-2.5">
            <div className={`${sz} bg-white/20 rounded-xl flex items-center justify-center flex-shrink-0`}>
                <svg viewBox="0 0 24 24" className="w-5 h-5 fill-none stroke-white stroke-2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="5" y="2" width="14" height="20" rx="2" />
                    <line x1="9" y1="7" x2="15" y2="7" />
                    <line x1="9" y1="11" x2="15" y2="11" />
                    <line x1="9" y1="15" x2="12" y2="15" />
                </svg>
            </div>
            <span className={`${txt} font-bold tracking-tight text-white`}>SplitEase</span>
        </div>
    );
}

function BrandPanel() {
    const features = [
        'Track shared expenses in real time',
        'Split bills fairly across friends',
        'Settle balances with full clarity',
    ];
    return (
        <div className="hidden lg:flex w-[46%] flex-shrink-0 bg-gradient-to-br from-indigo-600 via-indigo-700 to-violet-700 flex-col justify-between p-12 text-white">
            <AppLogo />
            <div>
                <h2 className="text-4xl font-bold leading-tight mb-4">
                    Split bills,<br />not friendships.
                </h2>
                <p className="text-indigo-200 text-base leading-relaxed">
                    Track shared expenses and settle up with ease — no awkward money conversations.
                </p>
                <ul className="mt-8 space-y-3">
                    {features.map((f) => (
                        <li key={f} className="flex items-center gap-3 text-indigo-100 text-sm">
                            <span className="w-5 h-5 rounded-full bg-white/20 flex items-center justify-center flex-shrink-0">
                                <svg viewBox="0 0 12 12" className="w-3 h-3" fill="none" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                                    <polyline points="2,6 5,9 10,3" />
                                </svg>
                            </span>
                            {f}
                        </li>
                    ))}
                </ul>
            </div>
            <p className="text-indigo-300 text-xs">© {new Date().getFullYear()} SplitEase</p>
        </div>
    );
}

export default function LoginForm() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const { login, isLoading, error, clearError } = useAuthStore();
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        clearError();
        const success = await login(email, password);
        if (success) {
            navigate('/dashboard');
        }
    };

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

                    <h1 className="text-2xl font-bold text-slate-900 mb-1">Welcome back</h1>
                    <p className="text-slate-500 text-sm mb-8">Sign in to your account to continue</p>

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

                        <div>
                            <div className="flex items-center justify-between mb-1.5">
                                <label className="block text-sm font-medium text-slate-700">
                                    Password
                                </label>
                                <Link to="/forgot-password" className="text-xs text-indigo-600 hover:text-indigo-700 font-medium">
                                    Forgot password?
                                </Link>
                            </div>
                            <div className="relative">
                                <input
                                    type={showPassword ? 'text' : 'password'}
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="input pr-10"
                                    placeholder="••••••••"
                                    autoComplete="current-password"
                                    required
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute inset-y-0 right-3 flex items-center text-slate-400 hover:text-slate-600"
                                    tabIndex={-1}
                                >
                                    {showPassword ? (
                                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                                        </svg>
                                    ) : (
                                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                        </svg>
                                    )}
                                </button>
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={isLoading}
                            className="btn btn-primary w-full mt-2 py-2.5"
                        >
                            {isLoading ? <span className="spinner" /> : 'Sign in'}
                        </button>
                    </form>

                    <p className="mt-6 text-center text-sm text-slate-500">
                        Don&apos;t have an account?{' '}
                        <Link to="/register" className="text-indigo-600 hover:text-indigo-700 font-semibold">
                            Create one
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}
