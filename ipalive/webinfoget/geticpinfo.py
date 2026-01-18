import requests
from bs4 import BeautifulSoup
from urllib.parse import urljoin, urlparse
import re
from typing import Optional, List

# 可选：处理动态加载的备案信息（需安装 selenium）
try:
    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    SELENIUM_AVAILABLE = True
except ImportError:
    SELENIUM_AVAILABLE = False

def clean_text(text: str) -> str:
    """清理文本：去除多余空格、制表符、换行，统一全角/半角"""
    if not text:
        return ""
    # 去除多余空白字符
    text = re.sub(r"\s+", "", text)
    # 全角转半角
    text = text.translate(str.maketrans('０１２３４５６７８９', '0123456789'))
    return text

def get_page_text(target_url: str, use_selenium: bool = False) -> Optional[str]:
    """
    获取网页文本（优先静态爬取，失败/指定时用Selenium处理动态加载）
    :param target_url: 目标URL
    :param use_selenium: 是否使用Selenium
    :return: 清理后的网页文本
    """
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        "Accept-Encoding": "gzip, deflate",
        "Connection": "keep-alive"
    }

    # 1. 静态爬取
    if not use_selenium:
        try:
            response = requests.get(
                target_url,
                headers=headers,
                timeout=10,
                allow_redirects=True
            )
            response.raise_for_status()
            response.encoding = response.apparent_encoding
            soup = BeautifulSoup(response.text, "html.parser")
            
            # 优先提取备案相关标签（常见的备案容器）
            beian_tags = soup.find_all(
                ["footer", "div", "p"],
                attrs={
                    "class": re.compile(r"beian|备案|footer", re.IGNORECASE),
                    "id": re.compile(r"beian|备案|footer", re.IGNORECASE)
                }
            )
            # 合并备案标签文本 + 全页文本（双重保障）
            beian_text = "".join([tag.get_text() for tag in beian_tags])
            full_text = soup.get_text()
            total_text = beian_text + full_text
            return clean_text(total_text)
        except Exception as e:
            print(f"静态爬取失败：{e}")
            if SELENIUM_AVAILABLE:
                print("尝试使用Selenium动态爬取...")
            else:
                return None

    # 2. 动态爬取（Selenium）
    if SELENIUM_AVAILABLE:
        try:
            chrome_options = Options()
            chrome_options.add_argument("--headless")  # 无头模式（不显示浏览器）
            chrome_options.add_argument("--no-sandbox")
            chrome_options.add_argument("--disable-dev-shm-usage")
            driver = webdriver.Chrome(options=chrome_options)
            driver.set_page_load_timeout(15)
            driver.get(target_url)
            # 优先提取备案标签文本
            beian_elements = driver.find_elements(
                "xpath",
                "//*[contains(@class, 'beian') or contains(@id, 'beian') or contains(text(), '备案') or self::footer]"
            )
            beian_text = "".join([elem.text for elem in beian_elements])
            full_text = driver.page_source
            soup = BeautifulSoup(full_text, "html.parser")
            total_text = beian_text + soup.get_text()
            driver.quit()
            return clean_text(total_text)
        except Exception as e:
            print(f"Selenium爬取失败：{e}")
            return None
    return None

