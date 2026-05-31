'use client';

import { type FormEvent, useState } from 'react';

import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import type { CreateCardInput } from '@/types/kanban';

interface Props {
  columnId: string;
  onAdd: (input: CreateCardInput) => Promise<void>;
}

export function AddCardForm({ columnId, onAdd }: Props) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    setLoading(true);
    try {
      await onAdd({ column_id: columnId, title: title.trim(), content: content.trim() });
      setTitle('');
      setContent('');
      setOpen(false);
    } finally {
      setLoading(false);
    }
  }

  if (!open) {
    return (
      <button onClick={() => setOpen(true)} className="w-full rounded-lg px-3 py-1.5 text-left text-sm text-gray-500 hover:bg-gray-100">
        + Add card
      </button>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <Input placeholder="Card title" value={title} onChange={(e) => setTitle(e.target.value)} autoFocus />
      <Input placeholder="Description (optional)" value={content} onChange={(e) => setContent(e.target.value)} />
      <div className="flex gap-2">
        <Button type="submit" className="text-xs" disabled={loading}>
          Add
        </Button>
        <Button type="button" variant="secondary" className="text-xs" onClick={() => setOpen(false)}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
