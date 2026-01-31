from pydantic import BaseModel


class Banned(BaseModel):
    userID: str
