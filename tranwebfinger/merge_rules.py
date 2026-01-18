#!/usr/bin/env python3
import json
import os

def load_json_file(file_path):
    """加载JSON文件"""
    if not os.path.exists(file_path):
        print(f"错误：文件 {file_path} 不存在")
        return None
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
        return data
    except json.JSONDecodeError as e:
        print(f"错误：解析 {file_path} 时出现JSON格式错误：{e}")
        return None
    except Exception as e:
        print(f"错误：读取 {file_path} 时出现问题：{e}")
        return None

def save_json_file(data, file_path):
    """保存JSON文件"""
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        print(f"成功将合并后的规则保存到 {file_path}")
        return True
    except Exception as e:
        print(f"错误：保存 {file_path} 时出现问题：{e}")
        return False

def merge_technologies(our_techs, wappalyzer_techs):
    """合并技术规则"""
    merged = {}
    
    # 首先添加Wappalyzer的所有技术
    for tech_name, tech_data in wappalyzer_techs.items():
        merged[tech_name] = tech_data
    
    # 然后添加或更新我们的技术
    for tech_name, tech_data in our_techs.items():
        if tech_name in merged:
            # 如果技术已存在，合并规则
            print(f"技术 {tech_name} 已存在，合并规则...")
            existing_tech = merged[tech_name]
            
            # 合并headers规则
            if 'headers' in tech_data:
                if 'headers' not in existing_tech:
                    existing_tech['headers'] = tech_data['headers']
                else:
                    for header, patterns in tech_data['headers'].items():
                        if header not in existing_tech['headers']:
                            existing_tech['headers'][header] = patterns
                        else:
                            # 合并模式，去重
                            for pattern in patterns:
                                if pattern not in existing_tech['headers'][header]:
                                    existing_tech['headers'][header].append(pattern)
            
            # 合并html规则
            if 'html' in tech_data:
                if 'html' not in existing_tech:
                    existing_tech['html'] = tech_data['html']
                else:
                    # 合并模式，去重
                    for pattern in tech_data['html']:
                        if pattern not in existing_tech['html']:
                            existing_tech['html'].append(pattern)
            
            # 合并scripts规则
            if 'scripts' in tech_data:
                if 'scripts' not in existing_tech:
                    existing_tech['scripts'] = tech_data['scripts']
                else:
                    # 合并模式，去重
                    for pattern in tech_data['scripts']:
                        if pattern not in existing_tech['scripts']:
                            existing_tech['scripts'].append(pattern)
            
            # 更新其他字段（如果我们的有值）
            if 'name' in tech_data and tech_data['name']:
                existing_tech['name'] = tech_data['name']
            if 'category' in tech_data and tech_data['category']:
                existing_tech['category'] = tech_data['category']
            if 'description' in tech_data and tech_data['description']:
                existing_tech['description'] = tech_data['description']
            if 'website' in tech_data and tech_data['website']:
                existing_tech['website'] = tech_data['website']
        else:
            # 如果技术不存在，直接添加
            print(f"添加新技术：{tech_name}")
            merged[tech_name] = tech_data
    
    return merged

def merge_categories(our_cats, wappalyzer_cats):
    """合并类别规则"""
    merged = {}
    
    # 首先添加Wappalyzer的所有类别
    for cat_name, cat_data in wappalyzer_cats.items():
        merged[cat_name] = cat_data
    
    # 然后添加或更新我们的类别
    for cat_name, cat_data in our_cats.items():
        if cat_name in merged:
            # 如果类别已存在，更新它（如果我们的优先级更高）
            existing_cat = merged[cat_name]
            if 'priority' in cat_data and 'priority' in existing_cat:
                if cat_data['priority'] < existing_cat['priority']:
                    existing_cat['priority'] = cat_data['priority']
            if 'name' in cat_data and cat_data['name']:
                existing_cat['name'] = cat_data['name']
        else:
            # 如果类别不存在，直接添加
            print(f"添加新类别：{cat_name}")
            merged[cat_name] = cat_data
    
    return merged

def merge_rules(our_file, wappalyzer_file, output_file):
    """合并两个规则文件"""
    print(f"正在加载我们的规则文件：{our_file}")
    our_data = load_json_file(our_file)
    if not our_data:
        return False
    
    print(f"正在加载Wappalyzer规则文件：{wappalyzer_file}")
    wappalyzer_data = load_json_file(wappalyzer_file)
    if not wappalyzer_data:
        return False
    
    # 合并技术规则
    print("\n开始合并技术规则...")
    merged_technologies = merge_technologies(
        our_data.get('technologies', {}),
        wappalyzer_data.get('technologies', {})
    )
    
    # 合并类别规则
    print("\n开始合并类别规则...")
    merged_categories = merge_categories(
        our_data.get('categories', {}),
        wappalyzer_data.get('categories', {})
    )
    
    # 创建合并后的规则数据
    merged_data = {
        'technologies': merged_technologies,
        'categories': merged_categories
    }
    
    # 保存合并后的规则
    print(f"\n合并完成，总共 {len(merged_technologies)} 种技术，{len(merged_categories)} 个类别")
    return save_json_file(merged_data, output_file)

def download_wappalyzer_rules(output_path):
    """下载最新的Wappalyzer规则"""
    import requests
    
    wappalyzer_url = "https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/technologies.json"
    
    print(f"正在从 {wappalyzer_url} 下载最新的Wappalyzer规则...")
    
    try:
        response = requests.get(wappalyzer_url, timeout=30)
        response.raise_for_status()
        
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(response.text)
        
        print(f"成功下载Wappalyzer规则到 {output_path}")
        return True
    except requests.exceptions.RequestException as e:
        print(f"错误：下载Wappalyzer规则失败：{e}")
        return False

def main():
    print("=== Wappalyzer规则合并工具 ===")
    
    # 默认文件路径
    our_rules_file = "technologies.json"
    wappalyzer_rules_file = "wappalyzer_technologies.json"
    output_file = "merged_technologies.json"
    
    # 询问用户是否要下载最新的Wappalyzer规则
    download_choice = input("是否要下载最新的Wappalyzer规则？(y/n): ").lower()
    
    if download_choice == 'y':
        if not download_wappalyzer_rules(wappalyzer_rules_file):
            print("无法下载Wappalyzer规则，将尝试使用本地文件")
    
    # 检查Wappalyzer规则文件是否存在
    if not os.path.exists(wappalyzer_rules_file):
        print(f"错误：Wappalyzer规则文件 {wappalyzer_rules_file} 不存在")
        print("请确保您已下载最新的Wappalyzer规则，或指定正确的文件路径")
        return
    
    # 检查我们的规则文件是否存在
    if not os.path.exists(our_rules_file):
        print(f"错误：我们的规则文件 {our_rules_file} 不存在")
        return
    
    # 执行合并
    if merge_rules(our_rules_file, wappalyzer_rules_file, output_file):
        print("\n规则合并成功！")
        print(f"合并后的规则文件：{output_file}")
        print("您可以将此文件用于Wappalyzer或其他兼容工具")
    else:
        print("\n规则合并失败！")

if __name__ == "__main__":
    main()