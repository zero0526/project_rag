from utils.gemini_client import rag
from flask import Flask, request, jsonify
from storage.query import getContext
from utils.gemini_client import rag
from flask_cors import CORS

app = Flask(__name__)
CORS(app) 
@app.route('/ask', methods=['POST'])
def ask():
    data = request.get_json()
    query = data.get('query')
    print(query)
    if not query:
        return jsonify({'error': 'Missing query'}), 400

    contexts = getContext(query)
    answer = rag(que=query, contexts= contexts)
    return jsonify({'answer': answer})
