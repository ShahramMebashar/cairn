// Saved board views — named filter presets persisted per workspace in localStorage.

export type SavedView = {
  name: string;
  filter: string; // base filter (all/active/backlog/ready)
  query?: string;
  label?: string;
  assignee?: string;
  priority?: string;
};

const key = (path: string) => `cairn-views:${path}`;

export function loadViews(path: string): SavedView[] {
  try {
    return JSON.parse(localStorage.getItem(key(path)) || "[]");
  } catch {
    return [];
  }
}

function save(path: string, views: SavedView[]) {
  localStorage.setItem(key(path), JSON.stringify(views));
}

export function addView(path: string, view: SavedView): SavedView[] {
  const next = [...loadViews(path).filter((v) => v.name !== view.name), view];
  save(path, next);
  return next;
}

export function removeView(path: string, name: string): SavedView[] {
  const next = loadViews(path).filter((v) => v.name !== name);
  save(path, next);
  return next;
}
