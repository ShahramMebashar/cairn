// Workspaces map a clean URL slug (the folder name) to its absolute path, kept in
// localStorage. This lets the URL stay readable — #/cairn/task/ACME-004 — while the full
// path (which is machine-specific and ugly) is resolved locally.

const REGISTRY_KEY = "cairn-workspaces";
const LAST_KEY = "cairn-current-folder";

type Registry = Record<string, string>; // slug -> absolute path

function load(): Registry {
  try {
    return JSON.parse(localStorage.getItem(REGISTRY_KEY) || "{}");
  } catch {
    return {};
  }
}

function save(reg: Registry) {
  localStorage.setItem(REGISTRY_KEY, JSON.stringify(reg));
}

/** slugify turns a path into a clean URL token from its last segment. */
export function slugify(path: string): string {
  const base = path.split("/").filter(Boolean).pop() || "project";
  return base.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "") || "project";
}

/** registerWorkspace returns a stable slug for a path, minting a unique one if needed. */
export function registerWorkspace(path: string): string {
  const reg = load();
  for (const [slug, p] of Object.entries(reg)) {
    if (p === path) {
      localStorage.setItem(LAST_KEY, path);
      return slug;
    }
  }
  const base = slugify(path);
  let slug = base;
  let n = 2;
  while (reg[slug] && reg[slug] !== path) slug = `${base}-${n++}`;
  reg[slug] = path;
  save(reg);
  localStorage.setItem(LAST_KEY, path);
  return slug;
}

/** resolveSlug returns the path for a slug, or null if unknown on this machine. */
export function resolveSlug(slug: string): string | null {
  return load()[slug] ?? null;
}

/** listWorkspaces returns every registered workspace (for the quick-capture switcher). */
export function listWorkspaces(): { slug: string; path: string }[] {
  return Object.entries(load()).map(([slug, path]) => ({ slug, path }));
}

/** lastWorkspace returns the most recently opened workspace, if any. */
export function lastWorkspace(): { slug: string; path: string } | null {
  const path = localStorage.getItem(LAST_KEY);
  if (!path) return null;
  return { slug: registerWorkspace(path), path };
}

/** forget removes a workspace from the registry (used when switching folders). */
export function forget(path: string) {
  const reg = load();
  for (const [slug, p] of Object.entries(reg)) if (p === path) delete reg[slug];
  save(reg);
  if (localStorage.getItem(LAST_KEY) === path) localStorage.removeItem(LAST_KEY);
}
