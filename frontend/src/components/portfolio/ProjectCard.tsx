import { Card } from '@/components/ui/Card';
import type { Project } from '@/types/portfolio';

interface Props {
  project: Project;
}

export function ProjectCard({ project }: Props) {
  return (
    <Card>
      <h3 className="text-lg font-semibold text-gray-900">{project.name}</h3>
      <p className="mt-1 text-sm text-gray-600">{project.description}</p>
      <div className="mt-3 flex flex-wrap gap-2">
        {project.tags.map((tag) => (
          <span key={tag} className="rounded-full bg-blue-50 px-2.5 py-0.5 text-xs font-medium text-blue-700">
            {tag}
          </span>
        ))}
      </div>
      {project.url && (
        <a href={project.url} target="_blank" rel="noopener noreferrer" className="mt-3 inline-block text-sm text-blue-600 hover:underline">
          View project &rarr;
        </a>
      )}
    </Card>
  );
}
