// web/src/components/admin/noticias/edit/EditSkeleton.tsx
export function EditSkeleton() {
  return (
    <div class="space-y-6 animate-pulse">
      <div class="bg-white rounded-3xl h-64 border border-gray-100" />
      <div class="bg-white rounded-3xl h-44 border border-gray-100" />
      <div class="bg-white rounded-3xl h-96 border border-gray-100" />
    </div>
  );
}