from utils.redis_client import redis_client
from processor.article_extractor import extract_from_html, extract_from_url
from processor.chunking import chunk_text
from embedding.embedder import embed_chunks
from storage.faiss_store import save_embeddings, save_index_to_disk, load_index_from_disk
from storage.mongodb import store_metadata
from models.raw_document import RawDocument
from models.extracted_document import ExtractedDocument
import time, json

CHECK_INTERVAL = 1
MAX_PROCESSED_BUFFER = 15
try:
    load_index_from_disk()
    while True:
        raw_json = redis_client.lpop("raw_html_list")
        if raw_json:
            try:
                raw_doc = RawDocument.model_validate_json(raw_json)
                extracted_doc = extract_from_html(raw_doc)
                if extracted_doc:
                    redis_client.rpush("extracted_list", extracted_doc.model_dump_json())
            except Exception as e:
                print(f"[ERROR] Cannot process: {e}")
        current_len = redis_client.llen("extracted_list")
        if current_len >= MAX_PROCESSED_BUFFER:
            try:
                pipe = redis_client.pipeline(transaction=True)
                pipe.lrange("extracted_list", 0, MAX_PROCESSED_BUFFER-1)
                pipe.ltrim("extracted_list", MAX_PROCESSED_BUFFER, -1)
                batch_json, _ = pipe.execute()

                documents = [ExtractedDocument.model_validate_json(j) for j in batch_json]

                for doc in documents:
                    if not doc.content:
                        doc = extract_from_url(doc)
                        if not doc.content:
                            continue
                    chunks = chunk_text(doc.content)
                    doc.chunks = chunks
                    vectors = embed_chunks(chunks)

                    record_id = store_metadata(doc, chunks)
                    save_embeddings(vectors, record_id)

            except Exception as e:
                print(f"Error: {e}")
        time.sleep(CHECK_INTERVAL)

except KeyboardInterrupt:
    print("Interrupted by user. Saving FAISS index...")
    save_index_to_disk()
    print("Index saved. Exiting.")


