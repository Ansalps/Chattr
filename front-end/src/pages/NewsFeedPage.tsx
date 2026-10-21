import { useEffect, useMemo, useState } from 'react';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

type FeedPost = {
  id: string;
  author: string;
  handle: string;
  time: string;
  content: string;
  likes: number;
  comments: number;
  shares: number;
  accent: string;
  tag: string;
};

const fallbackPosts: FeedPost[] = [
  {
    id: 'welcome-1',
    author: 'Chattr Team',
    handle: '@chattr',
    time: '2h ago',
    content:
      'Welcome to your new home feed. Start following creators, share updates, and keep the conversation moving with your community.',
    likes: 324,
    comments: 19,
    shares: 12,
    accent: 'from-sky-400 to-blue-600',
    tag: 'Community',
  },
  {
    id: 'welcome-2',
    author: 'Maya Johnson',
    handle: '@maya',
    time: '4h ago',
    content:
      'Designing clearer digital experiences for small teams. Small product choices really do build trust over time.',
    likes: 188,
    comments: 24,
    shares: 8,
    accent: 'from-violet-400 to-fuchsia-600',
    tag: 'Product',
  },
  {
    id: 'welcome-3',
    author: 'Alex Rivers',
    handle: '@alex',
    time: 'Yesterday',
    content:
      'Shipping one thoughtful feature a day beats chasing every idea. Keep it simple, measurable, and useful.',
    likes: 542,
    comments: 38,
    shares: 21,
    accent: 'from-emerald-400 to-teal-600',
    tag: 'Growth',
  },
];

const formatCount = (value: number) =>
  new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 }).format(value);

const normalizeFeedPosts = (value: unknown): FeedPost[] => {
  if (!Array.isArray(value)) return fallbackPosts;

  return value
    .map((item) => {
      const source = item as Record<string, unknown>;
      const author =
        (typeof source.author === 'string' && source.author) ||
        (typeof source.name === 'string' && source.name) ||
        (typeof source.UserName === 'string' && source.UserName) ||
        (typeof source.user === 'object' && source.user && typeof (source.user as Record<string, unknown>).name === 'string'
          ? ((source.user as Record<string, unknown>).name as string)
          : 'Chattr User');

      const handle =
        (typeof source.handle === 'string' && source.handle) ||
        `@${author.toLowerCase().replace(/\s+/g, '') || 'user'}`;

      const content =
        (typeof source.content === 'string' && source.content) ||
        (typeof source.caption === 'string' && source.caption) ||
        'Shared a new update with the community.';

      const likes = Number(source.likes ?? source.like_count ?? source.likesCount ?? 0) || 0;
      const comments = Number(source.comments ?? source.comment_count ?? source.commentCount ?? 0) || 0;
      const shares = Number(source.shares ?? source.share_count ?? source.shareCount ?? 0) || 0;

      return {
        id: String(source.id ?? source.post_id ?? `${author}-${Math.random()}`),
        author,
        handle,
        time: String(source.time ?? source.created_at ?? 'just now'),
        content,
        likes,
        comments,
        shares,
        accent:
          typeof source.accent === 'string' && source.accent
            ? source.accent
            : 'from-sky-400 to-blue-600',
        tag: typeof source.tag === 'string' && source.tag ? source.tag : 'Update',
      };
    })
    .filter((post) => !!post.content);
};

