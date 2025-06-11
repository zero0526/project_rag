from api.query import app
from storage.faiss_store import load_index_from_disk

load_index_from_disk()
app.run(debug=True, port=5001)
