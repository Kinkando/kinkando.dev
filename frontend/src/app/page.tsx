'use client';

import { useEffect, useState } from 'react';

import { Navbar } from '@/components/layout/Navbar';
import { ProjectCard } from '@/components/portfolio/ProjectCard';
import { SkillList } from '@/components/portfolio/SkillList';
import { Spinner } from '@/components/ui/Spinner';
import { getProjects, getSkills } from '@/lib/api/portfolio';
import type { Project } from '@/types/portfolio';
import type { SkillCategory } from '@/types/portfolio';

export default function HomePage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [skills, setSkills] = useState<SkillCategory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getProjects(), getSkills()])
      .then(([p, s]) => {
        setProjects(p);
        setSkills(s);
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-white">
      <Navbar />

      <main className="mx-auto max-w-6xl px-4 py-12">
        <section className="mb-16 text-center">
          <h1 className="text-4xl font-bold text-gray-900">kinkando.dev</h1>
          <p className="mt-3 text-lg text-gray-600">Personal dashboard — portfolio, finance tracker, and kanban board.</p>
        </section>

        {loading ? (
          <div className="flex justify-center py-12">
            <Spinner className="h-8 w-8" />
          </div>
        ) : (
          <>
            <section className="mb-12">
              <h2 className="mb-6 text-2xl font-bold text-gray-900">Projects</h2>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                {projects.map((p) => (
                  <ProjectCard key={p.name} project={p} />
                ))}
              </div>
            </section>

            <section>
              <h2 className="mb-6 text-2xl font-bold text-gray-900">Skills</h2>
              <SkillList skills={skills} />
            </section>
          </>
        )}
      </main>
    </div>
  );
}
