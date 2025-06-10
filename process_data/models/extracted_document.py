from typing import List
from pydantic import BaseModel

class ExtractedDocument(BaseModel):
    url: str
    category: str
    title: str
    author: str
    content: str
    date: str
    related_links: List[str]
