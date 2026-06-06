import { Download, Mail } from 'lucide-react'
import GitHubIcon from '../components/icons/GitHubIcon'
import {
  useEducation,
  useExperience,
  useProfile,
  useProjects,
  useSkills,
} from '../queries/usePortfolio'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export default function PortfolioPage() {
  useDocumentTitle('Portfolio')

  const { data: profile, isLoading: loadingProfile } = useProfile()
  const { data: experience, isLoading: loadingExperience } = useExperience()
  const { data: education, isLoading: loadingEducation } = useEducation()
  const {
    data: projects,
    isLoading: loadingProjects,
    error: projectsError,
  } = useProjects()
  const { data: skills, isLoading: loadingSkills } = useSkills()

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      {/* ── Hero ── */}
      <section className="animate-fade-in mb-14">
        {loadingProfile ? (
          <p className="text-gray-500">Loading…</p>
        ) : profile ? (
          <>
            <h1 className="mb-1 text-4xl font-bold text-gray-100">
              {profile.name}
            </h1>
            <p className="mb-4 text-lg font-medium text-indigo-400">
              {profile.title}
            </p>
            <p className="mb-6 max-w-2xl text-gray-400">{profile.summary}</p>

            {/* Contact + CV download */}
            <div className="flex flex-wrap items-center gap-3">
              <a
                href={profile.github}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-4 py-2 text-sm text-gray-300 transition-colors hover:border-indigo-700 hover:text-gray-100"
              >
                <GitHubIcon className="h-4 w-4" />
                GitHub
              </a>
              <a
                href={`mailto:${profile.email}`}
                className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-4 py-2 text-sm text-gray-300 transition-colors hover:border-indigo-700 hover:text-gray-100"
              >
                <Mail className="h-4 w-4" strokeWidth={1.5} />
                {profile.email}
              </a>
              <a
                href="/documents/CV_THANAWAT_YUWANSIRI.pdf"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-500"
              >
                <Download className="h-4 w-4" strokeWidth={1.5} />
                Download CV
              </a>
            </div>
          </>
        ) : null}
      </section>

      {/* ── Work Experience ── */}
      <section className="mb-12">
        <h2 className="mb-5 text-xl font-semibold text-gray-200">
          Work Experience
        </h2>
        {loadingExperience ? (
          <p className="text-gray-500">Loading…</p>
        ) : !experience?.length ? (
          <p className="text-gray-500">No experience listed.</p>
        ) : (
          <div className="flex flex-col gap-5">
            {experience.map((exp) => (
              <div
                key={exp.company}
                className="animate-fade-in rounded-xl border border-gray-800 bg-gray-900 p-6 transition-all hover:-translate-y-0.5 hover:border-indigo-700"
              >
                <div className="mb-4 flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <h3 className="font-semibold text-gray-100">{exp.role}</h3>
                    <p className="text-sm text-indigo-400">{exp.company}</p>
                  </div>
                  <span className="shrink-0 rounded-full bg-gray-800 px-3 py-0.5 text-xs text-gray-400">
                    {exp.period}
                  </span>
                </div>
                <ul className="space-y-1.5">
                  {exp.highlights.map((point, i) => (
                    <li key={i} className="flex gap-2 text-sm text-gray-400">
                      <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-indigo-500" />
                      {point}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* ── Education ── */}
      <section className="mb-12">
        <h2 className="mb-5 text-xl font-semibold text-gray-200">Education</h2>
        {loadingEducation ? (
          <p className="text-gray-500">Loading…</p>
        ) : !education?.length ? (
          <p className="text-gray-500">No education listed.</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
            {education.map((edu) => (
              <div
                key={edu.school}
                className="animate-fade-in rounded-xl border border-gray-800 bg-gray-900 p-5 transition-all hover:-translate-y-0.5 hover:border-indigo-700"
              >
                <div className="mb-1 flex flex-wrap items-start justify-between gap-2">
                  <h3 className="font-semibold text-gray-100">{edu.school}</h3>
                  <span className="shrink-0 rounded-full bg-gray-800 px-3 py-0.5 text-xs text-gray-400">
                    {edu.period}
                  </span>
                </div>
                <p className="text-sm text-indigo-400">{edu.degree}</p>
                <p className="mt-1 text-sm text-gray-500">{edu.detail}</p>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* ── Skills ── */}
      <section className="mb-12">
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
                className="animate-fade-in rounded-xl border border-gray-800 bg-gray-900 p-5"
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

      {/* ── Projects ── */}
      <section>
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
                className="animate-fade-in rounded-xl border border-gray-800 bg-gray-900 p-5 transition-all hover:-translate-y-0.5 hover:border-indigo-700"
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
    </main>
  )
}
