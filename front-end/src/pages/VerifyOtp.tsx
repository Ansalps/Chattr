import React, { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export default function VerifyOtp() {
  const navigate = useNavigate();
  const location = useLocation();
  const emailFromState = (location.state as { email?: string } | null)?.email ?? '';
  const [email, setEmail] = useState(emailFromState || sessionStorage.getItem('chattr_pending_signup_email') || '');
  const [otpCode, setOtpCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  useEffect(() => {
    if (email) {
      sessionStorage.setItem('chattr_pending_signup_email', email);
    }
  }, [email]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccessMessage('');

    if (!email.trim() || !otpCode.trim()) {
      setError('Email and OTP code are required.');
      return;
    }

    const otpValue = Number(otpCode);
    if (!Number.isInteger(otpValue) || otpValue <= 0) {
      setError('Please enter a valid OTP code.');
      return;
    }

    setIsLoading(true);

    try {
      const response = await fetch(`${API_BASE_URL}/user/verify-otp`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          otp_code: otpValue,
          purpose: 'user-signup',
          email: email,
        }),
      });

      const payload = await response.json();

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || 'OTP verification failed.');
      }

      setSuccessMessage(payload.message || 'OTP verified successfully. Redirecting...');
      sessionStorage.removeItem('chattr_pending_signup_email');

      window.setTimeout(() => {
        navigate('/login');
      }, 800);
    } catch (verifyError) {
      const message =
        verifyError instanceof Error
          ? verifyError.message
          : 'Unable to verify OTP. Please try again.';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-full bg-[#f3f7fb] text-slate-900">
      <div className="mx-auto flex min-h-screen max-w-5xl items-center justify-center px-4 py-10">
        <div className="grid w-full overflow-hidden rounded-[32px] border border-sky-100 bg-white/90 shadow-[0_25px_70px_rgba(14,116,144,0.12)] backdrop-blur-xl lg:grid-cols-[1fr_1fr]">
          <div className="relative hidden overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.22),_transparent_32%),linear-gradient(135deg,#ebf8ff_0%,#f8fbff_45%,#eff6ff_100%)] p-8 lg:flex lg:flex-col lg:justify-between">
            <div className="absolute inset-0 opacity-90">
              <div className="absolute -left-20 top-16 h-52 w-52 rounded-full bg-sky-300/40 blur-3xl" />
              <div className="absolute right-16 top-28 h-40 w-40 rounded-full bg-violet-300/30 blur-3xl" />
              <div className="absolute bottom-10 left-20 h-48 w-48 rounded-full bg-emerald-300/30 blur-3xl" />
            </div>

            <div className="relative z-10">
              <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-sky-200 bg-sky-100 px-3 py-1.5 text-xs font-semibold tracking-[0.2em] text-sky-700 uppercase">
                Secure verification
              </div>

              <div className="space-y-6">
                <h1 className="max-w-md text-5xl font-black leading-tight tracking-tight text-slate-900">
                  Verify your account.
                </h1>
                <p className="max-w-md text-base text-slate-600">
                  Enter the one-time code we sent to your email to complete signup and unlock your Chattr profile.
                </p>
              </div>
            </div>

            <div className="relative z-10 rounded-2xl border border-sky-100 bg-white p-5 shadow-sm">
              <p className="text-sm text-slate-500">Verification status</p>
              <p className="mt-2 text-2xl font-bold text-slate-900">Pending</p>
              <p className="mt-2 text-sm text-slate-600">A valid 6-digit OTP is required to activate your account.</p>
            </div>
          </div>

          <div className="flex items-center justify-center p-6 sm:p-8 lg:p-10">
            <div className="w-full max-w-md">
              <div className="mb-8 text-center lg:text-left">
                <div className="mb-3 inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-400 via-blue-500 to-indigo-600 text-2xl font-bold text-white shadow-lg shadow-sky-300/40">
                  C
                </div>
                <h2 className="text-3xl font-black tracking-tight text-slate-900">OTP Verification</h2>
                <p className="mt-2 text-sm text-slate-500">
                  We sent a code to {email || 'your email'}
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
                    placeholder="you@example.com"
                    autoComplete="email"
                    required
                    className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3.5 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
                  />
                </div>

                <div className="space-y-2">
                  <label htmlFor="otp" className="block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                    OTP code
                  </label>
                  <input
                    id="otp"
                    type="text"
                    inputMode="numeric"
                    value={otpCode}
                    onChange={(e) => setOtpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                    placeholder="123456"
                    autoComplete="one-time-code"
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
                  {isLoading ? 'Verifying...' : 'Verify OTP'}
                </button>
              </form>

              <p className="mt-6 text-center text-sm text-slate-500">
                Didn’t get the code?{' '}
                <button
                  type="button"
                  onClick={() => navigate('/signup')}
                  className="font-semibold text-sky-700 hover:text-sky-800"
                >
                  Try again
                </button>
              </p>

              <div className="mt-6 text-center">
                <Link to="/login" className="text-sm font-medium text-slate-500 hover:text-slate-700">
                  Back to login
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
