package service

import (
	"fmt"
	// "log"
	"net"
	"io/ioutil"
	"net/http"
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	// "strconv"
	"time"
	"x-ui/logger"
	"x-ui/util/common"

	// tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

//This should be global variable,and only one instance
// var botInstace *tgbotapi.BotAPI

//结构体类型大写表示可以被其他包访问
type BarkService struct {
	xrayService    XrayService
	serverService  ServerService
	inboundService InboundService
	settingService SettingService
}

func (s *BarkService) BarkGetsystemStatus() string {
	var status string
	//get hostname
	name, err := os.Hostname()
	if err != nil {
		fmt.Println("get hostname error:", err)
		return ""
	}
	status = fmt.Sprintf("主机名称:%s\r\n", name)
	status += fmt.Sprintf("系统类型:%s\r\n", runtime.GOOS)
	status += fmt.Sprintf("系统架构:%s\r\n", runtime.GOARCH)
	avgState, err := load.Avg()
	if err != nil {
		logger.Warning("get load avg failed:", err)
	} else {
		status += fmt.Sprintf("系统负载:%.2f,%.2f,%.2f\r\n", avgState.Load1, avgState.Load5, avgState.Load15)
	}
	upTime, err := host.Uptime()
	if err != nil {
		logger.Warning("get uptime failed:", err)
	} else {
		status += fmt.Sprintf("运行时间:%s\r\n", common.FormatTime(upTime))
	}
	//xray version
	status += fmt.Sprintf("xray版本:%s\r\n", s.xrayService.GetXrayVersion())
	//ip address
	var ip string
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip = ipnet.IP.String()
						break
					} else {
						ip = ipnet.IP.String()
						break
					}
				}
			}
		}
	}
	status += fmt.Sprintf("IP地址:%s\r\n \r\n", ip)
	//get traffic
	inbouds, err := s.inboundService.GetAllInbounds()
	if err != nil {
		logger.Warning("StatsNotifyJob run error:", err)
	}
	for _, inbound := range inbouds {
		status += fmt.Sprintf("节点名称:%s\r\n端口:%d\r\n上行流量↑:%s\r\n下行流量↓:%s\r\n总流量:%s\r\n", inbound.Remark, inbound.Port, common.FormatTraffic(inbound.Up), common.FormatTraffic(inbound.Down), common.FormatTraffic((inbound.Up + inbound.Down)))
		if inbound.ExpiryTime == 0 {
			status += fmt.Sprintf("到期时间:无限期\r\n \r\n")
		} else {
			status += fmt.Sprintf("到期时间:%s\r\n \r\n", time.Unix((inbound.ExpiryTime/1000), 0).Format("2006-01-02 15:04:05"))
		}
	}
	return status
}




//NOTE:This function can't be called repeatly
// func (s *BarkService) BarkStopRunAndClose() {
// 	if botInstace != nil {
// 		botInstace.StopReceivingUpdates()
// 	}
// }


func (s *BarkService) BarkStartRun() {
	// logger.Info("Bark推送开始运行...")
	// s.settingService = SettingService{}



}

//bark推送消息method
func (s *BarkService) BarkPush(msg string) {
	logger.Info("Bark推送开始运行...")

	//判断bark推送状态	
	currentBarkEnable, err := s.settingService.GetBarkEnabled()
	if currentBarkEnable == true || err != nil {

		barkUrl,err := s.settingService.GetBarkUrl()
		barkToken,err := s.settingService.GetBarkToken()
	
		if err != nil || barkUrl == "" || barkToken == ""{
			logger.Warning("bark推送运行失败,请检查barkUrl 或 BarkToken。err:%v,barkUrl:%s，barkToken:%s", err, barkUrl, barkToken)
			return
		}
		
		//推送消息
		logger.Info("Bark推送消息中...")
		logger.Infof("GetbarkUrl:%s, GetBarkToken:%s", barkUrl, barkToken)
		
		barkinfo := msg
		barkInFo := make(map[string]interface{})
		barkInFo["body"] = barkinfo
		barkInFo["device_key"] = barkToken
		barkInFo["title"] = "x-ui"
		barkInFo["badge"] = 1
		barkInFo["icon"] = "https://day.app/assets/images/avatar.jpg"
		barkInFo["group"] = "x-ui"
		barkInFo["category"] = "category"
		barkInFo["sound"] = "healthnotification.caf"
  
		bytesData, err := json.Marshal(barkInFo)
		if err != nil {
			fmt.Println(err.Error() )
			return
		}
		body := bytes.NewReader(bytesData)
		url :=  barkUrl
	  // body := bytes.NewBuffer(json)
  
	  // Create client
	  client := &http.Client{}
  
	  // Create request
	  req, err := http.NewRequest("POST", url, body)
	  if err != nil {
		  fmt.Println("Failure : ", err)
	  }
  
	  // Headers
	  req.Header.Add("Content-Type", "application/json; charset=utf-8")
  
	  // Fetch Request
	  resp, err := client.Do(req)
	  
	  if err != nil {
		  fmt.Println("Failure : ", err)
	  }
  
	  // Read Response Body
	  respBody, _ := ioutil.ReadAll(resp.Body)
  
	  // Display Results
	  fmt.Println("response Status : ", resp.Status)
	  fmt.Println("response Headers : ", resp.Header)
	  fmt.Println("response Body : ", string(respBody))

	}else{
		logger.Warning("BarkPush failed,GetBarkEnabled is close:", err)
		return
	}

}