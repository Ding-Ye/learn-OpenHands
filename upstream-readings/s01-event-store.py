# Upstream excerpt: openhands/app_server/event/event_service.py
# and openhands/app_server/event/filesystem_event_service.py
#
# Pinned to upstream commit a89778f3d7036b8d81d57a1f93e31c6df8219eff.
# Permalinks:
#   https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/event_service.py
#   https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/filesystem_event_service.py
#
# License: MIT (OpenHands' main tree). The annotations below are mine;
# the code is the upstream's, lightly trimmed.


# ------------------------------------------------------------------
# event_service.py — the abstract base. Everything stateful in the
# OpenHands server ultimately calls one of these four methods.
# ------------------------------------------------------------------
class EventService(ABC):
    """Event Service for getting events."""

    @abstractmethod
    async def get_event(self, conversation_id: UUID, event_id: UUID) -> Event | None:
        """Given an id, retrieve an event."""

    @abstractmethod
    async def search_events(
        self,
        conversation_id: UUID,
        kind__eq: EventKind | None = None,        # ← `Filter.Kind` in Go
        timestamp__gte: datetime | None = None,   # ← `Filter.SinceInclude`
        timestamp__lt: datetime | None = None,    # ← `Filter.UntilExclude`
        sort_order: EventSortOrder = EventSortOrder.TIMESTAMP,
        page_id: str | None = None,               # s01 doesn't paginate yet
        limit: int = 100,                          # ← `Filter.Limit`
    ) -> EventPage:
        """Search events matching the given filters."""

    @abstractmethod
    async def save_event(self, conversation_id: UUID, event: Event):
        """Save an event. Internal method intended not be part of the REST api."""

    # NOTE — `batch_get_events` is a concrete default that fans out to
    # get_event. We don't reproduce it in s01 because we have nothing to
    # parallelise yet; come back to it when s04 callbacks need to load
    # many events at once.


# ------------------------------------------------------------------
# filesystem_event_service.py — the only backend that's interesting
# without a cloud account. Strips down to ~70 lines because the bulk of
# the per-conversation logic lives in EventServiceBase.
# ------------------------------------------------------------------
@dataclass
class FilesystemEventService(EventServiceBase):
    """Event service based on file system"""

    limit: int = 500

    def _load_event(self, path: Path) -> Event | None:
        # The "trust the writer's clock, don't trust the writer's filename"
        # discipline: we read the timestamp out of the JSON, not out of
        # the filename. Same choice the Go FilesystemStore makes.
        try:
            content = path.read_text()
            content = Event.model_validate_json(content)
            return content
        except Exception:
            if path.exists():
                _logger.exception('Error reading event', stack_info=True)
            return None

    def _store_event(self, path: Path, event: Event):
        # Note: no atomic rename here. The Go `FilesystemStore.Save`
        # upgrades this slightly by writing to `.tmp` first — small win
        # for crash-safety, free given Go's stdlib.
        path.parent.mkdir(parents=True, exist_ok=True)
        content = event.model_dump_json(indent=2)
        path.write_text(content)

    def _search_paths(self, prefix: Path, page_id: str | None = None) -> list[Path]:
        # The whole "scan the directory, decode every file" pattern. Fine
        # for human-scale conversations (<1k events). Production
        # replaces this with an SQLite or RocksDB index — see
        # sql_app_conversation_info_service.py for the same shape using
        # SQLAlchemy.
        search_path = f'{prefix}/*'
        files = glob.glob(str(search_path))
        return [Path(f) for f in files]


# ------------------------------------------------------------------
# Wiring: how the service shows up to a request handler.
# ------------------------------------------------------------------
class FilesystemEventServiceInjector(EventServiceInjector):
    async def inject(
        self, state: InjectorState, request: Request | None = None
    ) -> AsyncGenerator[EventService, None]:
        # `Injector` is OpenHands' DI primitive (s06 hook-loader plays in
        # the same plug-in registry). For s01 we don't need a DI
        # container — `NewFilesystemStore(root)` is constructor enough.
        from openhands.app_server.config import (
            get_app_conversation_info_service,
            get_global_config,
            get_user_context,
        )
        async with (
            get_user_context(state, request) as user_context,
            get_app_conversation_info_service(state, request) as app_conversation_info_service,
        ):
            # Per-user namespacing happens here: prefix = global config's
            # persistence_dir; concrete dir is {prefix}/{user_id}/v1_conversations.
            # In s01 we collapse to a single root.
            prefix = get_global_config().persistence_dir
            user_id = await user_context.get_user_id()
            yield FilesystemEventService(
                prefix=prefix,
                user_id=user_id,
                app_conversation_info_service=app_conversation_info_service,
                app_conversation_info_load_tasks={},
            )


# ------------------------------------------------------------------
# Deliberate omissions in s01 (we'll come back to these):
# - per-user namespacing → s14 user-auth
# - pagination via page_id → covered when s10 live-status needs it
# - injectors / DI → s07 hook-loader uses the same registry
# - cloud backends (aws_event_service.py, google_cloud_event_service.py)
#   → appendix-b walks the full upstream-readings map
# ------------------------------------------------------------------
