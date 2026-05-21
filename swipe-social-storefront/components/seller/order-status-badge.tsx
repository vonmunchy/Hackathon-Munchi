'use client'

interface OrderStatusBadgeProps {
  status: string
}

const statusColors: Record<string, string> = {
  paid: 'bg-amethyst-100 text-amethyst-700',
  shipped: 'bg-blue-100 text-blue-700',
  delivered: 'bg-slate-100 text-slate-600',
  pending: 'bg-ruby-100 text-ruby-700',
  cancelled: 'bg-red-50 text-red-600',
}

export function OrderStatusBadge({ status }: OrderStatusBadgeProps) {
  const colors = statusColors[status] ?? 'bg-slate-100 text-slate-600'

  return (
    <span className={`inline-block rounded-full px-3 py-1 text-xs font-medium capitalize ${colors}`}>
      {status}
    </span>
  )
}
