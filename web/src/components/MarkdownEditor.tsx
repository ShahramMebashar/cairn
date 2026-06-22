import { useEffect } from "react";
import { useEditor, EditorContent, type Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { Markdown } from "tiptap-markdown";
import {
  Bold,
  Code,
  Code2,
  Heading1,
  Heading2,
  Italic,
  Link as LinkIcon,
  List,
  ListOrdered,
} from "lucide-react";
import { cn } from "@/lib/utils";

// tiptap-markdown adds a `markdown` storage with getMarkdown(); it isn't typed in v3.
function getMarkdown(editor: Editor): string {
  return (editor.storage as { markdown?: { getMarkdown(): string } }).markdown?.getMarkdown() ?? "";
}

// MarkdownEditor is a TipTap WYSIWYG that reads and writes **markdown** (via tiptap-markdown).
// Typing markdown shortcuts works (#, -, 1., ```, **bold**); the toolbar covers the rest.
// value/onChange are always markdown strings, so storage stays plain markdown.
export function MarkdownEditor({
  value,
  onChange,
  placeholder,
  minHeight = "5rem",
}: {
  value: string;
  onChange: (md: string) => void;
  placeholder?: string;
  minHeight?: string;
}) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({ link: { openOnClick: false } }),
      Placeholder.configure({ placeholder: placeholder ?? "Write…" }),
      Markdown.configure({ html: false, linkify: true, transformPastedText: true }),
    ],
    content: value,
    onUpdate: ({ editor }) => onChange(getMarkdown(editor)),
    editorProps: {
      attributes: {
        class: "prose prose-sm dark:prose-invert max-w-none focus:outline-none px-3 py-2",
        style: `min-height:${minHeight}`,
      },
    },
  });

  // Sync external value changes (e.g. reset to "" after submit) without clobbering typing.
  useEffect(() => {
    if (!editor) return;
    if (value !== getMarkdown(editor)) {
      editor.commands.setContent(value || "");
    }
  }, [value, editor]);

  return (
    <div className="overflow-hidden rounded-md border focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50">
      {editor && <Toolbar editor={editor} />}
      <EditorContent editor={editor} />
    </div>
  );
}

function Toolbar({ editor }: { editor: Editor }) {
  const btn = (
    active: boolean,
    onClick: () => void,
    Icon: typeof Bold,
    label: string,
  ) => (
    <button
      type="button"
      aria-label={label}
      onMouseDown={(e) => e.preventDefault()} // keep editor focus
      onClick={onClick}
      className={cn(
        "grid size-7 place-items-center rounded text-muted-foreground hover:bg-foreground/10 hover:text-foreground",
        active && "bg-foreground/10 text-foreground",
      )}
    >
      <Icon className="size-3.5" />
    </button>
  );

  const c = () => editor.chain().focus();

  return (
    <div className="flex flex-wrap items-center gap-0.5 border-b bg-muted/30 px-1.5 py-1">
      {btn(editor.isActive("bold"), () => c().toggleBold().run(), Bold, "Bold")}
      {btn(editor.isActive("italic"), () => c().toggleItalic().run(), Italic, "Italic")}
      {btn(editor.isActive("code"), () => c().toggleCode().run(), Code, "Inline code")}
      <span className="mx-0.5 h-4 w-px bg-border" />
      {btn(editor.isActive("heading", { level: 1 }), () => c().toggleHeading({ level: 1 }).run(), Heading1, "Heading 1")}
      {btn(editor.isActive("heading", { level: 2 }), () => c().toggleHeading({ level: 2 }).run(), Heading2, "Heading 2")}
      <span className="mx-0.5 h-4 w-px bg-border" />
      {btn(editor.isActive("bulletList"), () => c().toggleBulletList().run(), List, "Bullet list")}
      {btn(editor.isActive("orderedList"), () => c().toggleOrderedList().run(), ListOrdered, "Numbered list")}
      {btn(editor.isActive("codeBlock"), () => c().toggleCodeBlock().run(), Code2, "Code block")}
      {btn(editor.isActive("link"), () => toggleLink(editor), LinkIcon, "Link")}
    </div>
  );
}

function toggleLink(editor: Editor) {
  if (editor.isActive("link")) {
    editor.chain().focus().unsetLink().run();
    return;
  }
  const url = window.prompt("Link URL");
  if (url) editor.chain().focus().setLink({ href: url }).run();
}
