import Image from 'next/image';

interface AvatarProps {
  src?: string | null;
  name?: string | null;
  email?: string | null;
  size?: number;
  className?: string;
}

/**
 * Displays a circular avatar image, or an initials fallback when no photo URL
 * is available (e.g. email/password accounts that have no photoURL).
 *
 * referrerPolicy="no-referrer" is set on the image so Google avatar URLs
 * (lh3.googleusercontent.com) are not blocked by the referrer policy.
 * The project uses images.unoptimized: true globally so no domain allow-list is needed.
 */
export function Avatar({ src, name, email, size = 32, className = '' }: AvatarProps) {
  const initial = (name?.[0] ?? email?.[0] ?? '?').toUpperCase();

  if (src) {
    return (
      <Image
        src={src}
        alt={name ?? email ?? 'User avatar'}
        width={size}
        height={size}
        referrerPolicy="no-referrer"
        className={`rounded-full object-cover ${className}`}
        style={{ width: size, height: size }}
      />
    );
  }

  return (
    <span
      className={`inline-flex items-center justify-center rounded-full bg-gray-200 text-sm font-medium text-gray-600 ${className}`}
      style={{ width: size, height: size }}
      aria-label={name ?? email ?? 'User avatar'}
    >
      {initial}
    </span>
  );
}
