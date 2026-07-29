/**
 * A Discord OAuth profile object (returned during login).
 * https://github.com/nextauthjs/next-auth/blob/main/packages/core/src/providers/discord.ts
 */
export interface DiscordProfile {
  id: string;
  username: string;
  avatar?: string;
  discriminator: string;
  locale?: string;
  mfa_enabled?: boolean;
}

/**
 * Full Discord profile returned by Discord's OAuth API.
 * https://discord.com/developers/docs/resources/user#get-current-user
 *     scope: 'guilds guilds.members.read email identify'
 */
// WIP, not finished yet, need full refactor if I'm going to be using this,
// idk if I even want to use this need to research more on this.
export interface FullDiscordProfile {
  id: string;
  username: string;
  discriminator: string;
  avatar: string | null;
  public_flags: number;
  flags: number;
  banner: string | null;
  accent_color: number | null;
  global_name: string;
  avatar_decoration_data: {
    asset: string;
    sku_id: string;
    expires_at: string | null;
  } | null;
  collectibles: unknown | null;
  banner_color: string | null;
  clan: {
    identity_guild_id: string;
    identity_enabled: boolean;
    tag: string;
    badge: string;
  } | null;
  primary_guild: {
    identity_guild_id: string;
    identity_enabled: boolean;
    tag: string;
    badge: string;
  } | null;
  mfa_enabled: boolean;
  locale: string;
  premium_type: number;
  verified: boolean;
  image_url: string;
}
