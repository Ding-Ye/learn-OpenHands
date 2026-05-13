# learn-OpenHands

> Re-grow OpenHands' agent server from scratch in Go — one mechanism per chapter, each ending with a permalinked reading of the upstream Python.

[![Go](https://github.com/Ding-Ye/learn-OpenHands/actions/workflows/go.yml/badge.svg)](https://github.com/Ding-Ye/learn-OpenHands/actions/workflows/go.yml)
[![Docs](https://github.com/Ding-Ye/learn-OpenHands/actions/workflows/docs.yml/badge.svg)](https://github.com/Ding-Ye/learn-OpenHands/actions/workflows/docs.yml)
[![Web](https://github.com/Ding-Ye/learn-OpenHands/actions/workflows/web.yml/badge.svg)](https://github.com/Ding-Ye/learn-OpenHands/actions/workflows/web.yml)

## Why

[`OpenHands/OpenHands`](https://github.com/OpenHands/OpenHands) — formerly OpenDevin — is the open-source AI-driven development platform behind a 77.6 % SWE-bench score and a hosted product used by tens of thousands of engineers. Its main repo (the GUI / app-server / enterprise tier) is ~300 K LOC across `openhands/app_server/`, `openhands/server/`, `frontend/`, and `enterprise/`. Reading it cold is rough: the agent loop has been factored into a separate SDK, the server is a FastAPI app with dependency-injected services, and most of what's interesting lives behind abstractions like `EventService`, `AppConversationService`, `HookLoader`, `SkillLoader`.

This repo rebuilds the **server core** in Go, one chapter at a time. Each chapter is ≈ 200–500 lines of code that compiles independently, with a `## Upstream Source Reading` section at the end pinning permalinks against a frozen upstream SHA. You don't read about event sourcing — you implement an `EventService` that round-trips through disk and then go look at the Python it's modelled on.

## 为什么

[`OpenHands/OpenHands`](https://github.com/OpenHands/OpenHands)（原名 OpenDevin）是开源的 AI 驱动开发平台，背后是 77.6 % 的 SWE-bench 成绩和一个被数万工程师使用的托管产品。主仓库（GUI / app-server / enterprise）约 30 万行 LOC，分布在 `openhands/app_server/`、`openhands/server/`、`frontend/`、`enterprise/`。直接冷读很难：agent loop 已经被拆到独立 SDK 里、server 是个用 DI 容器组装的 FastAPI 应用、有意思的部分都藏在 `EventService` / `AppConversationService` / `HookLoader` / `SkillLoader` 这些抽象后面。

本仓库用 **Go** 一节一节把 **server 核心** 重建出来。每节 ≈ 200–500 行独立编译的代码，每节末尾的 `## 上游源码阅读` 都用固定 commit 的 permalink 锚定到上游 Python。你不是"读到事件溯源"——你是先实现一个能往磁盘里来回 round-trip 的 `EventService`，再去对照上游怎么写。

## Curriculum / 课程

| # | Slug | Title (EN) | 标题（中文） | Status |
|---|---|---|---|---|
| s01 | [`s01-event-store`](docs/en/s01-event-store.md) ([中](docs/zh/s01-event-store.md)) | Event store | 事件存储 | ✅ |
| s02 | `s02-conversation-lifecycle` | Conversation lifecycle | 对话生命周期 | ⏳ |
| s03 | `s03-pending-messages` | Pending-message queue | 待处理消息队列 | ⏳ |
| s04 | `s04-event-callbacks` | Event callbacks (webhook fan-out) | 事件回调（webhook 扇出） | ⏳ |
| s05 | `s05-skill-loader` | Skill loader (markdown frontmatter) | Skill 加载器（markdown frontmatter） | ⏳ |
| s06 | `s06-sandbox-runner` | Sandbox action runner | 沙箱 action runner | ⏳ |
| s07 | `s07-hook-loader` | Hook loader & plugin registry | Hook 加载器与插件注册表 | ⏳ |
| s08 | `s08-secrets-service` | Secrets service | 凭据存储 | ⏳ |
| s09 | `s09-file-store` | Per-conversation file store | 按对话隔离的文件存储 | ⏳ |
| s10 | `s10-live-status` | Live status (SSE push) | 实时状态（SSE 推送） | ⏳ |
| s11 | `s11-git-integration` | Git integration | Git 集成 | ⏳ |
| s12 | `s12-mcp-service` | MCP service | MCP 服务 | ⏳ |
| s13 | `s13-integrations` | Slack / Jira / Linear integrations | Slack / Jira / Linear 集成 | ⏳ |
| s14 | `s14-user-auth` | User auth & RBAC | 用户认证与 RBAC | ⏳ |
| s_full | `s_full-integration` | End-to-end integration | 端到端集成 | ⏳ |
| A | `appendix-a-event-sourcing` | The event-sourcing pattern | 事件溯源模式 | ⏳ |
| B | `appendix-b-upstream-map` | Upstream source-reading map | 上游源码导读地图 | ⏳ |

> ⏳ = curriculum slot reserved, not yet implemented. The schedule drips them in via the `learn-repo-generator` skill.

## Quickstart

```sh
cd agents/s01-event-store
go test ./...              # 7 tests, all green
go run ./cmd/demo          # 6-turn scripted conversation; replay + filter
```

Browse the curriculum docs:

```sh
cd web
npm install
npm run dev   # http://localhost:3000/en or /zh
```

## Layout

```
.
├── agents/                          # one Go module per chapter
│   └── s01-event-store/
│       ├── event.go                 # discriminated-union Event + payload types
│       ├── store.go                 # Store interface + FilesystemStore + Filter
│       ├── uuid.go                  # zero-dep RFC 4122 v4 UUID
│       ├── store_test.go            # 7 tests
│       └── cmd/demo/main.go         # scripted 6-turn conversation
├── docs/
│   ├── en/s01-event-store.md        # six-section English chapter
│   └── zh/s01-event-store.md        # six-section Chinese chapter
├── upstream-readings/
│   └── s01-event-store.py           # annotated excerpt of upstream
├── web/                             # Next.js bilingual doc viewer
├── .github/workflows/               # Go / web / docs CI
├── go.work                          # multi-module workspace
├── LICENSE                          # MIT, attributes upstream
└── README.md                        # this file
```

## Acknowledgements / 致谢

This repo is a learning derivative of [`OpenHands/OpenHands`](https://github.com/OpenHands/OpenHands), pinned at commit [`a89778f3`](https://github.com/OpenHands/OpenHands/tree/a89778f3d7036b8d81d57a1f93e31c6df8219eff). All upstream credit to the All-Hands team. Annotations, Go reimplementations, and bilingual docs are the contribution of this repo and ship under the same MIT license.

## License

MIT — see [LICENSE](./LICENSE). Upstream OpenHands is also MIT outside its `enterprise/` directory; that directory is excluded from this learn-* derivative.
