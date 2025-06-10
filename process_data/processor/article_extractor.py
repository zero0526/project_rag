from newspaper import Article
from datetime import datetime
from models.extracted_document import ExtractedDocument

def extract_from_html(raw_doc):
    article = Article(url=raw_doc.url)
    article.set_html(raw_doc.raw_html)

    try:
        article.parse()
    except Exception as e:
        print(f"[ERROR] Failed to parse article: {e}")
        return None

    return ExtractedDocument(
        url=raw_doc.url,
        category=raw_doc.category,
        title=article.title or "",
        author=article.authors[0] if article.authors else "unknown",
        content=article.text or "",
        date=article.publish_date.isoformat() if article.publish_date else datetime.utcnow().isoformat(),
        related_links=[]  
    )
