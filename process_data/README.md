contextual_retrieval/
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
├── main.py                         # Entry point: chạy vòng lặp xử lý chính
│
├── models/
│   ├── raw_document.py            # RawDocument model
│   └── extracted_document.py      # ExtractedDocument model
│
├── utils/
│   ├── redis_client.py            # Kết nối Redis, push/pop list
│   └── logging_config.py          # (Tùy chọn) cấu hình logging
│
├── processor/
│   ├── article_extractor.py       # Hàm extract_from_html()
│   └── chunking.py                # Hàm chunk_text()
│
├── embedding/
│   └── embedder.py                # Hàm embed_chunks(chunks)
│
├── storage/
│   ├── faiss_store.py             # Lưu vector vào FAISS
│   └── mongodb.py                 # Lưu metadata vào MongoDB
│
└── config.py                      # (Tùy chọn) biến cấu hình hệ thống
