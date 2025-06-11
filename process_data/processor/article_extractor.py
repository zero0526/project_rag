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

def extract_from_url(raw_doc):
    article = Article(url=raw_doc.url)

    try:
        article.download()
        article.parse()
    except Exception as e:
        print(f"[ERROR] Failed to download or parse article from URL {raw_doc.url}: {e}")
        return None

    if not article.text or not article.title:
        print(f"[WARNING] Article from {raw_doc.url} has empty content or title after parsing. Skipping.")
        return None

    author_name = article.authors[0] if article.authors else "unknown"

    publish_date_iso = article.publish_date.isoformat() if article.publish_date else datetime.utcnow().isoformat()

    return ExtractedDocument(
        url=raw_doc.url,
        category="", 
        title=article.title,       
        author=author_name,
        content=article.text,      
        date=publish_date_iso,
        related_links=[]           
    )


