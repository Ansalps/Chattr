import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const PROFILE_STORAGE_KEY = 'chattr_profile_info';

type CustomLink = {
  id: string;
  label: string;
  url: string;
};

type ProfileFormState = {
  name: string;
  bio: string;
  profileImageUrl: string;
  customLinks: CustomLink[];
};

const emptyForm: ProfileFormState = {
  name: '',
  bio: '',
  profileImageUrl: '',
  customLinks: [],
};

const readStoredProfile = (): ProfileFormState => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) {
      return emptyForm;
    }

    const parsed = JSON.parse(raw) as Partial<ProfileFormState & { username?: string; links?: Record<string, string> }>;

    const storedCustomLinks = Array.isArray(parsed.customLinks)
      ? (parsed.customLinks
          .filter((entry) => Boolean(entry) && typeof entry === 'object') as Array<Record<string, unknown>>)
          .map((entry, index) => ({
            id: typeof entry.id === 'string' ? entry.id : `link-${index}`,
            label: typeof entry.label === 'string' ? entry.label : 'Link',
            url: typeof entry.url === 'string' ? entry.url : '',
          }))
          .filter((entry) => entry.url.trim())
      : [];

    const legacyLinks = parsed.links && typeof parsed.links === 'object' ? parsed.links : {};
    const legacyEntries = Object.entries(legacyLinks)
      .filter(([, value]) => typeof value === 'string' && value.trim())
      .map(([key, value], index) => ({
        id: `legacy-${index}`,
        label: key
          .replace(/_/g, ' ')
          .replace(/\b\w/g, (char) => char.toUpperCase()),
        url: value,
      }));

    return {
      name: typeof parsed.name === 'string' ? parsed.name : '',
      bio: typeof parsed.bio === 'string' ? parsed.bio : '',
      profileImageUrl: typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : '',
      customLinks: storedCustomLinks.length > 0 ? storedCustomLinks : legacyEntries,
    };
  } catch (error) {
    return emptyForm;
  }
};

export default function EditProfilePage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<ProfileFormState>(emptyForm);
  const [isSaving, setIsSaving] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    setForm(readStoredProfile());
  }, []);

  const handleChange = (field: keyof ProfileFormState, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const addCustomLink = () => {
    setForm((current) => ({
      ...current,
      customLinks: [...current.customLinks, { id: `link-${Date.now()}`, label: '', url: '' }],
    }));
  };

  const updateCustomLink = (id: string, field: 'label' | 'url', value: string) => {
    setForm((current) => ({
      ...current,
      customLinks: current.customLinks.map((link) =>
        link.id === id ? { ...link, [field]: value } : link,
      ),
    }));
  };

  const removeCustomLink = (id: string) => {
    setForm((current) => ({
      ...current,
      customLinks: current.customLinks.filter((link) => link.id !== id),
    }));
  };

  const handleImageSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        setForm((current) => ({ ...current, profileImageUrl: reader.result as string }));
      }
    };
    reader.readAsDataURL(file);
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSaving(true);

    const validLinks = form.customLinks
      .filter((link) => link.label.trim() || link.url.trim())
      .map((link) => ({
        id: link.id,
        label: link.label.trim(),
        url: link.url.trim(),
      }))
      .filter((link) => link.url);

    const normalized = {
      ...form,
      username: '',
      customLinks: validLinks,
      links: Object.fromEntries(
        validLinks.map((link) => [link.label.toLowerCase().replace(/\s+/g, '_'), link.url]),
      ),
    };

    localStorage.setItem(PROFILE_STORAGE_KEY, JSON.stringify(normalized));

    setTimeout(() => {
      setIsSaving(false);
      navigate('/profile');
    }, 200);
  };

  return (
    <div className="min-h-screen bg-[#f3f7fb] px-4 py-8 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl rounded-[32px] border border-sky-100 bg-white p-6 shadow-sm sm:p-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-600">Profile</p>
            <h1 className="mt-2 text-3xl font-black text-slate-900">Edit profile</h1>
          </div>
          <button
            type="button"
            onClick={() => navigate('/profile')}
            className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 shadow-sm"
          >
            Back to profile
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div className="flex items-center gap-5 rounded-[28px] border border-sky-100 bg-sky-50 p-4">
            <div className="flex h-20 w-20 items-center justify-center overflow-hidden rounded-[24px] bg-gradient-to-br from-sky-500 to-blue-600 text-2xl font-black text-white shadow-lg shadow-sky-300/40">
              {form.profileImageUrl ? (
                <img src={form.profileImageUrl} alt={form.name || 'Profile avatar'} className="h-full w-full object-cover" />
              ) : (
                <span>{(form.name || 'A').slice(0, 1).toUpperCase()}</span>
              )}
            </div>

            <div className="flex-1">
              <label className="mb-1 block text-sm font-semibold text-slate-600">Profile image</label>
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="rounded-full bg-sky-600 px-4 py-2 text-sm font-semibold text-white shadow-sm"
                >
                  Upload image
                </button>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={handleImageSelect}
                />
                <span className="text-xs text-slate-500">{form.profileImageUrl ? 'Image selected' : 'No image selected'}</span>
              </div>
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm font-semibold text-slate-600">Full name</label>
            <input
              type="text"
              value={form.name}
              onChange={(event) => handleChange('name', event.target.value)}
              placeholder="Jane Doe"
              className="w-full rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-semibold text-slate-600">Bio</label>
            <textarea
              value={form.bio}
              rows={5}
              onChange={(event) => handleChange('bio', event.target.value)}
              placeholder="Tell people a little about yourself..."
              className="w-full resize-none rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
            />
          </div>

          <div className="space-y-4 rounded-[28px] border border-slate-200 bg-slate-50 p-4">
            <div className="flex items-center justify-between gap-3">
              <label className="block text-sm font-semibold text-slate-600">Custom links</label>
              <button
                type="button"
                onClick={addCustomLink}
                className="rounded-full bg-sky-600 px-3 py-1.5 text-xs font-semibold text-white"
              >
                + Add link
              </button>
            </div>

            {form.customLinks.length === 0 ? (
              <div className="rounded-2xl border border-dashed border-slate-200 bg-white px-3 py-4 text-sm text-slate-500">
                No links added yet.
              </div>
            ) : (
              form.customLinks.map((link) => (
                <div key={link.id} className="rounded-2xl border border-slate-200 bg-white p-3">
                  <div className="mb-2 flex items-center justify-between gap-2">
                    <span className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">Link</span>
                    <button
                      type="button"
                      onClick={() => removeCustomLink(link.id)}
                      className="text-xs font-medium text-red-500"
                    >
                      Remove
                    </button>
                  </div>

                  <div className="space-y-3">
                    <input
                      type="text"
                      value={link.label}
                      onChange={(event) => updateCustomLink(link.id, 'label', event.target.value)}
                      placeholder="Label (e.g. Portfolio, GitHub)"
                      className="w-full rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
                    />
                    <input
                      type="url"
                      value={link.url}
                      onChange={(event) => updateCustomLink(link.id, 'url', event.target.value)}
                      placeholder="https://example.com"
                      className="w-full rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
                    />
                  </div>
                </div>
              ))
            )}
          </div>

          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={() => navigate('/profile')}
              className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="rounded-full bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-2.5 text-sm font-semibold text-white shadow-lg shadow-sky-300/40 disabled:cursor-not-allowed disabled:opacity-70"
            >
              {isSaving ? 'Saving...' : 'Save changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
