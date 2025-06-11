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
    return str(result.inserted_id)

def store_chunking(meta_chunking):
    try:
        collectionChunk.insert_many(meta_chunking)
        print(f"Đã lưu {len(meta_chunking)} bản ghi vào MongoDB.")
    except Exception as e:
        print(f"Lỗi khi lưu vào MongoDB: {e}")

def get_context_window(chunksLength: int, index: int, window_size: int = 3):
    half = window_size // 2

    start = max(index - half, 0)
    end = min(index + half + 1, chunksLength)

    actual_window = end - start

    if actual_window < window_size:
        shortage = window_size - actual_window

        if end + shortage <= chunksLength:
            end += shortage
        else:
            start = max(start - (window_size - (end - start)), 0)

    return start, end

def indices2article(faiss_indices: list) -> list:
    if isinstance(faiss_indices, np.ndarray):
        faiss_indices = faiss_indices.flatten().tolist()
    elif faiss_indices and isinstance(faiss_indices[0], list):
        faiss_indices = [i for sub in faiss_indices for i in sub]
    
    faiss_indices = [i for i in faiss_indices if i != -1]

    if not faiss_indices:
        return []

    chunks = list(collectionChunk.find({"_id": {"$in": faiss_indices}}))
    chunksIdx = []
    for c in chunks:
        st, en = get_context_window(c["len"], c["chunk_index"])
        chunksIdx.append((st, en, ObjectId(c["record_id"])))
    record_ids = list({chunk["record_id"] for chunk in chunks})
    
    try:
        record_ids = [ObjectId(rid) for rid in record_ids]
    except Exception as e:
        print("Error: {e}")

    articles = list(collection.find({"_id": {"$in": record_ids}}))
    context = []
    for art in articles:
        for start, end, recordId in chunksIdx:
            if art["_id"]== recordId:
                list_context = art["chunks"][start:end]
                context.append(merge_chunks(list_context))
    print("svfsf", context)
    return context


def merge_chunks(chunks: list[str]) -> str:
    length = len(chunks)
    context = ""
    if length == 1:
        return chunks[0]
    for i in range(1,length):
        context = context + merge_strings(chunks[0], chunks[1])
    return context

def compute_lps(pattern):
    m = len(pattern)
    lps = [0] * m
    length = 0
    i = 1
    while i < m:
        if pattern[i] == pattern[length]:
            length += 1
            lps[i] = length
            i += 1
        else:
            if length != 0:
                length = lps[length - 1]
            else:
                lps[i] = 0
                i += 1
    return lps

def find_overlap_length(a, b):
    combined = b + "#" + a
    lps = compute_lps(combined)
    return lps[-1]

def merge_strings(a, b):
    overlap_len = find_overlap_length(a, b)
    return a + b[overlap_len:]
