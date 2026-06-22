import { useEffect, useState } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import { codeToHtml } from "shiki";
import { cn } from "@/lib/utils";

// Markdown renders task content (body, notes) as GitHub-flavored markdown. Code blocks are
// highlighted with Shiki — lazy-loaded per language so "all languages" doesn't bloat the
// bundle. react-markdown ignores raw HTML by default (no rehype-raw), so no sanitizer is
// needed for this local single-user tool.
export function Markdown({
  children,
  className,
  inline,
}: {
  children: string;
  className?: string;
  inline?: boolean;
}) {
  // Inline variant (e.g. provenance notes): no block margins, just spans/code/links.
  if (inline) {
    return (
      <span className={cn("cairn-md-inline", className)}>
        <ReactMarkdown remarkPlugins={[remarkGfm]} components={inlineComponents}>
          {children}
        </ReactMarkdown>
      </span>
    );
  }
  return (
    <div
      className={cn(
        "prose prose-sm max-w-none dark:prose-invert",
        // let Shiki own code styling; neutralize prose's pre/code chrome
        "prose-pre:m-0 prose-pre:bg-transparent prose-pre:p-0",
        "prose-headings:font-semibold prose-headings:tracking-tight prose-a:text-brand",
        className,
      )}
    >
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={blockComponents}>
        {children}
      </ReactMarkdown>
    </div>
  );
}

function CodeBlock({ lang, code }: { lang: string; code: string }) {
  const [html, setHtml] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    codeToHtml(code, {
      lang,
      themes: { light: "github-light-default", dark: "github-dark-default" },
      defaultColor: false,
    })
      .then((h) => active && setHtml(h))
      .catch(() => active && setHtml(null));
    return () => {
      active = false;
    };
  }, [lang, code]);

  return (
    <div className="my-3 overflow-hidden rounded-lg border text-[13px]">
      {html ? (
        <div
          className="[&_pre]:overflow-x-auto [&_pre]:p-3"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      ) : (
        <pre className="overflow-x-auto bg-muted/40 p-3 font-mono">
          <code>{code}</code>
        </pre>
      )}
    </div>
  );
}

const InlineCode = ({ children }: { children?: React.ReactNode }) => (
  <code className="rounded bg-muted px-1 py-0.5 font-mono text-[0.85em]">{children}</code>
);

// In react-markdown v9, fenced blocks arrive as <pre><code class="language-x">. We unwrap
// <pre> and let CodeBlock render its own Shiki <pre>; inline code stays a styled <code>.
const codeComponent: Components["code"] = ({ className, children }) => {
  const match = /language-(\w+)/.exec(className || "");
  if (match) return <CodeBlock lang={match[1]} code={String(children).replace(/\n$/, "")} />;
  return <InlineCode>{children}</InlineCode>;
};

const blockComponents: Components = {
  code: codeComponent,
  pre: ({ children }) => <>{children}</>,
};

// Inline rendering: drop block elements to their text/inline equivalents.
const inlineComponents: Components = {
  p: ({ children }) => <>{children}</>,
  code: ({ children }) => <InlineCode>{children}</InlineCode>,
  a: ({ href, children }) => (
    <a href={href} target="_blank" rel="noreferrer" className="text-brand underline">
      {children}
    </a>
  ),
};
