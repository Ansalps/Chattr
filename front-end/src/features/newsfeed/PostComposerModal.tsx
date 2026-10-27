import type { Dispatch, SetStateAction } from 'react';

export type PostComposerModalProps = {
  isOpen: boolean;
  editingPostId: string | null;
  postText: string;
  setPostText: Dispatch<SetStateAction<string>>;
  selectedFiles: File[];
  setSelectedFiles: Dispatch<SetStateAction<File[]>>;
  composerError: string;
  composerSuccess: string;
  isPosting: boolean;
  onClose: () => void;
  onPublish: () => Promise<void> | void;
};

export function PostComposerModal({
  isOpen,
  editingPostId,
  postText,
  setPostText,
  selectedFiles,
  setSelectedFiles,
  composerError,
  composerSuccess,
  isPosting,
  onClose,
  onPublish,
}: PostComposerModalProps) {
  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
      <div className="w-full max-w-xl rounded-[30px] border border-sky-100 bg-white p-5 shadow-[0_25px_80px_rgba(15,23,42,0.2)]">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-sky-600">
              {editingPostId ? 'Edit post' : 'Create post'}
            </p>
            <h2 className="mt-2 text-2xl font-black text-slate-900">
              {editingPostId ? 'Update your post' : 'Share an update'}
            </h2>
          </div>
          <button type="button" onClick={onClose} className="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-500">
            Close
          </button>
        </div>

        <div className="space-y-4">
          <textarea
            value={postText}
            onChange={(event) => setPostText(event.target.value)}
            rows={5}
            placeholder={editingPostId ? 'Edit your caption...' : 'What would you like to share with your community?'}
            className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
          />

          {!editingPostId && (
            <div>
              <label className="mb-2 block text-sm font-semibold text-slate-700">Add media</label>
              <input
                type="file"
                accept="image/*,video/*,audio/*"
                multiple
                onChange={(event) => setSelectedFiles(Array.from(event.target.files ?? []))}
                className="block w-full rounded-2xl border border-dashed border-sky-200 bg-sky-50 p-3 text-sm text-slate-600 file:mr-3 file:rounded-full file:border-0 file:bg-sky-100 file:px-3 file:py-2 file:text-sm file:font-semibold file:text-sky-700"
              />
              {selectedFiles.length > 0 && (
                <div className="mt-2 text-xs text-slate-500">
                  {selectedFiles.length} file(s) selected
                </div>
              )}
            </div>
          )}

          {composerError && (
            <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
              {composerError}
            </div>
          )}

          {composerSuccess && (
            <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              {composerSuccess}
            </div>
          )}

          <button
            type="button"
            disabled={isPosting}
            onClick={() => void onPublish()}
            className="flex w-full items-center justify-center rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-3.5 text-base font-semibold text-white shadow-lg shadow-sky-300/40 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-70"
          >
            {isPosting ? (editingPostId ? 'Updating...' : 'Posting...') : editingPostId ? 'Update Post' : 'Create Post'}
          </button>
        </div>
      </div>
    </div>
  );
}
