import string
import unicodedata

def get_overlap_suffix_prefix(a_clean: str, b_clean: str, max_overlap: int = 100) -> str:        
    max_len = min(len(a_clean), len(b_clean), max_overlap)
    
    for mid in range(max_len, 0, -1):
        if a_clean[-mid:] == b_clean[:mid]:
            return a_clean + b_clean[mid:]
    return ""

a = "hỗ trợ hiệu suất nâng cao thông qua"
b = "hiệu suất nâng cao thông qua các công cụ AI"

print(get_overlap_suffix_prefix(a, b))  

