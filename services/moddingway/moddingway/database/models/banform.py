from datetime import datetime

from pydantic import BaseModel

from moddingway.settings import get_settings

settings = get_settings()


class BanForm(BaseModel):
    form_id: int | None = None
    user_id: int
    reason: str
    approval: bool | None = None
    approved_by: int | None = None
    submission_timestamp: datetime
