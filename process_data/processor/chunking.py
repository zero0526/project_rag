import re
from langchain.text_splitter import RecursiveCharacterTextSplitter

def chunk_text(text: str, chunk_size=512, chunk_overlap=64)->list[str]:
    splitter = RecursiveCharacterTextSplitter(
        chunk_size=chunk_size,
        chunk_overlap=chunk_overlap,
        separators=["\n\n", "\n", ".","!","?", ",", " "]
    )
    return splitter.split_text(text)
