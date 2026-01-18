#!/usr/bin/env python3
import json
import random
from datetime import datetime

class EducationSiteGenerator:
    def __init__(self):
        # 教育网站模板
        self.templates = [
            "https://www.{}.edu",
            "https://www.{}.ac.uk",
            "https://www.{}.edu.cn",
            "https://www.{}.ac.cn",
            "https://www.{}.edu.au",
            "https://www.{}.edu.tw",
            "https://www.{}.edu.hk",
            "https://www.{}.edu.sg",
            "https://www.{}.edu.mx",
            "https://www.{}.edu.br",
            "https://www.{}.edu.in",
            "https://www.{}.edu.ng",
            "https://www.{}.edu.za",
            "https://www.{}.edu.tr",
            "https://www.{}.edu.ru",
            "https://www.{}-university.edu",
            "https://www.{}-college.edu",
            "https://www.{}-institute.edu",
            "https://www.{}-academy.edu",
            "https://www.{}-school.edu"
        ]
        
        # 常见大学/教育机构名称
        self.base_names = [
            "harvard", "stanford", "mit", "princeton", "yale", "columbia", "chicago", "caltech",
            "oxford", "cambridge", "imperial", "lse", "ucl", "manchester", "birmingham", "leeds",
            "sheffield", "nottingham", "liverpool", "newcastle", "edinburgh", "glasgow", "strathclyde",
            "durham", "st-andrews", "bristol", "bath", "exeter", "essex", "reading", "southampton",
            "sussex", "warwick", "east-anglia", "lancaster", "aston", "bangor", "bournemouth", "bradford",
            "brookes", "canterbury", "coventry", "derby", "dundee", "edgehill", "falmouth", "goldsmiths",
            "greenwich", "huddersfield", "hull", "kent", "keele", "lincoln", "london-met", "loughborough",
            "manchester-met", "napier", "northumbria", "nottingham-trent", "open", "plymouth", "portsmouth",
            "roehampton", "salford", "solent", "stirling", "surrey", "swansea", "teeside", "westminster",
            "worcester", "york", "aberdeen", "aberystwyth", "anglia-ruskin", "bedfordshire", "bcu", "uel",
            "middlesex", "northumbria", "queens-belfast", "royal-holloway", "soas", "staffordshire",
            "wlv", "tsinghua", "pku", "zju", "fudan", "nju", "whu", "sjtu", "ecnu", "hust", "csu",
            "xjtu", "ntu-sg", "nus", "ubc", "toronto", "mcgill", "utoronto", "mcmaster", "ualberta", "sfu",
            "mun", "uwaterloo", "queensu", "westernu", "yorku", "concordia", "uottawa", "calgary", "saskatoon",
            "regina", "uq", "unimelb", "usyd", "mq", "unsw", "adelaide", "monash", "uwa", "curtin", "anu",
            "griffith", "jcu", "deakin", "utas", "coursera", "udemy", "khanacademy", "edx", "udacity",
            "open-edu", "saylor", "mooc", "futurelearn", "alison", "pluralsight", "codecademy", "datacamp",
            "lynda", "teachable", "thinkific", "creativelive", "masterclass", "skillshare"
        ]
        
        # 额外的名称变体
        self.name_suffixes = ["", "-1", "-2", "-3", "-north", "-south", "-east", "-west",
                            "-central", "-city", "-state", "-national", "-international"]
    
    def generate_sites(self, count=1000):
        """生成指定数量的教育网站"""
        print(f"开始生成 {count} 个教育网站...")
        
        sites = set()
        
        # 第一阶段：基础名称 + 模板
        for name in self.base_names:
            if len(sites) >= count:
                break
            for template in self.templates[:10]:  # 使用前10个模板
                if len(sites) >= count:
                    break
                site = template.format(name)
                sites.add(site)
        
        # 第二阶段：基础名称 + 后缀 + 模板
        for name in self.base_names:
            if len(sites) >= count:
                break
            for suffix in self.name_suffixes:
                if len(sites) >= count:
                    break
                for template in self.templates[10:]:  # 使用剩余模板
                    if len(sites) >= count:
                        break
                    site = template.format(name + suffix)
                    sites.add(site)
        
        # 第三阶段：随机生成更多网站
        while len(sites) < count:
            name = random.choice(self.base_names) + str(random.randint(100, 999))
            template = random.choice(self.templates)
            site = template.format(name)
            sites.add(site)
        
        # 转换为列表
        site_list = list(sites)
        print(f"成功生成 {len(site_list)} 个教育网站")
        return site_list
    
    def save_to_file(self, sites, filename="education_sites.json"):
        """保存教育网站到文件"""
        print(f"保存 {len(sites)} 个教育网站到 {filename}...")
        data = {
            "total": len(sites),
            "timestamp": datetime.now().isoformat(),
            "sites": sites
        }
        
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        
        print(f"成功保存到 {filename}")

def main():
    generator = EducationSiteGenerator()
    education_sites = generator.generate_sites(1000)
    generator.save_to_file(education_sites)
    
    print(f"\n=== 教育网站生成完成 ===")
    print(f"总共生成了 {len(education_sites)} 个教育网站")
    print(f"已保存到 education_sites.json 文件")
    print(f"前10个网站：")
    for site in education_sites[:10]:
        print(f"- {site}")

if __name__ == "__main__":
    main()
