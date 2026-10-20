import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

type AdminLoginResponse = {
  status_code: number;
  message: string;
  data?: {
    admin?: {
      id?: number;
      email?: string;
    };
    Admin?: {
      ID?: number;
      Email?: string;
    };
    access_token?: string;
    refresh_token?: string;
    AccessToken?: string;
    RefreshToken?: string;
  };
};

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export default function AdminLogin() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccessMessage('');

    if (!email.trim() || !password.trim()) {
      setError('Email and password are required.');
      return;
    }

    setIsLoading(true);

    try {
      const response = await fetch(`${API_BASE_URL}/admin/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ email, password }),
      });

      const payload = (await response.json()) as AdminLoginResponse;

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || 'Admin login failed.');
      }

      const rawAdmin = (payload.data?.admin ?? payload.data?.Admin ?? {}) as {
        id?: number;
        email?: string;
        ID?: number;
        Email?: string;
      };
      const admin = {
        id: rawAdmin.id ?? rawAdmin.ID,
        email: rawAdmin.email ?? rawAdmin.Email,
      };
      const accessToken = payload.data?.access_token ?? payload.data?.AccessToken ?? '';
      const refreshToken = payload.data?.refresh_token ?? payload.data?.RefreshToken ?? '';

      if (accessToken) {
        localStorage.setItem('chattr_admin_access_token', accessToken);
      }

      if (refreshToken) {
        localStorage.setItem('chattr_admin_refresh_token', refreshToken);
      }

      if (admin.email) {
        localStorage.setItem('chattr_admin_user', JSON.stringify(admin));
      }

      setSuccessMessage(payload.message || 'Admin authenticated successfully.');
      setEmail('');
      setPassword('');

      window.setTimeout(() => {
        navigate('/admin/dashboard');
      }, 600);
    } catch (loginError) {
      const message =
        loginError instanceof Error
          ? loginError.message
          : 'Something went wrong while signing in.';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-full bg-[#f3f7fb] text-slate-900">
      <div className="mx-auto flex min-h-screen max-w-7xl items-center justify-center px-4 py-10">
        <div className="grid w-full overflow-hidden rounded-[32px] border border-sky-100 bg-white/90 shadow-[0_25px_70px_rgba(14,116,144,0.12)] backdrop-blur-xl lg:grid-cols-[1.1fr_0.9fr]">
          <div className="relative hidden overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.22),_transparent_32%),linear-gradient(135deg,#ebf8ff_0%,#f8fbff_45%,#eff6ff_100%)] p-8 lg:flex lg:flex-col lg:justify-between">
            <div className="absolute inset-0 opacity-90">
              <div className="absolute -left-20 top-16 h-52 w-52 rounded-full bg-sky-300/40 blur-3xl" />
              <div className="absolute right-16 top-28 h-40 w-40 rounded-full bg-violet-300/30 blur-3xl" />
              <div className="absolute bottom-10 left-20 h-48 w-48 rounded-full bg-emerald-300/30 blur-3xl" />
            </div>

            <div className="relative z-10">
              <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-sky-200 bg-sky-100 px-3 py-1.5 text-xs font-semibold tracking-[0.2em] text-sky-700 uppercase">
                Chattr social cloud
              </div>

              <div className="space-y-6">
                <h1 className="max-w-md text-5xl font-black leading-tight tracking-tight text-slate-900">
                  Run your social empire from one elegant dashboard.
                </h1>
                <p className="max-w-md text-base text-slate-600">
                  Moderate communities, protect your brand, and track engagement with a calmer, smarter admin workspace.
                </p>
              </div>
            </div>

            <div className="relative z-10 space-y-5">
              <div className="flex items-center gap-4 rounded-2xl border border-sky-100 bg-white p-4 shadow-sm shadow-sky-100">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-sky-400 to-blue-600 text-xl font-bold text-white">
                  24K
                </div>
                <div>
                  <p className="text-sm text-slate-500">Active users</p>
                  <p className="text-lg font-semibold text-slate-800">+18.4% this week</p>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm text-slate-700">
                <div className="rounded-2xl border border-sky-100 bg-white p-4 shadow-sm">
                  <p className="text-slate-500">Reports</p>
                  <p className="mt-2 text-2xl font-bold text-slate-900">1.2k</p>
                </div>
                <div className="rounded-2xl border border-sky-100 bg-white p-4 shadow-sm">
                  <p className="text-slate-500">Moderation</p>
                  <p className="mt-2 text-2xl font-bold text-slate-900">96%</p>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center justify-center p-6 sm:p-8 lg:p-10">
            <div className="w-full max-w-md">
              <div className="mb-8 text-center lg:text-left">
                <div className="mb-3 inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-400 via-blue-500 to-indigo-600 text-2xl font-bold text-white shadow-lg shadow-sky-300/40">
                  C
                </div>
                <h2 className="text-3xl font-black tracking-tight text-slate-900">Admin Sign in</h2>
                <p className="mt-2 text-sm text-slate-500">
                  Access your Chattr control center and social insights.
                </p>
              </div>

              <form onSubmit={handleSubmit} className="space-y-5">
                <div className="space-y-2">
                  <label htmlFor="email" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Email address
                  </label>
                  <input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="admin@chattr.com"
                    autoComplete="email"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                <div className="space-y-2">
                  <label htmlFor="password" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Password
                  </label>
                  <input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                    autoComplete="current-password"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                {error && (
                  <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
                    {error}
                  </div>
                )}

                {successMessage && (
                  <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
                    {successMessage}
                  </div>
                )}

                <button
                  type="submit"
                  disabled={isLoading}
                  className="flex w-full items-center justify-center rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-3.5 text-base font-semibold text-white shadow-lg shadow-sky-300/40 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-70"
                >
                  {isLoading ? 'Signing in...' : 'Sign in'}
                </button>
              </form>

              <div className="mt-6 flex items-center justify-center gap-3 text-xs text-slate-400">
                <span className="h-px flex-1 bg-slate-200" />
                <span>Secure gateway access</span>
                <span className="h-px flex-1 bg-slate-200" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}