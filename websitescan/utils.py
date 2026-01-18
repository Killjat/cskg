#!/usr/bin/env python3

def load_targets(file_path):
    """从文件中加载目标网站列表"""
    targets = []
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#'):
                    # 确保URL格式正确
                    if not line.startswith(('http://', 'https://')):
                        line = f'https://{line}'
                    targets.append(line)
    except Exception as e:
        print(f"加载目标文件失败: {e}")
    
    return targets
