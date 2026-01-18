
import requests
from bs4 import BeautifulSoup
import urllib.parse
import time
import json

class SimilarWebsiteSearcher:
    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        })
    
    def search_similar_sites(self, target_url, max_results=20):
        """
        通过搜索引擎查找与目标网站相似的网站
        """
        # 移除协议部分
        clean_url = target_url.replace('http://', '').replace('https://', '').split('/')[0]
        results = []
        
        # 使用多个搜索引擎查询
        search_engines = [
            self._search_google_similar,
            self._search_bing_related,
        ]
        
        for search_func in search_engines:
            try:
                engine_results = search_func(clean_url, max_results//len(search_engines))
                results.extend(engine_results)
                time.sleep(1)  # 避免请求过快
            except Exception as e:
                print(f"搜索引擎查询出错: {e}")
                continue
        
        # 去重并返回结果
        unique_results = []
        seen_urls = set()
        for item in results:
            if item['url'] not in seen_urls:
                unique_results.append(item)
                seen_urls.add(item['url'])
        
        return unique_results[:max_results]
    
    def _search_google_similar(self, domain, max_results):
        """
        使用Google查找相似网站
        """
        query = f"related:{domain}"
        search_url = f"https://www.google.com/search?q={urllib.parse.quote(query)}&num={max_results}"
        
        response = self.session.get(search_url)
        soup = BeautifulSoup(response.text, 'html.parser')
        
        results = []
        # 解析搜索结果
        for g in soup.find_all('div', class_='g')[:max_results]:
            anchor = g.find('a')
            if anchor and anchor.get('href'):
                title_elem = g.find('h3')
                title = title_elem.text if title_elem else "无标题"
                url = anchor.get('href')
                desc_elem = g.find('span', class_='st') or g.find('div', class_='VwiC3b')
                description = desc_elem.text if desc_elem else "无描述"
                
                # 过滤掉非网站链接
                if url.startswith('/url?q='):
                    url = urllib.parse.unquote(url.split('/url?q=')[1].split('&')[0])
                
                if url.startswith('http'):
                    results.append({
                        'title': title,
                        'url': url,
                        'description': description,
                        'engine': 'Google'
                    })
        
        return results
    
    def _search_bing_related(self, domain, max_results):
        """
        使用Bing查找相关网站
        """
        query = f"related:{domain}"
        search_url = f"https://www.bing.com/search?q={urllib.parse.quote(query)}&count={max_results}"
        
        response = self.session.get(search_url)
        soup = BeautifulSoup(response.text, 'html.parser')
        
        results = []
        # 解析Bing搜索结果
        for li in soup.find_all('li', class_='b_algo')[:max_results]:
            h2 = li.find('h2')
            if h2:
                anchor = h2.find('a')
                if anchor and anchor.get('href'):
                    title = anchor.text
                    url = anchor.get('href')
                    desc_div = li.find('p') or li.find('div', class_='b_caption')
                    description = desc_div.text if desc_div else "无描述"
                    
                    results.append({
                        'title': title,
                        'url': url,
                        'description': description,
                        'engine': 'Bing'
                    })
        
        return results

def main():
    searcher = SimilarWebsiteSearcher()
    
    # 获取用户输入
    target_website = input("请输入要查找相似网站的网址(例如: example.com): ").strip()
    
    if not target_website:
        print("请输入有效的网站地址")
        return
    
    print(f"\n正在查找与 {target_website} 相似的网站...")
    print("-" * 50)
    
    try:
        results = searcher.search_similar_sites(target_website)
        
        if not results:
            print("未找到相似网站或搜索失败")
            return
        
        print(f"找到 {len(results)} 个相似网站:\n")
        
        for i, site in enumerate(results, 1):
            print(f"{i}. {site['title']}")
            print(f"   网址: {site['url']}")
            print(f"   描述: {site['description']}")
            print(f"   来源: {site['engine']}")
            print()
            
        # 保存结果到文件
        output_file = f"similar_sites_{target_website.replace('http://', '').replace('https://', '').replace('/', '_')}.json"
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(results, f, ensure_ascii=False, indent=2)
        
        print(f"结果已保存到 {output_file}")
        
    except Exception as e:
        print(f"搜索过程中出现错误: {e}")

if __name__ == "__main__":
    main()
