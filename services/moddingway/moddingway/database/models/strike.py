from datetime import datetime

from pydantic import BaseModel

from moddingway.constants import StrikeSeverity
from moddingway.settings import get_settings

settings = get_settings()


class Strike(BaseModel):
    strike_id: int | None = None
    user_id: int
    severity: StrikeSeverity
    reason: str
    created_timestamp: datetime
    created_by: str
    last_edited_timestamp: datetime
    last_edited_by: str
