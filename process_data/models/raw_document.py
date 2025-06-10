from pydantic import BaseModel

class RawDocument(BaseModel):
    url: str
    category: str
    raw_html: str