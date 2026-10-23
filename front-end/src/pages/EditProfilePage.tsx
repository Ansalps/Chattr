import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const PROFILE_STORAGE_KEY = 'chattr_profile_info';

type ProfileFormState = {
  name: string;
  bio: string;
  profileImageUrl: string;
  website: string;
  linkedin: string;
  instagram: string;
};

const emptyForm: ProfileFormState = {
  name: '',
  bio: '',
  profileImageUrl: '',
  website: '',
  linkedin: '',
  instagram: '',
};

const readStoredProfile = (): ProfileFormState => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) {
      return emptyForm;
    }

    const parsed = JSON.parse(raw) as Partial<ProfileFormState & { username?: string; links?: Record<string, string> }>;
    return {
      name: typeof parsed.name === 'string' ? parsed.name : '',
      bio: typeof parsed.bio === 'string' ? parsed.bio : '',
      profileImageUrl: typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : '',
      website: typeof parsed.website === 'string' ? parsed.website : typeof parsed.links?.website === 'string' ? parsed.links.website : '',
      linkedin: typeof parsed.linkedin === 'string' ? parsed.linkedin : typeof parsed.links?.linkedin === 'string' ? parsed.links.linkedin : '',
      instagram: typeof parsed.instagram === 'string' ? parsed.instagram : typeof parsed.links?.instagram === 'string' ? parsed.links.instagram : '',
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

    const normalized = {
      ...form,
      username: '',
      links: {
        website: form.website,
        linkedin: form.linkedin,
        instagram: form.instagram,
      },
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
            <div>
              <label className="mb-1 block text-sm font-semibold text-slate-600">Website</label>
              <input
                type="url"
                value={form.website}
                onChange={(event) => handleChange('website', event.target.value)}
                placeholder="https://yourwebsite.com"
                className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
              />
            </div>

            <div>
              <label className="mb-1 block text-sm font-semibold text-slate-600">LinkedIn</label>
              <input
                type="url"
                value={form.linkedin}
                onChange={(event) => handleChange('linkedin', event.target.value)}
                placeholder="https://linkedin.com/in/yourname"
                className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
              />
            </div>

            <div>
              <label className="mb-1 block text-sm font-semibold text-slate-600">Instagram</label>
              <input
                type="url"
                value={form.instagram}
                onChange={(event) => handleChange('instagram', event.target.value)}
                placeholder="https://instagram.com/yourname"
                className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-700 outline-none transition focus:border-sky-300 focus:ring-2 focus:ring-sky-100"
              />
            </div>
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
