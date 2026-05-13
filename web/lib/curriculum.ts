// Curriculum locked from .learn/plan.md.

export type ChapterMeta = {
  slug: string;
  num: string;
  title: { zh: string; en: string };
  available: boolean;
};

export const CURRICULUM: ChapterMeta[] = [
  {
    slug: "s01-event-store",
    num: "s01",
    title: { zh: "事件存储", en: "Event store" },
    available: true,
  },
  {
    slug: "s02-conversation-lifecycle",
    num: "s02",
    title: { zh: "对话生命周期", en: "Conversation lifecycle" },
    available: false,
  },
  {
    slug: "s03-pending-messages",
    num: "s03",
    title: { zh: "待处理消息队列", en: "Pending-message queue" },
    available: false,
  },
  {
    slug: "s04-event-callbacks",
    num: "s04",
    title: { zh: "事件回调（webhook 扇出）", en: "Event callbacks (webhook fan-out)" },
    available: false,
  },
  {
    slug: "s05-skill-loader",
    num: "s05",
    title: { zh: "Skill 加载器（markdown frontmatter）", en: "Skill loader (markdown frontmatter)" },
    available: false,
  },
  {
    slug: "s06-sandbox-runner",
    num: "s06",
    title: { zh: "沙箱 action runner", en: "Sandbox action runner" },
    available: false,
  },
  {
    slug: "s07-hook-loader",
    num: "s07",
    title: { zh: "Hook 加载器与插件注册表", en: "Hook loader & plugin registry" },
    available: false,
  },
  {
    slug: "s08-secrets-service",
    num: "s08",
    title: { zh: "凭据存储", en: "Secrets service" },
    available: false,
  },
  {
    slug: "s09-file-store",
    num: "s09",
    title: { zh: "按对话隔离的文件存储", en: "Per-conversation file store" },
    available: false,
  },
  {
    slug: "s10-live-status",
    num: "s10",
    title: { zh: "实时状态（SSE 推送）", en: "Live status (SSE push)" },
    available: false,
  },
  {
    slug: "s11-git-integration",
    num: "s11",
    title: { zh: "Git 集成", en: "Git integration" },
    available: false,
  },
  {
    slug: "s12-mcp-service",
    num: "s12",
    title: { zh: "MCP 服务", en: "MCP service" },
    available: false,
  },
  {
    slug: "s13-integrations",
    num: "s13",
    title: { zh: "Slack / Jira / Linear 集成", en: "Slack / Jira / Linear integrations" },
    available: false,
  },
  {
    slug: "s14-user-auth",
    num: "s14",
    title: { zh: "用户认证与 RBAC", en: "User auth & RBAC" },
    available: false,
  },
  {
    slug: "s_full-integration",
    num: "s_full",
    title: { zh: "端到端集成", en: "End-to-end integration" },
    available: false,
  },
  {
    slug: "appendix-a-event-sourcing",
    num: "A",
    title: {
      zh: "附录 A · 事件溯源模式",
      en: "Appendix A · The event-sourcing pattern",
    },
    available: false,
  },
  {
    slug: "appendix-b-upstream-map",
    num: "B",
    title: { zh: "附录 B · 上游源码导读地图", en: "Appendix B · Upstream source-reading map" },
    available: false,
  },
];

export type Locale = "zh" | "en";

export function chapterTitle(c: ChapterMeta, locale: Locale): string {
  return c.title[locale];
}
