import { useState } from 'react'
import type { Medicine } from '../../lib/api/types'
import { useMedicines } from '../../queries/useMedicine'
import MedicineCard from './MedicineCard'
import MedicineFormDialog from './MedicineFormDialog'

export default function MedicinesTab() {
  const [includeArchived, setIncludeArchived] = useState(false)
  const [showAdd, setShowAdd] = useState(false)

  const { data: medicines, isLoading } = useMedicines(includeArchived)

  const active = medicines?.filter((m: Medicine) => !m.archived_at) ?? []
  const archived =
    medicines?.filter((m: Medicine) => m.archived_at != null) ?? []

  return (
    <div className="space-y-6">
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-4">
        <label className="flex cursor-pointer items-center gap-2 text-sm text-gray-400">
          <input
            type="checkbox"
            checked={includeArchived}
            onChange={(e) => setIncludeArchived(e.target.checked)}
            className="cursor-pointer rounded border-gray-600 bg-gray-800 text-indigo-600"
          />
          Show archived
        </label>
        <button
          onClick={() => setShowAdd(true)}
          className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
        >
          + Add Medicine
        </button>
      </div>

      {/* Loading */}
      {isLoading && (
        <p className="py-8 text-center text-sm text-gray-500">Loading…</p>
      )}

      {/* Active medicines */}
      {!isLoading && active.length === 0 && !includeArchived && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 py-12 text-center">
          <p className="text-sm text-gray-500">No medicines yet.</p>
          <button
            onClick={() => setShowAdd(true)}
            className="mt-3 cursor-pointer text-sm text-indigo-400 hover:text-indigo-300"
          >
            Add your first medicine
          </button>
        </div>
      )}

      {active.length > 0 && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {active.map((m: Medicine) => (
            <MedicineCard key={m.id} medicine={m} />
          ))}
        </div>
      )}

      {/* Archived medicines */}
      {includeArchived && archived.length > 0 && (
        <div>
          <h3 className="mb-3 text-xs font-medium tracking-wider text-gray-600 uppercase">
            Archived
          </h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {archived.map((m: Medicine) => (
              <MedicineCard key={m.id} medicine={m} />
            ))}
          </div>
        </div>
      )}

      {showAdd && <MedicineFormDialog onClose={() => setShowAdd(false)} />}
    </div>
  )
}
