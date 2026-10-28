import type { ProfilePost } from './profile.types';

type ProfilePostGridProps = {
  posts: ProfilePost[];
  isLoading: boolean;
  error: string;
  onOpenPost: (post: ProfilePost, imageIndex?: number) => Promise<void> | void;
};

export function ProfilePostGrid({ posts, isLoading, error, onOpenPost }: ProfilePostGridProps) {
  return (
    <div className="mt-6">
      {error && (
        <div className="mb-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="rounded-[28px] border border-sky-100 bg-white p-6 text-sm text-slate-500 shadow-sm">
          Loading profile posts...
        </div>
      ) : posts.length === 0 ? (
        <div className="rounded-[28px] border border-sky-100 bg-white p-6 text-sm text-slate-500 shadow-sm">
          No posts yet for this profile.
        </div>
      ) : (
        <div className="grid grid-cols-3 gap-0.5 bg-slate-200">
          {posts.map((post) => (
            <article key={post.id} className="w-full bg-white">
              {post.imageUrls.length > 0 && (
                <div className="overflow-hidden">
                  <div className="grid grid-cols-1">
                    {post.imageUrls.slice(0, 4).map((imageUrl, index) => (
                      <button
                        key={`${post.id}-${index}`}
                        type="button"
                        onClick={() => void onOpenPost(post, index)}
                        className="group relative aspect-square w-full overflow-hidden bg-slate-100"
                      >
                        <img
                          src={imageUrl}
                          alt={`Profile post ${index + 1}`}
                          className="h-full w-full object-cover transition duration-200 group-hover:scale-105"
                        />
                        <div className="absolute inset-0 flex items-center justify-center gap-4 bg-slate-900/45 opacity-0 transition duration-200 group-hover:opacity-100">
                          <span className={`text-sm font-semibold ${post.isLiked ? 'text-rose-500' : 'text-white'}`}>
                            ♥ {post.likes}
                          </span>
                          <span className="text-sm font-semibold text-white">💬 {post.comments}</span>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
