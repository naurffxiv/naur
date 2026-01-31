from datetime import datetime

from pydantic import BaseModel


class Note(BaseModel):
    note_id: int | None = None
    user_id: int
    is_warning: bool
    note: str
    created_timestamp: datetime
    created_by: str
    last_edited_timestamp: datetime
    last_edited_by: str
