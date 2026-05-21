"use client";

import { useState, type FormEvent } from "react";
import { validateName, validatePhone, validateAddress } from "@/lib/validators";

export interface DeliveryData {
  customerName: string;
  customerPhone: string;
  deliveryLocation: string;
  deliveryAddress: string;
  deliveryTimePreference?: string;
}

interface DeliveryFormProps {
  onSubmit: (data: DeliveryData) => void;
  isLoading: boolean;
}

export function DeliveryForm({ onSubmit, isLoading }: DeliveryFormProps) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [location, setLocation] = useState("Male'");
  const [address, setAddress] = useState("");
  const [timePreference, setTimePreference] = useState("");

  const [touched, setTouched] = useState({
    name: false,
    phone: false,
    address: false,
  });

  const nameResult = validateName(name);
  const phoneResult = validatePhone(phone);
  const addressResult = validateAddress(address);

  const isValid = nameResult.valid && phoneResult.valid && addressResult.valid;

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setTouched({ name: true, phone: true, address: true });
    if (!isValid) return;

    onSubmit({
      customerName: name.trim(),
      customerPhone: phone.replace(/\D/g, ""),
      deliveryLocation: location,
      deliveryAddress: address.trim(),
      deliveryTimePreference: timePreference.trim() || undefined,
    });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Name */}
      <div>
        <label className="mb-1 block text-sm font-medium text-slate-700">
          Name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onBlur={() => setTouched((t) => ({ ...t, name: true }))}
          placeholder="Your full name"
          className="w-full rounded-lg border border-slate-200 px-3 py-2.5 text-slate-800 placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
        />
        {touched.name && !nameResult.valid && (
          <p className="mt-1 text-sm text-error">{nameResult.error}</p>
        )}
      </div>

      {/* Phone */}
      <div>
        <label className="mb-1 block text-sm font-medium text-slate-700">
          Phone
        </label>
        <input
          type="tel"
          inputMode="numeric"
          value={phone}
          onChange={(e) => setPhone(e.target.value.replace(/[^\d]/g, "").slice(0, 7))}
          onBlur={() => setTouched((t) => ({ ...t, phone: true }))}
          placeholder="7XXXXXX or 9XXXXXX"
          className="w-full rounded-lg border border-slate-200 px-3 py-2.5 font-mono text-slate-800 placeholder:font-sans placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
        />
        {touched.phone && !phoneResult.valid && (
          <p className="mt-1 text-sm text-error">{phoneResult.error}</p>
        )}
      </div>

      {/* Location */}
      <div>
        <label className="mb-1 block text-sm font-medium text-slate-700">
          Delivery Location
        </label>
        <div className="flex gap-4">
          {["Male'", "Hulhumale'"].map((loc) => (
            <label key={loc} className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="location"
                value={loc}
                checked={location === loc}
                onChange={() => setLocation(loc)}
                className="h-4 w-4 accent-amethyst-500"
              />
              <span className="text-sm text-slate-700">{loc}</span>
            </label>
          ))}
        </div>
      </div>

      {/* Address */}
      <div>
        <label className="mb-1 block text-sm font-medium text-slate-700">
          Address
        </label>
        <textarea
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          onBlur={() => setTouched((t) => ({ ...t, address: true }))}
          placeholder="Building name, floor, road..."
          rows={3}
          className="w-full rounded-lg border border-slate-200 px-3 py-2.5 text-slate-800 placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
        />
        {touched.address && !addressResult.valid && (
          <p className="mt-1 text-sm text-error">{addressResult.error}</p>
        )}
      </div>

      {/* Time preference */}
      <div>
        <label className="mb-1 block text-sm font-medium text-slate-700">
          Preferred Delivery Time{" "}
          <span className="text-slate-400">(optional)</span>
        </label>
        <input
          type="text"
          value={timePreference}
          onChange={(e) => setTimePreference(e.target.value)}
          placeholder="e.g. After 5pm, Weekend morning"
          className="w-full rounded-lg border border-slate-200 px-3 py-2.5 text-slate-800 placeholder:text-slate-400 focus:border-amethyst-400 focus:outline-none focus:ring-2 focus:ring-amethyst-100"
        />
      </div>

      {/* Submit */}
      <button
        type="submit"
        disabled={!isValid || isLoading}
        className="w-full rounded-xl bg-ruby-500 py-3 text-base font-semibold text-white transition-colors hover:bg-ruby-600 disabled:cursor-not-allowed disabled:bg-slate-300"
      >
        {isLoading ? "Placing Order..." : "Place Order"}
      </button>
    </form>
  );
}
