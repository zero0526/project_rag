from typing import List
from pydantic import BaseModel
from typing import List, Optional

class ExtractedDocument(BaseModel):
    url: str
    category: str
    title: str
    author: str
    content: str
    date: str
    chunks: Optional[List[str]] = []
    related_links: List[str]
