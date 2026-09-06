// web/src/components/ui/RichTextEditor.tsx
import { onMount, onCleanup, Show, createSignal, createEffect } from "solid-js";
import { Editor } from "@tiptap/core";
import { sanitizeHtml } from "~/lib/sanitize-html";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TextAlign from "@tiptap/extension-text-align";
import Link from "@tiptap/extension-link";

interface RichTextEditorProps {
  content: string;
  onUpdate: (html: string) => void;
  label?: string;
}

export function RichTextEditor(props: RichTextEditorProps) {
  let editorRef!: HTMLDivElement;
  let editor: Editor;

  const [active, setActive] = createSignal({
    bold: false,
    italic: false,
    underline: false,
    h1: false,
    h2: false,
    h3: false,
    bulletList: false,
    orderedList: false,
    blockquote: false,
    alignLeft: false,
    alignCenter: false,
    alignRight: false,
    alignJustify: false,
    link: false,
  });

  const [showPreview, setShowPreview] = createSignal(false);
  const [previewHtml, setPreviewHtml] = createSignal(props.content);

  onMount(() => {
    editor = new Editor({
      element: editorRef,
      extensions: [
        StarterKit.configure({
          underline: false,
          link: false,
        }),
        Underline,
        TextAlign.configure({ types:["heading", "paragraph"], }),
        Link.configure({ openOnClick: false }), // Importante: falso para poder editar el link sin abrirlo
      ],
      content: props.content,
      editorProps: {
        attributes: {
          class:
            "prose prose-blue max-w-none focus:outline-none min-h-[260px] p-6 text-gray-800 leading-relaxed text-[15px]",
        },
      },
      onUpdate: ({ editor }) => {
        const html = editor.getHTML();
        props.onUpdate(html);
        setPreviewHtml(html);
        updateActiveStates(editor);
      },
      onSelectionUpdate: ({ editor }) => {
        updateActiveStates(editor);
      },
    });
  });

  // ── REACTIVIDAD EXTERNA (Para cuando los datos llegan por Fetch/API) ──
  createEffect(() => {
    const newContent = props.content;
    
    // Si el editor ya existe y el contenido externo es diferente al interno
    if (editor && newContent !== undefined) {
      const currentEditorContent = editor.getHTML();
      
      // Verificación clave para evitar bucles infinitos de re-renderizado
      if (newContent !== currentEditorContent && newContent !== "<p></p>") {
        
        // FIX SENIOR: Tiptap ahora espera un objeto de configuración (SetContentOptions)
        editor.commands.setContent(newContent, { emitUpdate: false }); 
        
        setPreviewHtml(newContent);
      }
    }
  });

  onCleanup(() => {
    if (editor) editor.destroy();
  });

  const updateActiveStates = (ed: Editor) => {
    setActive({
      bold: ed.isActive("bold"),
      italic: ed.isActive("italic"),
      underline: ed.isActive("underline"),
      h1: ed.isActive("heading", { level: 1 }),
      h2: ed.isActive("heading", { level: 2 }),
      h3: ed.isActive("heading", { level: 3 }),
      bulletList: ed.isActive("bulletList"),
      orderedList: ed.isActive("orderedList"),
      blockquote: ed.isActive("blockquote"),
      alignLeft: ed.isActive({ textAlign: "left" }),
      alignCenter: ed.isActive({ textAlign: "center" }),
      alignRight: ed.isActive({ textAlign: "right" }),
      alignJustify: ed.isActive({ textAlign: "justify" }),
      link: ed.isActive("link"),
    });
  };

  const handleSetLink = () => {
    // Si ya hay un link seleccionado, lo mostramos para que el usuario pueda editarlo
    const previousUrl = editor.getAttributes("link").href || "";
    const url = window.prompt("Ingresa la URL del enlace:", previousUrl);
    
    // Si presiona Cancelar, no hacemos nada
    if (url === null) return; 

    // Si vacía el campo, eliminamos el enlace
    if (url === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      return;
    }

    // Normalización: Aseguramos que el enlace tenga un protocolo válido para evitar rutas relativas rotas
    let safeUrl = url.trim();
    if (!safeUrl.startsWith("http://") && !safeUrl.startsWith("https://") && !safeUrl.startsWith("mailto:")) {
      safeUrl = "https://" + safeUrl;
    }

    editor.chain().focus().extendMarkRange("link").setLink({ href: safeUrl }).run();
  };

  // ── ESTILOS DE BOTONES (UI) ──────────────────────────────────────────────────
  const btn = (isActive: boolean, extra = "") =>
    `inline-flex items-center justify-center h-8 w-8 rounded-lg text-sm font-semibold transition-all select-none cursor-pointer border ${
      isActive
        ? "bg-colpsi-blue text-white border-colpsi-blue shadow-inner"
        : "bg-white text-gray-600 border-gray-200 hover:bg-gray-100 hover:border-gray-300"
    } ${extra}`;

  const wideBtn = (isActive: boolean) =>
    `inline-flex items-center gap-1.5 px-3 h-8 rounded-lg text-xs font-bold transition-all select-none cursor-pointer border ${
      isActive
        ? "bg-colpsi-blue text-white border-colpsi-blue shadow-inner"
        : "bg-white text-gray-600 border-gray-200 hover:bg-gray-100 hover:border-gray-300"
    }`;

  const Sep = () => <div class="w-px h-6 bg-gray-200 mx-1 self-center" />;
  

  return (
    <div class="space-y-2 font-sans">
      <Show when={props.label}>
        <label class="block text-xs font-bold text-gray-500 uppercase tracking-wider ml-1 mb-1">
          {props.label}
        </label>
      </Show>

      {/* ── CONTENEDOR PRINCIPAL ─────────────────────────────────────────── */}
      <div class="border-2 border-gray-200 rounded-2xl overflow-hidden bg-white shadow-sm focus-within:border-colpsi-blue focus-within:shadow-md transition-all duration-200">

        {/* ── TOOLBAR ─────────────────────────────────────────────────────── */}
        <div class="flex flex-wrap items-center gap-1 px-3 py-2 border-b border-colpsi-border bg-gradient-to-r from-gray-50 to-white">

          {/* Encabezados */}
          <button type="button" title="Título 1" class={wideBtn(active().h1)}
            onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}>
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" d="M4 6h16M4 12h8M4 18h8" />
            </svg>
            H1
          </button>
          <button type="button" title="Título 2" class={wideBtn(active().h2)}
            onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}>
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" d="M4 6h16M4 12h8M4 18h8" />
            </svg>
            H2
          </button>
          <button type="button" title="Título 3" class={wideBtn(active().h3)}
            onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}>
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" d="M4 6h12M4 12h8M4 18h8" />
            </svg>
            H3
          </button>

          <Sep />

          {/* Formato de texto */}
          <button type="button" title="Negrita (Ctrl+B)" class={btn(active().bold)}
            onClick={() => editor.chain().focus().toggleBold().run()}>
            <span class="font-black text-[13px]">B</span>
          </button>
          <button type="button" title="Cursiva (Ctrl+I)" class={btn(active().italic)}
            onClick={() => editor.chain().focus().toggleItalic().run()}>
            <span class="italic font-bold text-[13px]">I</span>
          </button>
          <button type="button" title="Subrayado (Ctrl+U)" class={btn(active().underline)}
            onClick={() => editor.chain().focus().toggleUnderline().run()}>
            <span class="underline font-bold text-[13px]">U</span>
          </button>

          <Sep />

          {/* Alineación */}
          <button type="button" title="Alinear izquierda" class={btn(active().alignLeft)}
            onClick={() => editor.chain().focus().setTextAlign("left").run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <rect x="3" y="5" width="18" height="2" rx="1"/>
              <rect x="3" y="10" width="12" height="2" rx="1"/>
              <rect x="3" y="15" width="18" height="2" rx="1"/>
              <rect x="3" y="20" width="10" height="2" rx="1"/>
            </svg>
          </button>
          <button type="button" title="Centrar" class={btn(active().alignCenter)}
            onClick={() => editor.chain().focus().setTextAlign("center").run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <rect x="3" y="5" width="18" height="2" rx="1"/>
              <rect x="6" y="10" width="12" height="2" rx="1"/>
              <rect x="3" y="15" width="18" height="2" rx="1"/>
              <rect x="7" y="20" width="10" height="2" rx="1"/>
            </svg>
          </button>
          <button type="button" title="Alinear derecha" class={btn(active().alignRight)}
            onClick={() => editor.chain().focus().setTextAlign("right").run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <rect x="3" y="5" width="18" height="2" rx="1"/>
              <rect x="9" y="10" width="12" height="2" rx="1"/>
              <rect x="3" y="15" width="18" height="2" rx="1"/>
              <rect x="11" y="20" width="10" height="2" rx="1"/>
            </svg>
          </button>
          <button type="button" title="Justificar" class={btn(active().alignJustify)}
            onClick={() => editor.chain().focus().setTextAlign("justify").run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <rect x="3" y="5" width="18" height="2" rx="1"/>
              <rect x="3" y="10" width="18" height="2" rx="1"/>
              <rect x="3" y="15" width="18" height="2" rx="1"/>
              <rect x="3" y="20" width="18" height="2" rx="1"/>
            </svg>
          </button>

          <Sep />

          {/* Listas & Citas */}
          <button type="button" title="Lista con viñetas" class={btn(active().bulletList)}
            onClick={() => editor.chain().focus().toggleBulletList().run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <circle cx="5" cy="7" r="1.5"/>
              <rect x="9" y="6" width="12" height="2" rx="1"/>
              <circle cx="5" cy="12" r="1.5"/>
              <rect x="9" y="11" width="12" height="2" rx="1"/>
              <circle cx="5" cy="17" r="1.5"/>
              <rect x="9" y="16" width="12" height="2" rx="1"/>
            </svg>
          </button>
          <button type="button" title="Lista numerada" class={btn(active().orderedList)}
            onClick={() => editor.chain().focus().toggleOrderedList().run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <text x="3" y="9" font-size="7" font-weight="bold">1.</text>
              <rect x="9" y="6" width="12" height="2" rx="1"/>
              <text x="3" y="14" font-size="7" font-weight="bold">2.</text>
              <rect x="9" y="11" width="12" height="2" rx="1"/>
              <text x="3" y="19" font-size="7" font-weight="bold">3.</text>
              <rect x="9" y="16" width="12" height="2" rx="1"/>
            </svg>
          </button>
          <button type="button" title="Cita" class={btn(active().blockquote)}
            onClick={() => editor.chain().focus().toggleBlockquote().run()}>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M6.5 10C6.5 8.07 8.07 6.5 10 6.5v2A1.5 1.5 0 0 0 8.5 10v1H10v4H6.5v-5zm7 0c0-1.93 1.57-3.5 3.5-3.5v2A1.5 1.5 0 0 0 15.5 10v1H17v4h-3.5v-5z"/>
            </svg>
          </button>

          <Sep />

          {/* Enlace */}
          <button type="button" title="Insertar/Editar enlace" class={btn(active().link)}
            onClick={handleSetLink}>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
          </button>

          {/* Deshacer / Rehacer */}
          <Sep />
          <button type="button" title="Deshacer (Ctrl+Z)" class={btn(false)}
            onClick={() => editor.chain().focus().undo().run()}>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h10a8 8 0 010 16H3m0-16L7 6m-4 4l4 4" />
            </svg>
          </button>
          <button type="button" title="Rehacer (Ctrl+Y)" class={btn(false)}
            onClick={() => editor.chain().focus().redo().run()}>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 10H11a8 8 0 000 16h10m0-16l-4-4m4 4l-4 4" />
            </svg>
          </button>

          <div class="flex-1" />

          {/* Botón Vista Previa */}
          <button
            type="button"
            title="Vista previa del contenido"
            onClick={() => setShowPreview(true)}
            class="inline-flex items-center gap-1.5 px-3 h-8 rounded-lg text-xs font-bold bg-colpsi-yellow/20 text-colpsi-blue border border-colpsi-yellow/40 hover:bg-colpsi-yellow/40 transition-all"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
            Vista previa
          </button>
        </div>

        {/* ── ÁREA DE ESCRITURA ─────────────────────────────────────────── */}
        <div
          ref={editorRef}
          class="bg-white cursor-text"
          onClick={() => editor?.commands.focus()}
        />
      </div>

      {/* ── MODAL DE VISTA PREVIA ─────────────────────────────────────────── */}
      <Show when={showPreview()}>
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
          onClick={(e) => { if (e.target === e.currentTarget) setShowPreview(false); }}
        >
          <div class="bg-white rounded-2xl shadow-2xl w-full max-w-3xl max-h-[80vh] flex flex-col overflow-hidden border border-gray-200">
            
            {/* Header modal */}
            <div class="flex items-center justify-between px-6 py-4 border-b border-colpsi-border bg-colpsi-surface">
              <div class="flex items-center gap-2">
                <svg class="w-5 h-5 text-colpsi-blue" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <h2 class="font-bold text-gray-800 text-sm uppercase tracking-wide">Vista previa de la publicación</h2>
              </div>
              <button
                type="button"
                onClick={() => setShowPreview(false)}
                class="w-8 h-8 flex items-center justify-center rounded-full text-gray-400 hover:bg-gray-200 hover:text-gray-700 transition-all font-bold text-lg"
              >
                ×
              </button>
            </div>

            {/* Contenido renderizado */}
            <div class="overflow-y-auto p-8">
              <div
                class="prose prose-blue max-w-none text-gray-800"
                innerHTML={sanitizeHtml(previewHtml())}
              />
            </div>

            {/* Footer */}
            <div class="px-6 py-4 border-t border-colpsi-border bg-colpsi-surface flex justify-end">
              <button
                type="button"
                onClick={() => setShowPreview(false)}
                class="px-6 py-2.5 rounded-xl bg-colpsi-blue text-white text-sm font-bold hover:bg-blue-800 transition-all shadow-md active:scale-95"
              >
                Cerrar vista previa
              </button>
            </div>
          </div>
        </div>
      </Show>
    </div>
  );
}