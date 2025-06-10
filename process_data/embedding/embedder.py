from sentence_transformers import SentenceTransformer
# 768
model_name = "bkai-foundation-models/vietnamese-bi-encoder"
model = SentenceTransformer(model_name)

def embed_chunks(chunks):
    return model.encode(chunks, convert_to_tensor=False).tolist()
