# s01 — 事件存储

> **上游：** `openhands/app_server/event/event_service.py` 与
> `openhands/app_server/event/filesystem_event_service.py`，commit
> [`a89778f3`](https://github.com/OpenHands/OpenHands/tree/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event)。

## 问题

OpenHands 的服务端本质是 *agent 化* 的——用户与自主 agent 之间会进行长对话，agent 发出 action，沙箱返回 observation，agent 再思考。栈底必须满足三件事：

1. **可回放。** live status 重连、事后排错、晚到的 callback——全都需要从 t=0 重读这条对话。
2. **类型不抹平。** 用户消息和沙箱观察不能互换，存储得记得每条事件是哪种。
3. **拒绝"当前状态"陷阱。** 如果唯一真相只活在 agent 进程内存里，重启就完了。所以存储才是事实源头，agent 只是它的一次 fold。

上游的答案就是 **`EventService`** 抽象：每条对话一份 append-only 日志，事件是 discriminated union，背后挂四种后端（filesystem / AWS / GCS / SQLite）。s01 用约 400 行 Go 把 filesystem 这套搬过来。

## 解法

```go
type Store interface {
    Save(conv UUID, e Event) error
    Get(conv, id UUID) (Event, error)
    Search(conv UUID, f Filter) ([]Event, error)
    Count(conv UUID, f Filter) (int, error)
}
```

加三种 payload 类型——`Message` / `Action` / `Observation`——以及一个 `Event` 外壳用 `Kind` 区分。filesystem 实现把每条事件写成
`<root>/<conversation>/<event-id>.json` 一个文件。

## 工作原理

```
            ┌──────────── 一次对话 ─────────────┐
            │                                    │
NewMessage ─▶ Event{Kind=message, Payload=…}      │
  Save      │                                    │
            ▼                                    │
       .../conv-abc/                             │
         ├── ev-001.json   (message  20:08:56.214) │
         ├── ev-002.json   (action   20:08:56.215) │
         ├── ev-003.json   (observ.  20:08:56.215) │
         ├── ev-004.json   (action   20:08:56.216) │
         ├── ev-005.json   (observ.  20:08:56.216) │
         └── ev-006.json   (message  20:08:56.216) │
                                                  │
Search(Filter{Kind:KindAction, SortAsc:true}) ───┤
  ↓                                               │
  扫目录、逐个 decode、过滤、排序、截断             │
  ↓                                               │
  []Event  ←──  调用方按 .Kind 分发再解码           │
            └────────────────────────────────────┘
```

几个关键选择，以及它们对应上游哪段：

- **文件名是 id，timestamp 写在 payload 里。** 对应上游 `_load_event` 的做法（时间从 JSON 取，不从文件名取）。让存储对跨写者时钟漂移更宽容。
- **`.tmp` + rename 准原子化 Save。** s01 对上游 `path.write_text` 的小升级——Go 标准库免费送的崩溃容忍。
- **`Filter` 是值，不是 builder。** 零值匹配所有事件。和上游"可选 kwarg 搜索"是同一种 ergonomics。
- **排序在读侧。** 无索引、无 compaction。一对话几千事件以内都没问题；什么时候要切到 SQL 见 appendix-b。

## 与上一节的 diff

这是第一章，diff 是相对空 workspace 的。唯一承重的设计决定是"事件长什么样"——这条钉死之后，后续每一章（callbacks、status、hooks）都可以默认 `Store` 已经在那里。

## 动手试

```sh
cd agents/s01-event-store
go test ./...         # 7 个测试
go run ./cmd/demo     # 6 步脚本化对话；输出 replay
```

看看 demo 写了哪些文件：

```sh
DIR=$(go run ./cmd/demo 2>&1 | awk '/store root:/{print $3}')
ls "$DIR"/*/
jq . "$DIR"/*/*.json | head -50
```

每条事件都是可读 JSON——这是刻意的。上游服务端的运维工具（`oh logs`、`oh export-conversation`）基本上就是"`cat` 这个目录"。

## 上游源码阅读

打开 [`upstream-readings/s01-event-store.py`](../../upstream-readings/s01-event-store.py)。它节选了 `event_service.py`（抽象基类）和 `filesystem_event_service.py`（唯一不靠云账号就能跑的后端），注释里标出了每个 Go 构造对应的 Python 方法。

之后把上游文件完整读一遍：

- [`openhands/app_server/event/event_service.py`](https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/event_service.py)
- [`openhands/app_server/event/filesystem_event_service.py`](https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/filesystem_event_service.py)
- [`openhands/app_server/event/event_service_base.py`](https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/event_service_base.py)

把抽象基类 + filesystem 后端看明白之后，再扫一眼 `aws_event_service.py` 和 `google_cloud_event_service.py`——它们换了底层存储原语，但四方法的形状原封不动。这种纪律性是 s02 的起点。
