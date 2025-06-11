import google.generativeai as genai
genai.configure(api_key="AIzaSyB42Hqv4qedgNR-8gv3DPdZUR7qtykH2C8")

model = genai.GenerativeModel("models/gemini-2.0-flash")
pattern  = """Bạn là một trợ lý thông tin đáng tin cậy, chuyên trả lời các câu hỏi dựa trên thông tin đã cung cấp.

Context:
{}

Question:
{}
Instructions:
1. Chỉ sử dụng thông tin từ phần 'Context' để trả lời câu hỏi.
2. Nếu thông tin cần thiết không có trong 'Context', hãy trả lời: "Tôi xin lỗi, tôi không thể tìm thấy thông tin này trong dữ liệu được cung cấp."
3. Trả lời đầy đủ và trực tiếp câu hỏi.
4. Tránh suy đoán hoặc thêm thông tin bên ngoài.
5. Làm mượt mà câu trả lời viết thành 1 đoạn văn 

Answer:"""

def rag(que, contexts):
    global pattern
    prompt = pattern.format(", ".join(contexts), que)
    print(prompt)
    return model.generate_content(prompt).text
