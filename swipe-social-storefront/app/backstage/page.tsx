"use client";

import { useState, useCallback } from "react";

interface StoreInfo {
  _id: string;
  name: string;
  slug: string;
  sellerType: string | null;
  onboardingComplete: boolean;
  hasSwipeCredentials: boolean;
  contactPhone: string | null;
}

export default function BackstagePage() {
  const [passphrase, setPassphrase] = useState("");
  const [authed, setAuthed] = useState(false);
  const [authError, setAuthError] = useState("");
  const [storedPassphrase, setStoredPassphrase] = useState("");
  const [stores, setStores] = useState<StoreInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [resettingId, setResettingId] = useState<string | null>(null);
  const [confirmResetId, setConfirmResetId] = useState<string | null>(null);
  const [resetSuccess, setResetSuccess] = useState<string | null>(null);

  const fetchStores = useCallback(async (pass: string) => {
    setLoading(true);
    try {
      const res = await fetch("/api/admin/stores", {
        headers: { Authorization: pass },
      });
      if (res.status === 401) {
        setAuthError("Incorrect passphrase");
        setPassphrase("");
        setLoading(false);
        return false;
      }
      if (!res.ok) throw new Error("Failed to fetch");
      const data = await res.json();
      setStores(data);
      setAuthed(true);
      setStoredPassphrase(pass);
      setAuthError("");
      return true;
    } catch {
      setAuthError("Connection error");
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    await fetchStores(passphrase);
  };

  const handleReset = async (storeId: string) => {
    setResettingId(storeId);
    setResetSuccess(null);
    try {
      const res = await fetch(`/api/admin/stores/${storeId}/reset`, {
        method: "POST",
        headers: { Authorization: storedPassphrase },
      });
      if (!res.ok) throw new Error("Reset failed");
      const data = await res.json();
      setResetSuccess(data.storeName || storeId);
      setConfirmResetId(null);
      // Refresh the list
      await fetchStores(storedPassphrase);
    } catch {
      setAuthError("Reset failed — try again");
    } finally {
      setResettingId(null);
    }
  };

  // Passphrase gate
  if (!authed) {
    return (
      <div className="flex min-h-dvh flex-col items-center justify-center px-4">
        <div className="w-full max-w-sm">
          <div className="mb-8 text-center">
            <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-amethyst-500/20">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-7 w-7 text-amethyst-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z" />
              </svg>
            </div>
            <h1 className="font-display text-xl font-bold text-white">
              Backstage
            </h1>
            <p className="mt-1 text-sm text-slate-400">
              Admin access required
            </p>
          </div>

          <form onSubmit={handleLogin} className="space-y-4">
            <input
              type="password"
              value={passphrase}
              onChange={(e) => {
                setPassphrase(e.target.value);
                setAuthError("");
              }}
              placeholder="Enter passphrase"
              autoFocus
              className="w-full rounded-xl border border-slate-700 bg-slate-800 px-4 py-3 text-sm text-white placeholder:text-slate-500 focus:border-amethyst-500 focus:outline-none focus:ring-2 focus:ring-amethyst-500/20"
            />
            {authError && (
              <p className="text-center text-sm font-medium text-red-400">
                {authError}
              </p>
            )}
            <button
              type="submit"
              disabled={!passphrase || loading}
              className="w-full rounded-xl bg-amethyst-500 py-3 text-sm font-semibold text-white transition-colors hover:bg-amethyst-600 disabled:cursor-not-allowed disabled:bg-slate-700 disabled:text-slate-500"
            >
              {loading ? "Checking..." : "Enter"}
            </button>
          </form>
        </div>
      </div>
    );
  }

  // Admin panel
  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="font-display text-xl font-bold text-white">
            Backstage — Seller Management
          </h1>
          <p className="mt-1 text-sm text-slate-400">
            {stores.length} registered seller{stores.length !== 1 ? "s" : ""}
          </p>
        </div>
        <button
          onClick={() => fetchStores(storedPassphrase)}
          className="rounded-lg bg-slate-800 px-3 py-2 text-xs font-medium text-slate-300 transition-colors hover:bg-slate-700"
        >
          Refresh
        </button>
      </div>

      {resetSuccess && (
        <div className="mb-4 flex items-center gap-2 rounded-lg border border-emerald-800 bg-emerald-900/50 px-4 py-3">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
          </svg>
          <p className="text-sm text-emerald-300">
            <span className="font-semibold">{resetSuccess}</span> has been reset — they will go through onboarding on next login
          </p>
        </div>
      )}

      {stores.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-slate-700 bg-slate-800 py-16 text-center">
          <p className="text-sm text-slate-400">No sellers registered yet</p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-slate-700 bg-slate-800">
          <table className="w-full min-w-[640px]">
            <thead>
              <tr className="border-b border-slate-700 bg-slate-800/80">
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Store
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Slug
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Type
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Swipe
                </th>
                <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-slate-400">
                  Action
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-700/50">
              {stores.map((store) => (
                <tr key={store._id} className="transition-colors hover:bg-slate-700/30">
                  <td className="px-4 py-3 text-sm font-medium text-white">
                    {store.name}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">
                    {store.slug}
                  </td>
                  <td className="px-4 py-3 text-sm text-slate-300">
                    {store.sellerType ?? "—"}
                  </td>
                  <td className="px-4 py-3">
                    {store.onboardingComplete ? (
                      <span className="inline-flex items-center rounded-full bg-emerald-900/50 px-2.5 py-0.5 text-xs font-medium text-emerald-400">
                        Complete
                      </span>
                    ) : (
                      <span className="inline-flex items-center rounded-full bg-slate-700 px-2.5 py-0.5 text-xs font-medium text-slate-400">
                        Incomplete
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {store.hasSwipeCredentials ? (
                      <span className="inline-flex items-center rounded-full bg-amethyst-900/50 px-2.5 py-0.5 text-xs font-medium text-amethyst-400">
                        Connected
                      </span>
                    ) : (
                      <span className="inline-flex items-center rounded-full bg-slate-700 px-2.5 py-0.5 text-xs font-medium text-slate-500">
                        None
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {confirmResetId === store._id ? (
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => setConfirmResetId(null)}
                          className="rounded-lg bg-slate-700 px-3 py-1.5 text-xs font-medium text-slate-300 hover:bg-slate-600"
                        >
                          Cancel
                        </button>
                        <button
                          onClick={() => handleReset(store._id)}
                          disabled={resettingId === store._id}
                          className="rounded-lg bg-ruby-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-ruby-700 disabled:opacity-50"
                        >
                          {resettingId === store._id ? "Resetting..." : "Confirm Reset"}
                        </button>
                      </div>
                    ) : (
                      <button
                        onClick={() => {
                          setConfirmResetId(store._id);
                          setResetSuccess(null);
                        }}
                        className="rounded-lg border border-ruby-700 bg-ruby-900/30 px-3 py-1.5 text-xs font-medium text-ruby-400 transition-colors hover:bg-ruby-900/60"
                      >
                        Reset
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
