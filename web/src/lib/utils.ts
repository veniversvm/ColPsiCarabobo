// web/src/lib/utils.ts
// Utilidades
import { PostGrade } from "~/types/psi";

export function sortPostGradesByYear(postGrades: PostGrade[] | undefined): PostGrade[] {
  if (!postGrades) return [];
  return [...postGrades].sort((a, b) => {
    if (!a.year) return 1;
    if (!b.year) return -1;
    return parseInt(a.year) - parseInt(b.year);
  });
}