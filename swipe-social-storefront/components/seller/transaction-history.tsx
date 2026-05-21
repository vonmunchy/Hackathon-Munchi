'use client'

import { formatMVR, formatMaldivesTime } from '@/lib/format'
import { useIsMobile } from '@/lib/use-device'

interface Transaction {
  id: string
  reference: string
  amount: number
  currency: string
  status: string
  description: string
  gross_amount: number
  fee_amount: number
  net_amount: number
  created_at: string
}

interface TransactionHistoryProps {
  transactions: Transaction[]
  loading: boolean
}

function StatusBadge({ status }: { status: string }) {
  const colors = status === 'COMPLETED'
    ? 'bg-amethyst-100 text-amethyst-700'
    : status === 'PENDING'
    ? 'bg-ruby-100 text-ruby-700'
    : 'bg-slate-100 text-slate-600'

  return (
    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${colors}`}>
      {status}
    </span>
  )
}

export function TransactionHistory({ transactions, loading }: TransactionHistoryProps) {
  const isMobile = useIsMobile()

  if (loading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-slate-100 rounded-lg h-20 animate-pulse" />
        ))}
      </div>
    )
  }

  if (transactions.length === 0) {
    return (
      <div className="text-center py-8 text-slate-400">
        <p className="text-lg">No transactions yet</p>
        <p className="text-sm mt-1">Transactions will appear here after your first sale</p>
      </div>
    )
  }

  if (isMobile) {
    return (
      <div className="space-y-3">
        {transactions.map((txn) => (
          <div key={txn.id} className="bg-white rounded-lg shadow-sm p-4">
            <div className="flex items-start justify-between">
              <div>
                <p className="font-mono text-xs text-slate-500">{txn.reference}</p>
                <p className="text-sm text-slate-700 mt-0.5">{txn.description}</p>
              </div>
              <StatusBadge status={txn.status} />
            </div>
            <div className="mt-3 flex items-end justify-between">
              <div className="text-xs text-slate-400">
                <p>{formatMVR(txn.gross_amount ?? txn.amount ?? 0)} - {formatMVR(txn.fee_amount ?? 0)} fee</p>
                <p className="text-slate-500 font-medium">Net: {formatMVR(txn.net_amount ?? txn.amount ?? 0)}</p>
              </div>
              <p className="font-mono text-lg font-semibold text-amethyst-700">
                {formatMVR(txn.amount)}
              </p>
            </div>
            <p className="text-xs text-slate-400 mt-2">
              {formatMaldivesTime(new Date(txn.created_at).getTime())}
            </p>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-200 text-left text-slate-500">
            <th className="pb-2 font-medium">Reference</th>
            <th className="pb-2 font-medium">Amount</th>
            <th className="pb-2 font-medium">Fee</th>
            <th className="pb-2 font-medium">Net</th>
            <th className="pb-2 font-medium">Status</th>
            <th className="pb-2 font-medium">Time</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {transactions.map((txn) => (
            <tr key={txn.id} className="hover:bg-slate-50">
              <td className="py-3 font-mono text-xs text-slate-600">{txn.reference}</td>
              <td className="py-3 font-mono font-medium text-amethyst-700">{formatMVR(txn.amount ?? 0)}</td>
              <td className="py-3 font-mono text-xs text-slate-400">{formatMVR(txn.fee_amount ?? 0)}</td>
              <td className="py-3 font-mono text-sm text-slate-700">{formatMVR(txn.net_amount ?? txn.amount ?? 0)}</td>
              <td className="py-3"><StatusBadge status={txn.status} /></td>
              <td className="py-3 text-xs text-slate-400">
                {formatMaldivesTime(new Date(txn.created_at).getTime())}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