export default function NewsFeedPage() {
  const [posts, setPosts] = useState<FeedPost[]>(fallbackPosts);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchNewsFeed = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/user/newsfeed`, {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        const payload = await response.json();

        if (!response.ok || payload.status_code >= 400) {
          throw new Error(payload.message || 'Unable to load your news feed right now.');
        }

        const feedData =
          Array.isArray(payload.data)
            ? payload.data
            : Array.isArray(payload.data?.posts)
              ? payload.data.posts
              : Array.isArray(payload.data?.newsfeed)
                ? payload.data.newsfeed
                : Array.isArray(payload.data?.items)
                  ? payload.data.items
                  : fallbackPosts;

        setPosts(normalizeFeedPosts(feedData));
      } catch (feedError) {
        const message =
          feedError instanceof Error ? feedError.message : 'Unable to load your news feed right now.';

        setError(message);
        setPosts(fallbackPosts);
      } finally {
        setIsLoading(false);
      }
    };

    fetchNewsFeed();
  }, []);

  const quickStats = useMemo(
    () => [
      { label: 'Followers', value: '18.4K' },
      { label: 'Following', value: '612' },
      { label: 'Posts', value: '1.2K' },
      { label: 'Reach', value: '92.6K' },
    ],
    [],
  );

  const sidebarPeople = [
    { name: 'Rina Shah', role: 'Product Lead', accent: 'from-pink-400 to-rose-500' },
    { name: 'Noah Kim', role: 'UX Strategist', accent: 'from-sky-400 to-cyan-500' },
    { name: 'Aisha Noor', role: 'Community Builder', accent: 'from-violet-400 to-purple-500' },
  ];

  return (
    <div className="min-h-screen bg-[#f3f7fb] text-slate-800">
      <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <header className="mb-6 rounded-[28px] border border-sky-100 bg-white/90 p-4 shadow-[0_20px_50px_rgba(14,116,144,0.08)] backdrop-blur-xl">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-500 to-blue-600 text-lg font-black text-white shadow-lg shadow-sky-300/40">
                C
              </div>
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-[0.22em] text-sky-600">Chattr</p>
                <h1 className="text-2xl font-black tracking-tight text-slate-900">News feed</h1>
              </div>
            </div>

            <div className="flex flex-1 items-center justify-end gap-3">
              <div className="hidden w-full max-w-md rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-slate-500 md:block">
                Search creators, topics, and communities…
              </div>
              <button className="rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-lg shadow-sky-300/40">
                Share update
              </button>
            </div>
          </div>
        </header>

        <div className="grid gap-6 lg:grid-cols-[260px_minmax(0,1fr)_320px]">
          <aside className="space-y-6">
            <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
              <div className="mb-4 flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-400 to-indigo-500 text-sm font-bold text-white">
                  JD
                </div>
                <div>
                  <p className="font-bold text-slate-900">Jane Doe</p>
                  <p className="text-sm text-slate-500">@janedoe</p>
                </div>
              </div>

              <div className="space-y-3">
                {quickStats.map((stat) => (
                  <div key={stat.label} className="flex items-center justify-between rounded-2xl bg-sky-50 px-3 py-2 text-sm">
                    <span className="text-slate-500">{stat.label}</span>
                    <span className="font-bold text-slate-900">{stat.value}</span>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
              <p className="mb-4 text-sm font-bold uppercase tracking-[0.18em] text-slate-500">Trending</p>
              <div className="space-y-3 text-sm text-slate-600">
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#productDesign</div>
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#remoteWork</div>
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#creatorEconomy</div>
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#community</div>
              </div>
            </div>
          </aside>

          <main className="space-y-5">
            <div className="rounded-[28px] border border-sky-100 bg-white p-4 shadow-sm">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-400 to-blue-600 text-sm font-bold text-white">
                  JD
                </div>
                <div className="flex-1 rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-slate-500">
                  Share something with your community...
                </div>
              </div>
            </div>

            {error && (
              <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
                {error}
              </div>
            )}

            {isLoading ? (
              <div className="rounded-[28px] border border-sky-100 bg-white p-6 text-sm text-slate-500 shadow-sm">
                Loading feed...
              </div>
            ) : (
              posts.map((post) => (
                <article key={post.id} className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
                  <div className="mb-4 flex items-start justify-between gap-3">
                    <div className="flex items-start gap-3">
                      <div className={`flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br ${post.accent} text-sm font-bold text-white`}>
                        {post.author
                          .split(' ')
                          .slice(0, 2)
                          .map((part) => part[0])
                          .join('')
                          .slice(0, 2)
                          .toUpperCase()}
                      </div>

                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-bold text-slate-900">{post.author}</p>
                          <span className="text-sm text-slate-500">{post.handle}</span>
                        </div>
                        <div className="mt-1 flex items-center gap-2 text-xs text-slate-500">
                          <span>{post.time}</span>
                          <span>•</span>
                          <span className="rounded-full bg-sky-50 px-2 py-0.5 font-medium text-sky-700">{post.tag}</span>
                        </div>
                      </div>
                    </div>

                    <button className="rounded-full border border-slate-200 px-2 py-1 text-xs font-medium text-slate-500">
                      •••
                    </button>
                  </div>

                  <p className="text-[15px] leading-7 text-slate-700">{post.content}</p>

                  <div className="mt-5 flex items-center justify-between border-t border-slate-100 pt-4 text-sm text-slate-500">
                    <div className="flex items-center gap-5">
                      <span>♥ {formatCount(post.likes)}</span>
                      <span>💬 {formatCount(post.comments)}</span>
                      <span>↻ {formatCount(post.shares)}</span>
                    </div>
                    <button className="rounded-full bg-sky-50 px-3 py-1.5 font-medium text-sky-700">View post</button>
                  </div>
                </article>
              ))
            )}
          </main>

          <aside className="space-y-6">
            <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
              <p className="mb-4 text-sm font-bold uppercase tracking-[0.18em] text-slate-500">Suggested</p>
              <div className="space-y-4">
                {sidebarPeople.map((person) => (
                  <div key={person.name} className="flex items-center justify-between gap-3 rounded-2xl bg-slate-50 p-3">
                    <div className="flex items-center gap-3">
                      <div className={`flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${person.accent} text-xs font-bold text-white`}>
                        {person.name
                          .split(' ')
                          .map((part) => part[0])
                          .join('')
                          .slice(0, 2)}
                      </div>
                      <div>
                        <p className="font-semibold text-slate-800">{person.name}</p>
                        <p className="text-xs text-slate-500">{person.role}</p>
                      </div>
                    </div>
                    <button className="rounded-full bg-sky-50 px-3 py-1.5 text-xs font-medium text-sky-700">
                      Follow
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-[28px] border border-sky-100 bg-gradient-to-br from-sky-500 to-blue-600 p-5 text-white shadow-[0_20px_40px_rgba(59,130,246,0.25)]">
              <p className="text-sm uppercase tracking-[0.18em] text-sky-100">Your pulse</p>
              <p className="mt-4 text-4xl font-black">96%</p>
              <p className="mt-2 text-sm text-sky-100">Engagement is trending up after your recent posts.</p>
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}
