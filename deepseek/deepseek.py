import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ.get('sk-c29c015340d34eb08d38a7bf8a3800f0'),
    base_url="https://api.deepseek.com")

class WapalyzerFingerprintTrainer:
    def __init__(self):
        self.deepseek_model = None  # DeepSeek 模型实例
        
    def extract_features(self, website_data):
        """
        提取网站的特征用于指纹识别
        """
        features = {
            'headers': website_data.get('headers', {}),
            'html_content': website_data.get('html_content', ''),
            'js_libraries': website_data.get('js_libraries', []),
            'css_properties': website_data.get('css_properties', [])
        }
        return features
    
    def train_with_deepseek(self, training_data):
        """
        使用 DeepSeek 模型训练指纹识别
        """
        # 将特征转换为适合 DeepSeek 模型的格式
        processed_data = self.preprocess_data(training_data)
        
        # 使用 DeepSeek 进行训练
        trained_model = self.deepseek_model.train(processed_data)
        
        return trained_model
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[
        {"role": "system", "content": "You are a helpful assistant"},
        {"role": "user", "content": "Hello"},
    ],
    stream=False
)

print(response.choices[0].message.content)