package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// 命令行参数
	output := flag.String("output", "../targets.txt", "输出文件路径")
	count := flag.Int("count", 100, "获取IP数量")
	flag.Parse()
	
	fmt.Println("====================================")
	fmt.Println("    台湾网站IP获取工具")
	fmt.Println("====================================")
	fmt.Println()
	
	// 获取绝对路径
	absOutput, err := filepath.Abs(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法解析输出路径: %v\n", err)
		os.Exit(1)
	}
	
	// 执行获取
	if err := fetchTaiwanIPs(absOutput, *count); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("\n完成!")
}

// TaiwanWebsite 台湾网站结构
type TaiwanWebsite struct {
	Domain string
	IP     string
}

// getTaiwanWebsites 获取台湾热门网站列表
func getTaiwanWebsites() []string {
	return []string{
		// 政府机构
		"www.gov.tw",
		"www.president.gov.tw",
		"www.ey.gov.tw",
		"www.mof.gov.tw",
		"www.mofa.gov.tw",
		"www.moj.gov.tw",
		"www.mnd.gov.tw",
		"www.moea.gov.tw",
		"www.moe.gov.tw",
		"www.mohw.gov.tw",
		"www.moi.gov.tw",
		"www.motc.gov.tw",
		"www.coa.gov.tw",
		"www.nsc.gov.tw",
		"www.veterans.gov.tw",
		
		// 新闻媒体
		"www.chinatimes.com",
		"www.libertytimes.com.tw",
		"www.appledaily.com.tw",
		"www.udn.com",
		"www.cna.com.tw",
		"www.tvbs.com.tw",
		"news.pts.org.tw",
		"www.ettoday.net",
		"www.setn.com",
		"www.nownews.com",
		"www.storm.mg",
		"www.thenewslens.com",
		"www.mirrormedia.mg",
		
		// 电商平台
		"www.pchome.com.tw",
		"24h.pchome.com.tw",
		"www.momoshop.com.tw",
		"www.yahoo.com.tw",
		"tw.buy.yahoo.com",
		"www.books.com.tw",
		"www.rakuten.com.tw",
		"www.friday.tw",
		"www.eslite.com",
		
		// 银行金融
		"www.cathaybk.com.tw",
		"www.ctbcbank.com",
		"www.esunbank.com.tw",
		"www.bot.com.tw",
		"www.landbank.com.tw",
		"www.tcb-bank.com.tw",
		"www.taishinbank.com.tw",
		"www.megabank.com.tw",
		"www.hncb.com.tw",
		"www.scsb.com.tw",
		
		// 电信运营商
		"www.cht.com.tw",
		"www.taiwanmobile.com",
		"www.fetnet.net",
		"www.aptg.com.tw",
		
		// 教育机构
		"www.ntu.edu.tw",
		"www.ncku.edu.tw",
		"www.nthu.edu.tw",
		"www.nctu.edu.tw",
		"www.ccu.edu.tw",
		"www.nchu.edu.tw",
		"www.nsysu.edu.tw",
		"www.cycu.edu.tw",
		"www.fcu.edu.tw",
		"www.tku.edu.tw",
		
		// 交通运输
		"www.railway.gov.tw",
		"www.thsrc.com.tw",
		"www.metro.taipei",
		"www.krtco.com.tw",
		"www.taoyuan-airport.com",
		
		// 科技公司
		"www.asus.com",
		"www.acer.com",
		"www.htc.com",
		"www.tsmc.com",
		"www.mediatek.com",
		"www.msi.com",
		"www.gigabyte.com",
		"www.asus.com.tw",
		"www.trendmicro.com",
		
		// 社交论坛
		"www.ptt.cc",
		"www.mobile01.com",
		"www.dcard.tw",
		"www.pixnet.net",
		"www.blogspot.tw",
		
		// 旅游酒店
		"www.liontravel.com",
		"www.eztravel.com.tw",
		"www.colatour.com.tw",
		"www.settour.com.tw",
		"www.hotel.com.tw",
		"www.ezfly.com",
		
		// 餐饮美食
		"www.foodpanda.com.tw",
		"www.ubereats.com/tw",
		"www.inline.app",
		
		// 求职招聘
		"www.104.com.tw",
		"www.1111.com.tw",
		"www.518.com.tw",
		"www.yes123.com.tw",
		
		// 房地产
		"www.591.com.tw",
		"www.sinyi.com.tw",
		"www.yungching.com.tw",
		"www.rakuya.com.tw",
		
		// 其他服务
		"www.cpami.gov.tw",
		"www.npa.gov.tw",
		"www.post.gov.tw",
		"www.cwb.gov.tw",
		"www.nhi.gov.tw",
	}
}

// resolveIP 解析域名到IP
func resolveIP(domain string) (string, error) {
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	ips, err := resolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		return "", err
	}
	
	if len(ips) == 0 {
		return "", fmt.Errorf("no IP found for domain: %s", domain)
	}
	
	return ips[0].String(), nil
}

// fetchTaiwanIPs 获取台湾网站IP并写入文件
func fetchTaiwanIPs(outputFile string, maxCount int) error {
	websites := getTaiwanWebsites()
	results := []TaiwanWebsite{}
	
	fmt.Printf("开始解析台湾网站IP地址...\n")
	fmt.Printf("目标数量: %d\n\n", maxCount)
	
	count := 0
	for _, domain := range websites {
		if count >= maxCount {
			break
		}
		
		fmt.Printf("[%d/%d] 正在解析: %s ... ", count+1, maxCount, domain)
		
		ip, err := resolveIP(domain)
		if err != nil {
			fmt.Printf("失败: %v\n", err)
			continue
		}
		
		fmt.Printf("成功: %s\n", ip)
		
		results = append(results, TaiwanWebsite{
			Domain: domain,
			IP:     ip,
		})
		count++
		
		time.Sleep(100 * time.Millisecond)
	}
	
	if err := saveToFile(outputFile, results); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}
	
	fmt.Printf("\n成功获取 %d 个IP地址并保存到: %s\n", len(results), outputFile)
	return nil
}

// saveToFile 保存结果到文件
func saveToFile(filename string, results []TaiwanWebsite) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	
	writer.WriteString("# 台湾网站IP地址列表\n")
	writer.WriteString(fmt.Sprintf("# 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	writer.WriteString(fmt.Sprintf("# 总数量: %d\n\n", len(results)))
	
	for _, result := range results {
		line := fmt.Sprintf("%s  # %s\n", result.IP, result.Domain)
		writer.WriteString(line)
	}
	
	return writer.Flush()
}
