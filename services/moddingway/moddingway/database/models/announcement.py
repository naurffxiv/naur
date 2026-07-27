from pydantic import BaseModel, Field


class AnnouncementRevision(BaseModel):
    version: int
    content: str = Field(..., max_length=4000)
    author_id: int


class Announcement(BaseModel):
    announcement_id: int | None = None
    discord_message_id: int | None = None
    announcement_revisions: list[AnnouncementRevision] = []
    sent_flag: bool = False
