import { Card } from '@/components/ui/Card';
import type { SkillCategory } from '@/types/portfolio';

interface Props {
  skills: SkillCategory[];
}

export function SkillList({ skills }: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
      {skills.map((cat) => (
        <Card key={cat.category}>
          <h3 className="mb-2 text-sm font-semibold text-gray-700">{cat.category}</h3>
          <div className="flex flex-wrap gap-2">
            {cat.items.map((item) => (
              <span key={item} className="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700">
                {item}
              </span>
            ))}
          </div>
        </Card>
      ))}
    </div>
  );
}
