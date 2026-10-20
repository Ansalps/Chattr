import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export default function UserSignup() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [userName, setUserName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccessMessage('');

    if (!name.trim() || !userName.trim() || !email.trim() || !phone.trim() || !password.trim() || !confirmPassword.trim()) {
      setError('Name, username, email, phone, password, and confirm password are required.');
      return;
    }

    if (password !== confirmPassword) {
      setError('Password and confirm password must match.');
      return;
    }

    setIsLoading(true);

    try {
      const response = await fetch(`${API_BASE_URL}/user/signup`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          Name: name,
          UserName: userName,
          Email: email,
          Password: password,
          ConfirmPassword: confirmPassword,
          phone: phone,
        }),
      });

      const payload = await response.json();

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || 'User signup failed.');
      }

      setSuccessMessage(payload.message || 'OTP sent successfully. Redirecting to verification...');
      setName('');
      setUserName('');
      setEmail('');
      setPhone('');
      setPassword('');
      setConfirmPassword('');

      sessionStorage.setItem('chattr_pending_signup_email', email);
      if (payload.data?.access_token || payload.data?.AccessToken) {
        const token = payload.data?.access_token ?? payload.data?.AccessToken;
        localStorage.setItem('chattr_user_access_token', token);
      }

      window.setTimeout(() => {
        navigate('/verify-otp', { state: { email } });
      }, 800);
    } catch (signupError) {
      const message =
        signupError instanceof Error
          ? signupError.message
          : 'Unable to create your account right now. Please try again.';
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
                  Join the community that keeps conversations alive.
                </h1>
                <p className="max-w-md text-base text-slate-600">
                  Create your Chattr account and start connecting with people, ideas, and communities you care about.
                </p>
              </div>
            </div>

            <div className="relative z-10 space-y-5">
              <div className="flex items-center gap-4 rounded-2xl border border-sky-100 bg-white p-4 shadow-sm shadow-sky-100">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-sky-400 to-blue-600 text-xl font-bold text-white">
                  24K
                </div>
                <div>
                  <p className="text-sm text-slate-500">Members</p>
                  <p className="text-lg font-semibold text-slate-800">Welcome community</p>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm text-slate-700">
                <div className="rounded-2xl border border-sky-100 bg-white p-4 shadow-sm">
                  <p className="text-slate-500">Groups</p>
                  <p className="mt-2 text-2xl font-bold text-slate-900">18K</p>
                </div>
                <div className="rounded-2xl border border-sky-100 bg-white p-4 shadow-sm">
                  <p className="text-slate-500">Creators</p>
                  <p className="mt-2 text-2xl font-bold text-slate-900">240K</p>
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
                <h2 className="text-3xl font-black tracking-tight text-slate-900">User Sign up</h2>
                <p className="mt-2 text-sm text-slate-500">
                  Create your account and start connecting today.
                </p>
              </div>

              <form onSubmit={handleSubmit} className="space-y-5">
                <div className="space-y-2">
                  <label htmlFor="name" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Full name
                  </label>
                  <input
                    id="name"
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Jane Doe"
                    autoComplete="name"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                <div className="space-y-2">
                  <label htmlFor="userName" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Username
                  </label>
                  <input
                    id="userName"
                    type="text"
                    value={userName}
                    onChange={(e) => setUserName(e.target.value)}
                    placeholder="janedoe"
                    autoComplete="username"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                <div className="space-y-2">
                  <label htmlFor="email" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Email address
                  </label>
                  <input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@example.com"
                    autoComplete="email"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                <div className="space-y-2">
                  <label htmlFor="phone" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Phone
                  </label>
                  <input
                    id="phone"
                    type="tel"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="+1 234 567 8900"
                    autoComplete="tel"
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
                    autoComplete="new-password"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                <div className="space-y-2">
                  <label htmlFor="confirmPassword" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    Confirm password
                  </label>
                  <input
                    id="confirmPassword"
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="••••••••"
                    autoComplete="new-password"
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
                  {isLoading ? 'Creating account...' : 'Create account'}
                </button>
              </form>

              <p className="mt-6 text-center text-sm text-slate-500">
                Already have an account?{' '}
                <Link to="/login" className="font-semibold text-sky-700 hover:text-sky-800">
                  Sign in
                </Link>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
