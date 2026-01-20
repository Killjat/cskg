#!/bin/bash

# Elasticsearch索引创建脚本
# 创建CyberStroll所需的索引和映射

echo "等待Elasticsearch服务启动..."
sleep 30

# Elasticsearch URL
ES_URL="http://localhost:9200"

# 检查Elasticsearch是否就绪
echo "检查Elasticsearch服务状态..."
curl -f $ES_URL/_cluster/health

if [ $? -eq 0 ]; then
    echo "Elasticsearch服务已就绪，开始创建索引..."
    
    # 创建cyberstroll_ip_scan索引
    echo "创建索引: cyberstroll_ip_scan"
    curl -X PUT "$ES_URL/cyberstroll_ip_scan" \
        -H "Content-Type: application/json" \
        -d '{
            "settings": {
                "number_of_shards": 3,
                "number_of_replicas": 0,
                "refresh_interval": "5s",
                "max_result_window": 50000
            },
            "mappings": {
                "properties": {
                    "ip": {
                        "type": "ip"
                    },
                    "port": {
                        "type": "integer"
                    },
                    "protocol": {
                        "type": "keyword"
                    },
                    "service": {
                        "type": "keyword"
                    },
                    "banner": {
                        "type": "text",
                        "analyzer": "standard"
                    },
                    "status": {
                        "type": "keyword"
                    },
                    "response_time": {
                        "type": "integer"
                    },
                    "country": {
                        "type": "keyword"
                    },
                    "city": {
                        "type": "keyword"
                    },
                    "organization": {
                        "type": "keyword"
                    },
                    "timestamp": {
                        "type": "date",
                        "format": "strict_date_optional_time||epoch_millis"
                    },
                    "scan_time": {
                        "type": "date",
                        "format": "strict_date_optional_time||epoch_millis"
                    },
                    "task_id": {
                        "type": "keyword"
                    },
                    "fingerprint": {
                        "type": "nested",
                        "properties": {
                            "technology": {
                                "type": "keyword"
                            },
                            "version": {
                                "type": "keyword"
                            },
                            "confidence": {
                                "type": "integer"
                            }
                        }
                    },
                    "enrichment": {
                        "type": "nested",
                        "properties": {
                            "certificate": {
                                "type": "object",
                                "properties": {
                                    "subject": {"type": "keyword"},
                                    "issuer": {"type": "keyword"},
                                    "valid_from": {"type": "date"},
                                    "valid_to": {"type": "date"},
                                    "dns_names": {"type": "keyword"}
                                }
                            },
                            "website_info": {
                                "type": "object",
                                "properties": {
                                    "title": {"type": "text"},
                                    "description": {"type": "text"},
                                    "keywords": {"type": "keyword"},
                                    "language": {"type": "keyword"}
                                }
                            },
                            "api_info": {
                                "type": "object",
                                "properties": {
                                    "endpoints": {"type": "keyword"},
                                    "methods": {"type": "keyword"},
                                    "framework": {"type": "keyword"}
                                }
                            },
                            "content": {
                                "type": "object",
                                "properties": {
                                    "status_code": {"type": "integer"},
                                    "content_type": {"type": "keyword"},
                                    "content_length": {"type": "integer"},
                                    "headers": {"type": "object"},
                                    "body_hash": {"type": "keyword"}
                                }
                            },
                            "enriched_at": {
                                "type": "date",
                                "format": "strict_date_optional_time||epoch_millis"
                            }
                        }
                    }
                }
            }
        }'
    
    echo ""
    echo "创建索引模板: cyberstroll_template"
    curl -X PUT "$ES_URL/_index_template/cyberstroll_template" \
        -H "Content-Type: application/json" \
        -d '{
            "index_patterns": ["cyberstroll_*"],
            "priority": 1,
            "template": {
                "settings": {
                    "number_of_shards": 3,
                    "number_of_replicas": 0,
                    "refresh_interval": "5s"
                }
            }
        }'
    
    echo ""
    echo "检查索引状态:"
    curl -X GET "$ES_URL/_cat/indices/cyberstroll_*?v"
    
    echo ""
    echo "检查索引映射:"
    curl -X GET "$ES_URL/cyberstroll_ip_scan/_mapping?pretty"
    
    echo ""
    echo "Elasticsearch索引创建完成!"
else
    echo "错误: Elasticsearch服务未就绪，请检查服务状态"
    exit 1
fi