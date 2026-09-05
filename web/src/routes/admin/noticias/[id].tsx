// web/src/routes/admin/noticias/[id].tsx
import { createResource, createSignal, Show } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";
import { apiGet, apiPatch } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import {
  PostDetail,
  EditHeader,
  EditFeedback,
  EditSkeleton,
  EditNotFound,
  EditMetadataSection,
  EditImageSection,
  EditContentSection,
  EditActions,
  imgUrl
} from "~/components/admin/noticias/edit";

// ─────────────────────────────────────────────────────────────────────────────
export default function AdminEditarNoticiaPage() {
  const params   = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [post] = createResource(
    () => params.id,
    async (id) => {
      try {
        return await apiGet<PostDetail>(`/posts/${id}`);
      } catch (err: any) {
        console.error("[edit] error cargando post:", err?.status, err?.message);
        return null;
      }
    }
  );

  // Estado del formulario
  const [title,            setTitle]            = createSignal("");
  const [shortDescription, setShortDescription] = createSignal("");
  const [content,          setContent]          = createSignal("");
  const [type,             setType]             = createSignal<"public" | "psi">("public");
  const [status,           setStatus]           = createSignal<PostDetail["status"]>("draft");
  const [publishAt,        setPublishAt]        = createSignal("");
  const [imageFile,        setImageFile]        = createSignal<File | null>(null);
  const [imagePreview,     setImagePreview]     = createSignal<string | null>(null);
  const [initialized,      setInitialized]      = createSignal(false);

  const [saving,  setSaving]  = createSignal(false);
  const [error,   setError]   = createSignal<string | null>(null);
  const [success, setSuccess] = createSignal(false);

  const initForm = (p: PostDetail) => {
    if (initialized()) return;
    setTitle(p.title ?? "");
    setShortDescription(p.short_description ?? "");
    setContent(p.text?.content ?? "");
    setType(p.type ?? "public");
    setStatus(p.status ?? "draft");
    if (p.publish_at) {
      setPublishAt(new Date(p.publish_at).toISOString().slice(0, 16));
    }
    setInitialized(true);
  };

  const currentImage = () =>
    imagePreview() ?? (post()?.image_url ? imgUrl(post()!.image_url) : null);

  const handleImageChange = (e: Event) => {
    const file = (e.currentTarget as HTMLInputElement).files?.[0] ?? null;
    setImageFile(file);
    if (file) {
      const reader = new FileReader();
      reader.onload = (ev) => setImagePreview(ev.target?.result as string);
      reader.readAsDataURL(file);
    } else {
      setImagePreview(null);
    }
  };

  const clearImage = () => {
    setImageFile(null);
    setImagePreview(null);
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!title().trim())                                   { setError("El título es obligatorio."); return; }
    if (!content().trim() || content() === "<p></p>")     { setError("El contenido no puede estar vacío."); return; }
    if (status() === "scheduled" && !publishAt().trim())  { setError("Un post programado requiere fecha de publicación."); return; }

    setSaving(true);
    setError(null);
    setSuccess(false);

    try {
      const fd = new FormData();
      fd.append("title",             title().trim());
      fd.append("short_description", shortDescription().trim());
      fd.append("content",           content());
      fd.append("type",              type());
      fd.append("status",            status());
      if (status() === "scheduled" && publishAt()) {
        fd.append("publish_at", new Date(publishAt()).toISOString());
      }
      if (imageFile()) fd.append("image", imageFile()!);

      await apiPatch(`/admin/posts/${params.id}`, fd);

      setSuccess(true);
      window.scrollTo({ top: 0, behavior: "smooth" });
      setTimeout(() => navigate("/admin/noticias"), 1200);
    } catch (err: any) {
      setError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main class="pb-28 animate-in fade-in duration-500 max-w-4xl mx-auto">

      <EditHeader 
        post={post() ?? null}
        loading={post.loading}
        onBack={() => navigate(-1)}
      />

      <EditFeedback error={error()} success={success()} />

      <Show when={post.loading}>
        <EditSkeleton />
      </Show>

      <Show when={!post.loading && post() === null}>
        <EditNotFound onBack={() => navigate("/admin/noticias")} />
      </Show>

      <Show when={post()}>
        {(p) => {
          initForm(p());
          return (
            <form onSubmit={handleSubmit} class="space-y-6">

              <EditMetadataSection
                title={title}
                setTitle={setTitle}
                shortDescription={shortDescription}
                setShortDescription={setShortDescription}
                type={type}
                setType={setType}
                status={status}
                setStatus={setStatus}
                publishAt={publishAt}
                setPublishAt={setPublishAt}
              />

              <EditImageSection
                currentImageUrl={currentImage()}
                imageFile={imageFile}
                imagePreview={imagePreview}
                onImageChange={handleImageChange}
                onClearImage={clearImage}
              />

              <EditContentSection
                content={content()}
                onUpdate={(html) => setContent(html)}
              />

              <EditActions
                saving={saving()}
                status={status()}
                onCancel={() => navigate(-1)}
              />

            </form>
          );
        }}
      </Show>

    </main>
  );
}