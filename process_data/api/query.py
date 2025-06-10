from utils.gemini_client import rag
from flask import Flask, request, jsonify
from storage.query import getContext
from utils.gemini_client import rag
app = Flask(__name__)

@app.route('/ask', methods=['POST'])
def ask():
    data = request.get_json()
    query = data.get('query')
    if not query:
        return jsonify({'error': 'Missing query'}), 400

    contexts = getContext(query)
    answer = rag(que=query, contexts= contexts)
    print(answer)
    return jsonify({'answer': answer})
