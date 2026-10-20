import { Link } from 'react-router-dom';

const stats = [
  { label: 'Daily posts', value: '1.2M+' },
  { label: 'Creators', value: '240K' },
  { label: 'Communities', value: '18K' },
  { label: 'Uptime', value: '99.9%' },
];

const features = [
  {
    title: 'Real-time communities',
    text: 'Build vibrant spaces where people connect, share moments, and discover ideas instantly.',
  },
  {
    title: 'Creator-first tools',
    text: 'Empower creators with publishing tools, engagement insights, and monetization-ready experiences.',
  },
  {
    title: 'Safe by design',
    text: 'Smart moderation, clear policies, and community-level controls help every conversation stay healthy.',
  },
];

const testimonials = [
  {
    name: 'Aisha K.',
    role: 'Community Lead',
    quote: 'Chattr gives our community the speed of social media with the clarity of a modern platform ecosystem.',
  },
  {
    name: 'Daniel R.',
    role: 'Creator',
    quote: 'The experience feels premium and frictionless. Our audience engagement jumped within weeks.',
  },
];

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[#f5f9ff] text-slate-900">
      <header className="mx-auto flex max-w-7xl items-center justify-between px-4 py-5 sm:px-6 lg:px-8">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-500 to-blue-600 text-lg font-black text-white shadow-lg shadow-sky-300/40">
            C
          </div>
          <div>
            <p className="text-lg font-black tracking-tight">Chattr</p>
          </div>
        </div>

        <nav className="hidden items-center gap-8 text-sm font-medium text-slate-600 md:flex">
          <a href="#features" className="transition hover:text-sky-700">Features</a>
          <a href="#community" className="transition hover:text-sky-700">Community</a>
          <a href="#testimonials" className="transition hover:text-sky-700">Stories</a>
        </nav>

        <div className="flex items-center gap-3">
          <Link
            to="/login"
            className="rounded-full border border-sky-200 bg-white px-4 py-2 text-sm font-semibold text-sky-700 transition hover:border-sky-300 hover:bg-sky-50"
          >
            User Login
          </Link>
          <Link
            to="/signup"
            className="rounded-full bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-sky-300/40 transition hover:brightness-105"
          >
            User Sign up
          </Link>
        </div>
      </header>

      <main>
        <section className="mx-auto grid max-w-7xl items-center gap-10 px-4 pb-16 pt-10 sm:px-6 lg:grid-cols-[1.1fr_0.9fr] lg:px-8 lg:pb-24 lg:pt-16">
          <div>
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-sky-200 bg-sky-100 px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.18em] text-sky-700">
              Social media, reimagined
            </div>

            <h1 className="max-w-xl text-4xl font-black tracking-tight text-slate-900 sm:text-5xl lg:text-6xl">
              Build meaningful conversations at scale.
            </h1>

            <p className="mt-6 max-w-lg text-lg leading-8 text-slate-600">
              Chattr helps communities grow faster, creators connect deeper, and brands turn engagement into trust through a modern social platform experience.
            </p>

            <div className="mt-8 flex flex-wrap items-center gap-4">
              <Link
                to="/signup"
                className="rounded-full bg-gradient-to-r from-sky-500 to-blue-600 px-6 py-3 text-base font-semibold text-white shadow-lg shadow-sky-300/40 transition hover:brightness-105"
              >
                User Sign up
              </Link>
              <Link
                to="/login"
                className="rounded-full border border-slate-200 bg-white px-6 py-3 text-base font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              >
                User Login
              </Link>
            </div>

            <div className="mt-10 grid max-w-lg grid-cols-2 gap-4 sm:grid-cols-4">
              {stats.map((stat) => (
                <div key={stat.label} className="rounded-2xl border border-sky-100 bg-white p-3 shadow-sm">
                  <p className="text-lg font-black text-slate-900">{stat.value}</p>
                  <p className="mt-1 text-xs text-slate-500">{stat.label}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="relative">
            <div className="absolute -left-8 top-6 h-32 w-32 rounded-full bg-sky-200/60 blur-3xl" />
            <div className="absolute -right-4 bottom-8 h-40 w-40 rounded-full bg-violet-200/50 blur-3xl" />

            <div className="relative overflow-hidden rounded-[32px] border border-sky-100 bg-white p-4 shadow-[0_25px_70px_rgba(14,116,144,0.12)]">
              <div className="rounded-[24px] bg-gradient-to-br from-sky-50 via-white to-blue-50 p-4">
                <div className="mb-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-sm font-bold text-white">
                      A
                    </div>
                    <div>
                      <p className="font-semibold text-slate-800">Ariana</p>
                      <p className="text-xs text-slate-500">@ariana • 2m ago</p>
                    </div>
                  </div>
                  <span className="rounded-full bg-emerald-100 px-2.5 py-1 text-xs font-semibold text-emerald-700">
                    Live
                  </span>
                </div>

                <div className="rounded-2xl bg-white p-3 shadow-sm">
                  <div className="mb-3 h-44 rounded-2xl bg-[linear-gradient(135deg,#dbeafe,#dbeafe_35%,#c7d2fe_100%)]" />
                  <p className="text-sm leading-6 text-slate-600">
                    “Building a community that feels personal, creative, and connected is the future of social media.”
                  </p>
                </div>

                <div className="mt-4 flex items-center justify-between text-sm text-slate-500">
                  <div className="flex items-center gap-4">
                    <span>♥ 1.2k</span>
                    <span>💬 346</span>
                    <span>↻ 98</span>
                  </div>
                  <span className="rounded-full bg-sky-100 px-2.5 py-1 text-xs font-semibold text-sky-700">
                    Trending
                  </span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="features" className="mx-auto max-w-7xl px-4 py-20 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-2xl text-center">
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-sky-700">Why Chattr</p>
            <h2 className="mt-4 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
              Every feature built for modern social experiences.
            </h2>
          </div>

          <div className="mt-12 grid gap-6 md:grid-cols-3">
            {features.map((feature) => (
              <div key={feature.title} className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm">
                <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-sky-100 text-xl text-sky-700">★</div>
                <h3 className="text-xl font-bold text-slate-900">{feature.title}</h3>
                <p className="mt-3 text-base leading-7 text-slate-600">{feature.text}</p>
              </div>
            ))}
          </div>
        </section>

        <section id="community" className="bg-white py-20">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid gap-8 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
              <div>
                <p className="text-sm font-semibold uppercase tracking-[0.2em] text-sky-700">Community growth</p>
                <h2 className="mt-4 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
                  Turn attention into a thriving creator ecosystem.
                </h2>
                <p className="mt-5 text-lg leading-8 text-slate-600">
                  Chattr gives communities the tools to share, discover, and engage with clarity—from creators launching communities to brands creating loyal fan spaces.
                </p>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                {[
                  ['Creator friendly', 'Publish content with speed and confidence.'],
                  ['Community insights', 'Track engagement and campaign momentum in real time.'],
                  ['Secure moderation', 'Protect your audience with smarter policy actions.'],
                  ['Monetization ready', 'Support premium experiences and sustainable growth.'],
                ].map(([title, text]) => (
                  <div key={title} className="rounded-[24px] border border-sky-100 bg-sky-50 p-5">
                    <p className="text-lg font-bold text-slate-900">{title}</p>
                    <p className="mt-2 text-sm leading-6 text-slate-600">{text}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section id="testimonials" className="mx-auto max-w-7xl px-4 py-20 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-2xl text-center">
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-sky-700">Loved by teams</p>
            <h2 className="mt-4 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
              Real conversations, real growth.
            </h2>
          </div>

          <div className="mt-12 grid gap-6 md:grid-cols-2">
            {testimonials.map((item) => (
              <div key={item.name} className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm">
                <p className="text-lg leading-8 text-slate-700">“{item.quote}”</p>
                <div className="mt-6">
                  <p className="font-bold text-slate-900">{item.name}</p>
                  <p className="text-sm text-slate-500">{item.role}</p>
                </div>
              </div>
            ))}
          </div>
        </section>
      </main>

      <footer className="border-t border-sky-100 bg-white">
        <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 px-4 py-8 text-sm text-slate-500 sm:flex-row sm:px-6 lg:px-8">
          <p>© 2026 Chattr. Built for communities.</p>
          <div className="flex items-center gap-4">
            <a href="#features" className="hover:text-sky-700">Features</a>
            <a href="#community" className="hover:text-sky-700">Community</a>
            <Link to="/admin/login" className="hover:text-sky-700">Admin</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
