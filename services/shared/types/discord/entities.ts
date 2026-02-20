/**
 * A minimal Discord user object (used for avatars or mentions).
 */
export interface DiscordUser {
  id: string;
  avatar?: string;
  discriminator?: string;
}

/**
 * A role returned from the Discord API.
 */
export interface DiscordRole {
  id: string;
  name: string;
  color: number;
}

/**
 * A session user passed to the client.
 * Derived from JWT and enriched with role info.
 */
export interface DiscordSessionUser {
  id: string;
  name?: string;
  discriminator?: string;
  avatar?: string;
  roles?: (string | DiscordRole)[];
  cacheExpires?: number | null;
}

/**
 * A Discord member object from the Discord API.
 */
export interface DiscordMember {
  user: {
    id: string;
    username: string;
    avatar?: string;
  };
  roles: string[];
  additionalFields?: Record<string, unknown>; // Can be typed as needed
}

/**
 * Cached Discord member enriched with full role objects.
 */
export type CachedDiscordMember = DiscordMember & {
  enrichedRoles: DiscordRole[];
};

/**
 * Result of resolving a Discord member from API/cache.
 */
export type ResolvedDiscordMember =
  | { data: CachedDiscordMember; ttl: number }
  | { data: null; ttl: null };

/**
 * Simplified user object used internally for permission checks.
 */
export type UserWithRoles = {
  id: string;
  roles: (string | DiscordRole)[];
};

/**
 * Convert a raw or enriched DiscordMember into a normalized UserWithRoles object.
 */
export function toUserWithRoles(
  member: DiscordMember | CachedDiscordMember | null | undefined,
): UserWithRoles | null {
  if (!member) return null;

  const roleSet = new Set<string>();
  const enrichedRoles: DiscordRole[] = [];

  if (Array.isArray(member.roles)) {
    member.roles.forEach((r) => {
      if (typeof r === "string") roleSet.add(r);
    });
  }

  if ("enrichedRoles" in member && Array.isArray(member.enrichedRoles)) {
    member.enrichedRoles.forEach((r) => {
      roleSet.add(r.id);
      enrichedRoles.push(r);
    });
  }

  return {
    id: member.user.id,
    roles: [...roleSet, ...enrichedRoles],
  };
}
