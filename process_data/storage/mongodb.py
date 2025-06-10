from pymongo import MongoClient
from bson.objectid import ObjectId
import numpy as np
client = MongoClient("mongodb://localhost:27018/")
db = client["contextual_db"]
collection = db["articles"]
collectionChunk = db["chunking"]

def store_metadata(doc, chunks):
    record = doc.dict()
    result = collection.insert_one(record)
    print(result)
    return str(result.inserted_id)

def store_chunking(meta_chunking):
    """    
    Args:
        meta_chunking (list): Danh sách các dictionary chứa metadata của vector
    """
    try:
        collectionChunk.insert_many(meta_chunking)
        print(f"Đã lưu {len(meta_chunking)} bản ghi vào MongoDB.")
    except Exception as e:
        print(f"Lỗi khi lưu vào MongoDB: {e}")

def indices2article(faiss_indices: list) -> list:
    if isinstance(faiss_indices, np.ndarray):
        faiss_indices = faiss_indices.flatten().tolist()
    elif faiss_indices and isinstance(faiss_indices[0], list):
        faiss_indices = [i for sub in faiss_indices for i in sub]

    faiss_indices = [i for i in faiss_indices if i != -1]

    if not faiss_indices:
        return []

    chunks = list(collectionChunk.find({"_id": {"$in": faiss_indices}}))

    record_ids = list({chunk["record_id"] for chunk in chunks})

    try:
        record_ids = [ObjectId(rid) for rid in record_ids]
    except Exception:
        pass

    articles = list(collection.find({"_id": {"$in": record_ids}}))
    for art in articles:
        print(art["content"])
    return articles
