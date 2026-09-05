// web/src/types/projects.ts
// Tipos del módulo Proyectos (Kanban) del panel admin. Espejan los DTOs de la API Go.

export type ProjectMemberRole = "viewer" | "editor";
export type ProjectAccessLevel = "viewer" | "editor" | "owner" | "master";

export interface ProjectUser {
  id: string;
  username?: string;
  email?: string;
  first_name?: string;
  last_name?: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  member_count: number;
  card_count: number;
  my_role?: ProjectMemberRole;
  is_master: boolean;
  is_owner: boolean;
  owner?: ProjectUser;
  created_at: string;
  updated_at: string;
  create_by: string;
}

export interface ProjectMember {
  id: string;
  project_id: string;
  user_admin_id: string;
  role: ProjectMemberRole;
  user?: ProjectUser;
  created_at: string;
  create_by: string;
}

export interface ProjectNote {
  id: string;
  card_id: string;
  content: string;
  created_at: string;
  updated_at: string;
  create_by: string;
  create_by_id?: string | null;
}

export interface ProjectCard {
  id: string;
  project_id: string;
  column_id: string;
  title: string;
  description: string;
  position: number;
  notes?: ProjectNote[];
  created_at: string;
  updated_at: string;
  create_by: string;
}

export interface ProjectColumn {
  id: string;
  project_id: string;
  title: string;
  position: number;
  cards?: ProjectCard[];
  created_at: string;
  updated_at: string;
}

export interface ProjectBoard {
  project: Project;
  columns: ProjectColumn[];
}

/** Columna del tablero con sus tarjetas (vista local/DnD). */
export interface BoardColumn {
  id: string;
  project_id: string;
  title: string;
  position: number;
  cards?: BoardCard[];
}

/** Tarjeta del tablero (vista local/DnD). */
export interface BoardCard {
  id: string;
  project_id: string;
  column_id: string;
  title: string;
  description: string;
  position: number;
  notes?: ProjectNote[];
  created_at: string;
  updated_at: string;
  create_by: string;
}

/** Miembro del proyecto (vista local). */
export interface BoardMember {
  id: string;
  project_id: string;
  user_admin_id: string;
  role: ProjectMemberRole;
  user?: ProjectUser;
  created_at: string;
  create_by: string;
}

export const MAX_NOTES_PER_CARD = 10;
export const MAX_NOTE_LENGTH_CHARS = 500;

export function canEditProject(p: Pick<Project, "is_owner" | "is_master" | "my_role">): boolean {
  return p.is_owner || p.is_master || p.my_role === "editor";
}

export function canManageProject(p: Pick<Project, "is_owner" | "is_master">): boolean {
  return p.is_owner || p.is_master;
}