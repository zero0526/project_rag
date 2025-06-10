from newspaper import Article

def process_article(url):
    try:
        article = Article(url)
        article.download()
        article.parse()
        return {
            "url": url,
            "title": article.title,
            "text": article.text
        }
    except Exception as e:
        print(f"Error processing {url}: {e}")
        return None
