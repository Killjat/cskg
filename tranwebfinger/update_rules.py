#!/usr/bin/env python3
import json
import os
import requests
import logging
from datetime import datetime

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('rule_updates.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# 文件路径配置
OUR_RULES_FILE = 'technologies.json'
WAPPALYZER_RULES_FILE = 'wappalyzer_technologies.json'
MERGED_RULES_FILE = 'merged_technologies.json'
CONFIG_FILE = 'config.json'

# Wappalyzer规则URL
WAPPALYZER_RULES_URL = 'https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/technologies.json'

def load_json_file(file_path):
    """加载JSON文件"""
    if not os.path.exists(file_path):
        logger.error(f"文件 {file_path} 不存在")
        return None
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
        return data
    except json.JSONDecodeError as e:
        logger.error(f"解析 {file_path} 时出现JSON格式错误：{e}")
        return None
    except Exception as e:
        logger.error(f"读取 {file_path} 时出现问题：{e}")
        return None

def save_json_file(data, file_path):
    """保存JSON文件"""
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        logger.info(f"成功将数据保存到 {file_path}")
        return True
    except Exception as e:
        logger.error(f"保存 {file_path} 时出现问题：{e}")
        return False

def download_wappalyzer_rules():
    """下载最新的Wappalyzer规则"""
    logger.info(f"正在从 {WAPPALYZER_RULES_URL} 下载最新的Wappalyzer规则...")
    
    try:
        response = requests.get(WAPPALYZER_RULES_URL, timeout=30)
        response.raise_for_status()
        
        with open(WAPPALYZER_RULES_FILE, 'w', encoding='utf-8') as f:
            f.write(response.text)
        
        logger.info(f"成功下载Wappalyzer规则到 {WAPPALYZER_RULES_FILE}")
        return True
    except requests.exceptions.RequestException as e:
        logger.error(f"下载Wappalyzer规则失败：{e}")
        return False

def merge_technologies(our_techs, wappalyzer_techs):
    """合并技术规则"""
    merged = {}
    added = 0
    updated = 0
    
    # 首先添加Wappalyzer的所有技术
    for tech_name, tech_data in wappalyzer_techs.items():
        merged[tech_name] = tech_data
    
    # 然后添加或更新我们的技术
    for tech_name, tech_data in our_techs.items():
        if tech_name in merged:
            # 如果技术已存在，合并规则
            updated += 1
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
            added += 1
            merged[tech_name] = tech_data
    
    logger.info(f"合并完成：添加了 {added} 种技术，更新了 {updated} 种技术")
    return merged

def merge_categories(our_cats, wappalyzer_cats):
    """合并类别规则"""
    merged = {}
    added = 0
    updated = 0
    
    # 首先添加Wappalyzer的所有类别
    for cat_name, cat_data in wappalyzer_cats.items():
        merged[cat_name] = cat_data
    
    # 然后添加或更新我们的类别
    for cat_name, cat_data in our_cats.items():
        if cat_name in merged:
            # 如果类别已存在，更新它
            updated += 1
            existing_cat = merged[cat_name]
            if 'priority' in cat_data and 'priority' in existing_cat:
                if cat_data['priority'] < existing_cat['priority']:
                    existing_cat['priority'] = cat_data['priority']
            if 'name' in cat_data and cat_data['name']:
                existing_cat['name'] = cat_data['name']
        else:
            # 如果类别不存在，直接添加
            added += 1
            merged[cat_name] = cat_data
    
    logger.info(f"合并完成：添加了 {added} 个类别，更新了 {updated} 个类别")
    return merged

def merge_rules():
    """合并规则"""
    # 加载我们的规则
    our_data = load_json_file(OUR_RULES_FILE)
    if not our_data:
        return False
    
    # 加载Wappalyzer规则
    wappalyzer_data = load_json_file(WAPPALYZER_RULES_FILE)
    if not wappalyzer_data:
        return False
    
    # 合并技术规则
    logger.info("开始合并技术规则...")
    merged_technologies = merge_technologies(
        our_data.get('technologies', {}),
        wappalyzer_data.get('technologies', {})
    )
    
    # 合并类别规则
    logger.info("开始合并类别规则...")
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
    logger.info(f"合并完成，总共 {len(merged_technologies)} 种技术，{len(merged_categories)} 个类别")
    return save_json_file(merged_data, MERGED_RULES_FILE)

def update_main_rules():
    """更新主规则文件"""
    # 备份当前主规则文件
    backup_file = f"technologies_backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    try:
        import shutil
        shutil.copy2(OUR_RULES_FILE, backup_file)
        logger.info(f"已备份当前主规则文件到 {backup_file}")
    except Exception as e:
        logger.error(f"备份主规则文件失败：{e}")
        return False
    
    # 复制合并后的规则到主规则文件
    try:
        shutil.copy2(MERGED_RULES_FILE, OUR_RULES_FILE)
        logger.info(f"已将合并后的规则更新到主规则文件 {OUR_RULES_FILE}")
        return True
    except Exception as e:
        logger.error(f"更新主规则文件失败：{e}")
        # 恢复备份
        try:
            shutil.copy2(backup_file, OUR_RULES_FILE)
            logger.info(f"已恢复主规则文件从备份 {backup_file}")
        except Exception as e2:
            logger.error(f"恢复主规则文件失败：{e2}")
        return False

def main():
    """主函数"""
    logger.info("=== 开始更新规则 ===")
    
    # 下载最新的Wappalyzer规则
    if not download_wappalyzer_rules():
        logger.error("无法下载Wappalyzer规则，更新失败")
        return False
    
    # 合并规则
    if not merge_rules():
        logger.error("合并规则失败，更新失败")
        return False
    
    # 更新主规则文件
    if not update_main_rules():
        logger.error("更新主规则文件失败")
        return False
    
    logger.info("=== 规则更新成功 ===")
    return True

if __name__ == "__main__":
    main()