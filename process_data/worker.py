from utils.redis_client import redis_client
from processor.article_extractor import extract_from_html
from processor.chunking import chunk_text
from embedding.embedder import embed_chunks
from storage.faiss_store import save_embeddings, save_index_to_disk
from storage.mongodb import store_metadata
from models.raw_document import RawDocument
from models.extracted_document import ExtractedDocument
import time, json

CHECK_INTERVAL = 1
MAX_PROCESSED_BUFFER = 15
try:
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
                batch_json = redis_client.lrange("extracted_list", 0, -1)
                redis_client.delete("extracted_list")
                documents = []
                for doc_json in batch_json:
                    try:
                        doc = ExtractedDocument.model_validate_json(doc_json)
                        documents.append(doc)
                    except Exception as e:
                        print(f"Error: {e}")

                for doc in documents:
                    chunks = chunk_text(doc.content)
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


