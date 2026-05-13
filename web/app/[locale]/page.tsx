import Link from "next/link";
import { notFound } from "next/navigation";
import { CURRICULUM, chapterTitle, type Locale } from "@/lib/curriculum";

export default async function Landing({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  if (locale !== "zh" && locale !== "en") notFound();
  const l = locale as Locale;

  const intro = l === "zh" ? INTRO_ZH : INTRO_EN;
  const ctaLabel = l === "zh" ? "从 s01 开始 →" : "Start at s01 →";

  return (
    <article className="prose-doc">
      <h1>learn-OpenHands</h1>
      <p className="text-[var(--fg-muted)]">
        {l === "zh"
          ? "用 Go 从零渐进重建 OpenHands agent 服务端的核心抽象，每节末尾对照上游 Python 源码。"
          : "Rebuild the core of OpenHands' agent server in Go, chapter by chapter — each one ends with an upstream Python source reading."}
      </p>

      {intro.map((p, i) => (
        <p key={i}>{p}</p>
      ))}

      <p>
        <Link
          href={`/${l}/s/s01-event-store`}
          className="inline-block mt-2 px-4 py-2 rounded border border-[var(--accent-soft)] hover:border-[var(--accent)]"
        >
          {ctaLabel}
        </Link>
      </p>

      <h2>{l === "zh" ? "课程" : "Curriculum"}</h2>
      <ul>
        {CURRICULUM.map((c) => (
          <li key={c.slug}>
            <span className="font-mono text-[var(--fg-muted)] mr-2">
              {c.num}
            </span>
            {c.available ? (
              <Link href={`/${l}/s/${c.slug}`}>{chapterTitle(c, l)}</Link>
            ) : (
              <span className="text-[var(--fg-muted)]">
                {chapterTitle(c, l)}{" "}
                <span className="text-xs">
                  ({l === "zh" ? "未发布" : "not yet"})
                </span>
              </span>
            )}
          </li>
        ))}
      </ul>
    </article>
  );
}

const INTRO_ZH = [
  "OpenHands（前身 OpenDevin）是开源的 AI 驱动开发平台，背后是 77.6% 的 SWE-bench 成绩和一个被数万工程师使用的托管产品。主仓库约 30 万行 LOC，分布在 `openhands/app_server/`、`openhands/server/`、`frontend/`、`enterprise/`。直接冷读很难：agent loop 已经被拆到独立 SDK 里、server 是个用 DI 容器组装的 FastAPI 应用。",
  "这个仓库不是教你「跑」 OpenHands，而是把它的服务端核心一节一节重建出来——事件存储、对话生命周期、pending 消息队列、回调扇出、skill 加载器、沙箱 runner、hook loader、secrets、文件存储、live status、git/MCP/integrations、用户认证——每节加一个机制，用 Go 写一份精简实现。",
  "Go 实现是教学骨架，上游 Python 是生产实现。每节末尾的「上游源码阅读」用固定 commit 的 permalink 锚定到上游真实代码。",
];

const INTRO_EN = [
  "OpenHands (formerly OpenDevin) is the open-source AI-driven development platform behind a 77.6 % SWE-bench score and a hosted product used by tens of thousands of engineers. The main repo is ~300 K LOC across `openhands/app_server/`, `openhands/server/`, `frontend/`, and `enterprise/`. Reading it cold is rough: the agent loop has been factored into a separate SDK, and the server is a FastAPI app with dependency-injected services.",
  "This repo doesn't teach you how to *run* OpenHands. It rebuilds the server core in Go, one mechanism per chapter — event store, conversation lifecycle, pending-message queue, callback fan-out, skill loader, sandbox runner, hook loader, secrets, file store, live status, git/MCP/integrations, user auth. After fourteen chapters, the upstream server stops looking like magic.",
  "Go is the teaching skeleton; the upstream Python is the production implementation. Every 'Upstream Source Reading' pins permalinks against a frozen upstream SHA.",
];
