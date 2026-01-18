#!/bin/bash

echo "🧪 测试Banner引擎的各种功能"
echo "=========================="

echo "1. 测试SSH Banner:"
go run banner_engine.go -banner "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5"

echo -e "\n2. 测试Nginx Banner:"
go run banner_engine.go -banner "nginx/1.18.0"

echo -e "\n3. 测试Apache Banner:"
go run banner_engine.go -banner "Apache/2.4.41 (Ubuntu)"

echo -e "\n4. 测试MySQL Banner:"
go run banner_engine.go -banner "5.7.34-0ubuntu0.18.04.1-log mysql_native_password"

echo -e "\n5. 测试Redis Banner:"
go run banner_engine.go -banner "+PONG"

echo -e "\n6. 测试FTP Banner:"
go run banner_engine.go -banner "220 (vsFTPd 3.0.3)"

echo -e "\n7. 测试SMTP Banner:"
go run banner_engine.go -banner "220 mail.example.com ESMTP Postfix"

echo -e "\n8. 测试IIS Banner:"
go run banner_engine.go -banner "Microsoft-IIS/10.0"

echo -e "\n9. 测试未知Banner:"
go run banner_engine.go -banner "UnknownService/1.0"

echo -e "\n10. 测试JSON输出:"
go run banner_engine.go -output json -banner "nginx/1.18.0"

echo -e "\n✅ 所有测试完成!"
echo -e "\n💡 要进入交互模式，请运行:"
echo "   go run banner_engine.go -interactive"