def extract_icp_beian(text: str) -> dict:
    """
    精准提取ICP备案号、公安备案号
    :param text: 清理后的网页文本
    :return: 备案信息字典
    """
    result = {
        "icp_record": [],       # ICP备案号（去重）
        "police_record": [],    # 公安备案号（去重）
        "record_owner": None    # 备案主体
    }

    # 1. 优化后的ICP备案号正则（兼容空格、分隔符、旧号段）
    # 匹配规则：[省市简称] + 任意字符 + ICP备 + 数字 + 号 + 可选后缀
    icp_pattern = re.compile(
        r"(京|沪|粤|苏|浙|鲁|川|渝|津|冀|晋|蒙|辽|吉|黑|皖|闽|赣|豫|鄂|湘|桂|琼|贵|云|陕|甘|青|宁|新|港|澳|台)"
        r".*?ICP备.*?(\d{6,8})(?:号)?(?:-(\d+))?",
        re.IGNORECASE
    )
    icp_matches = icp_pattern.findall(text)
    # 格式化ICP备案号（统一格式：省市+ICP备+数字+号+后缀）
    for match in icp_matches:
        province = match[0]
        num = match[1]
        suffix = match[2] if match[2] else ""
        icp_no = f"{province}ICP备{num}号"
        if suffix:
            icp_no += f"-{suffix}"
        result["icp_record"].append(icp_no.upper())

    # 2. 优化后的公安备案号正则（兼容空格、不同位数）
    police_pattern = re.compile(
        r"(京|沪|粤|苏|浙|鲁|川|渝|津|冀|晋|蒙|辽|吉|黑|皖|闽|赣|豫|鄂|湘|桂|琼|贵|云|陕|甘|青|宁|新|港|澳|台)"
        r".*?公网安备.*?(\d{6,12})(?:号)?",
        re.IGNORECASE
    )
    police_matches = police_pattern.findall(text)
    # 格式化公安备案号
    for match in police_matches:
        province = match[0]
        num = match[1]
        police_no = f"{province}公网安备{num}号"
        result["police_record"].append(police_no.upper())

    # 3. 优化备案主体提取（兼容更多格式）
    owner_pattern = re.compile(
        r"(?:主办单位|网站主办者|版权所有|©).*?:?([^，。；！？]{2,50})",
        re.IGNORECASE
    )
    owner_matches = owner_pattern.findall(text)
    if owner_matches:
        # 过滤无效主体，保留企业/个人名称
        valid_owners = [
            owner for owner in owner_matches
            if not re.match(r"^\d+$", owner) and len(owner) > 2
        ]
        if valid_owners:
            result["record_owner"] = valid_owners[0].strip()

    # 去重
    result["icp_record"] = list(set(result["icp_record"]))
    result["police_record"] = list(set(result["police_record"]))
    return result

def extract_webpage_icp_info(target_url: str, use_selenium: bool = False) -> dict:
    """
    提取目标网页备案信息（主函数）
    :param target_url: 目标URL
    :param use_selenium: 是否使用Selenium处理动态加载
    :return: 最终结果
    """
    final_result = {
        "icp_record": [],
        "police_record": [],
        "record_owner": None,
        "error": None
    }

    # 1. 获取网页文本
    page_text = get_page_text(target_url, use_selenium)
    if not page_text:
        final_result["error"] = "网页文本提取失败"
        return final_result

    # 2. 提取备案信息
    beian_info = extract_icp_beian(page_text)
    final_result.update(beian_info)
    return final_result

# ===================== 示例调用 =====================
if __name__ == "__main__":
    # 测试URL（替换为你的目标URL）
    target_url = "https://www.tipray.com/product_cont2.php?id=196https://www.tipray.com/&sdclkid=ALf615fsbrDNbJDzb_&bd_vid=9484997432240655851"
    
    # 第一步：静态提取（优先）
    icp_info = extract_webpage_icp_info(target_url)
    
    # 第二步：若静态提取不到，尝试动态提取（需安装Selenium）
    if not icp_info["icp_record"] and not icp_info["police_record"] and SELENIUM_AVAILABLE:
        icp_info = extract_webpage_icp_info(target_url, use_selenium=True)

    # 格式化输出
    print("=" * 60)
    print(f"目标URL：{target_url}")
    print("=" * 60)
    if icp_info["error"]:
        print(f"❌ 错误：{icp_info['error']}")
    else:
        print(f"🌐 ICP备案号：{', '.join(icp_info['icp_record']) or '无'}")
        print(f"🚨 公安备案号：{', '.join(icp_info['police_record']) or '无'}")
        print(f"🏢 备案主体：{icp_info['record_owner'] or '无'}")
    print("=" * 60)