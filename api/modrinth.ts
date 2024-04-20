import { fetchJSONData } from "../util.ts";

export interface Project {
  client_side: string;
  server_side: string;
  game_versions: string[];
  id: string;
  slug: string;
  project_type: string;
  team: string;
  organization: any;
  title: string;
  description: string;
  body: string;
  body_url: any;
  published: string;
  updated: string;
  approved: string;
  queued: any;
  status: string;
  requested_status: any;
  moderator_message: any;
  license: {
    id: string;
    name: string;
    url: any;
  };
  downloads: number;
  followers: number;
  categories: string[];
  additional_categories: any[];
  loaders: string[];
  versions: string[];
  icon_url: string;
  issues_url: string;
  source_url: string;
  wiki_url: string;
  discord_url: string;
  donation_urls: any[];
  gallery: any[];
  color: number;
  thread_id: string;
  monetization_status: string;
}

export interface ProjectVersion {
  game_versions: string[];
  loaders: string[];
  id: string;
  project_id: string;
  author_id: string;
  featured: boolean;
  name: string;
  version_number: string;
  changelog: string;
  changelog_url: any;
  date_published: string;
  downloads: number;
  version_type: string;
  status: string;
  requested_status: any;
  files: {
    hashes: {
      sha1: string;
      sha512: string;
    };
    url: string;
    filename: string;
    primary: boolean;
    size: number;
    file_type: any;
  }[];
  dependencies: any[];
}

export async function search(q: string) {
  const results: { hits: ProjectVersion[] } = await fetchJSONData(
    `https://api.modrinth.com/v2/search?query=${q}`,
  );
  return results;
}

export async function getProject(id: string) {
  const project: Project = await fetchJSONData(
    `https://api.modrinth.com/v2/project/${id}`,
  );
  return project;
}
export async function getProjectVersion(id: string, gameVersion: string) {
  const versions: ProjectVersion[] = await fetchJSONData(
    `https://api.modrinth.com/v2/project/${id}/version`,
  );
  const version = versions.find((element) =>
    element.game_versions.includes(gameVersion),
  );
  if (!version) {
    throw Error(
      "Could not find a project version for the specified game version",
    );
  }
  return version;
}
