import numpy as np
from storage.mongodb import indices2article
from embedding.embedder import embed_chunks
from storage.faiss_store import search_embeddings

def getContext(query: str)->list:
    vec = embed_chunks(query)
    dis, inds  = search_embeddings(vec, 3)
    articles = indices2article(inds)
    return articles


