'use client'

interface DashboardCardProps {
  title: string
  value: string
  subtitle?: string
  accentColor?: string
}

export function DashboardCard({ title, value, subtitle, accentColor = 'text-amethyst-700' }: DashboardCardProps) {
  return (
    <div className="bg-white rounded-xl border border-slate-100 shadow-sm p-5">
      <p className="text-xs font-medium uppercase tracking-wide text-slate-400">{title}</p>
      <p className={`text-2xl font-semibold font-mono mt-2 ${accentColor}`}>{value}</p>
      {subtitle && <p className="text-xs text-slate-400 mt-1.5">{subtitle}</p>}
    </div>
  )
}
