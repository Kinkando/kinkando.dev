import { useProjects, useSkills } from '../queries/usePortfolio'

export default function PortfolioPage() {
  const {
    data: projects,
    isLoading: loadingProjects,
    error: projectsError,
  } = useProjects()
  const { data: skills, isLoading: loadingSkills } = useSkills()

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-2 text-3xl font-bold text-gray-100">Portfolio</h1>
      <p className="mb-10 text-gray-400">Projects and skills.</p>

      <section className="mb-12">
        <h2 className="mb-5 text-xl font-semibold text-gray-200">Projects</h2>
        {loadingProjects ? (
          <p className="text-gray-500">Loading…</p>
        ) : projectsError ? (
          <p className="text-sm text-red-400">Failed to load projects.</p>
        ) : !projects?.length ? (
          <p className="text-gray-500">No projects yet.</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <a
                key={project.name}
                href={project.url}
                target="_blank"
                rel="noopener noreferrer"
                className="rounded-xl border border-gray-800 bg-gray-900 p-5 transition-colors hover:border-indigo-700"
              >
                <h3 className="mb-1 font-semibold text-gray-100">
                  {project.name}
                </h3>
                <p className="mb-3 text-sm text-gray-400">
                  {project.description}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {project.tags.map((tag) => (
                    <span
                      key={tag}
                      className="rounded-full bg-indigo-950 px-2 py-0.5 text-xs text-indigo-300"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </a>
            ))}
          </div>
        )}
      </section>

      <section>
        <h2 className="mb-5 text-xl font-semibold text-gray-200">Skills</h2>
        {loadingSkills ? (
          <p className="text-gray-500">Loading…</p>
        ) : !skills?.length ? (
          <p className="text-gray-500">No skills listed.</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {skills.map((skill) => (
              <div
                key={skill.category}
                className="rounded-xl border border-gray-800 bg-gray-900 p-5"
              >
                <h3 className="mb-3 text-sm font-semibold tracking-wider text-indigo-400 uppercase">
                  {skill.category}
                </h3>
                <div className="flex flex-wrap gap-1.5">
                  {skill.items.map((item) => (
                    <span
                      key={item}
                      className="rounded-full bg-gray-800 px-2 py-0.5 text-xs text-gray-300"
                    >
                      {item}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </main>
  )
}
