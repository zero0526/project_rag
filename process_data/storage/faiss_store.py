import numpy as np
import faiss
from datetime import  datetime
from storage.mongodb import store_chunking
d = 768 
index = faiss.IndexFlatL2(d)  # Sử dụng IndexFlatL2 cho tìm kiếm chính xác

def save_embeddings(vectors, record_id):
    """
    Lưu vector embeddings vào Faiss và metadata vào MongoDB.
    
    Args:
        vectors (list or np.array): Danh sách các vector embeddings
        record_id (str): ID của bản ghi (record)
    """
    global index
    
    vectors = np.array(vectors).astype("float32")
    start_faiss_id = index.ntotal
    index.add(vectors)
    
    length = len(vectors)
    metaChunkings = []
    for i in range(length):
        faiss_id = start_faiss_id + i

        metadata = {
            "_id": faiss_id,                  
            "record_id": record_id,            
            "chunk_index": i,              
            "len": length,    
            "created_at": datetime.utcnow()
        }
        metaChunkings.append(metadata)
    store_chunking(metaChunkings)
    
def search_embeddings(query_vector, k=5):
    global index
    query_vector = np.array([query_vector], dtype='float32')
    return index.search(query_vector, k)


def save_index_to_disk(filename="faiss_index.bin"):
    """
    Lưu Faiss index vào đĩa để tái sử dụng.
    """
    try:
        faiss.write_index(index, filename)
        print(f"Đã lưu Faiss index vào {filename}.")
    except Exception as e:
        print(f"Lỗi khi lưu Faiss index: {e}")

def load_index_from_disk(filename="faiss_index.bin"):
    global index
    try:
        index = faiss.read_index(filename)
        print(f"Đã tải Faiss index từ {filename}.")
    except Exception as e:
        print(f"Lỗi khi tải Faiss index: {e}")

    print(index.ntotal)
