package main

/**
 * Source code modified from: https://blog.yessure.org/index.php/archives/203 Mod by guowanghushifu
 */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
)

var (
	appKey        = os.Getenv("APP_KEY")       // OVH的应用key
	appSecret     = os.Getenv("APP_SECRET")    // OVH的应用secret
	consumerKey   = os.Getenv("CONSUMER_KEY")  // OVH的消费者key
	region        = os.Getenv("REGION")        // 区域设置为, e.g. ovh-eu
	tgtoken       = os.Getenv("TG_TOKEN")      // 你的Telegram Bot Token
	tgchatid      = os.Getenv("TG_CHATID")     // 你希望发送消息的Telegram Chat ID
	zone          = os.Getenv("ZONE")          // OVH子公司区域设置, e.g. IE
	plancode      = os.Getenv("PLANCODE")      // 需要购买的产品的planCode, e.g. 25skleb01
	fqncode       = os.Getenv("FQN")           // 需要购买的产品的FQN, e.g. 24sk20.ram-32g-ecc-2133.softraid-2x450nvme
	optionsenv    = os.Getenv("OPTIONS")       // 选择的配置用逗号分隔, e.g. bandwidth-300-25skle,ram-32g-ecc-2400-25skle,softraid-2x450nvme-25skle
	datacenterenv = os.Getenv("DATACENTER")    // 选择的配置用逗号分隔, e.g. gra,rbx,bhs
	autopay       = os.Getenv("AUTOPAY")       // 是否自动支付, e.g. true
	frequency     = os.Getenv("FREQUENCY")     // 检查频率单位为秒, e.g. 5
	buyNum        = os.Getenv("BUYNUM")        // 一次买几个, e.g. 2
	userTag       = os.Getenv("USER_TAG")      // 用户标记，用于多个用户同时抢的时候，通知信息里面区分
	debugSw       = os.Getenv("DEBUGSW")       // 调试标记
	oneHourHigh   = os.Getenv("ONE_HOUR_HIGH") // 仅仅抢购1H-high的机器，表示大批量放机
)

var boughtNum = 0
var buyNumInt = 1
var triedTimes = 0
var lastTrySucess = 0

func Contains(arr []string, target string) bool {
	if len(arr) == 1 && arr[0] == "" {
		return true
	}

	if target == "any" {
		return true
	}

	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

func printEnvVars() {
	envVars := []struct {
		Name string
		Val  string
	}{
		{"APP_KEY", appKey},
		{"APP_SECRET", appSecret},
		{"CONSUMER_KEY", consumerKey},
		{"REGION", region},
		{"TG_TOKEN", tgtoken},
		{"TG_CHATID", tgchatid},
		{"ZONE", zone},
		{"PLANCODE", plancode},
		{"FQN", fqncode},
		{"OPTIONS", optionsenv},
		{"DATACENTER", datacenterenv},
		{"AUTOPAY", autopay},
		{"FREQUENCY", frequency},
		{"BUYNUM", buyNum},
		{"USER_TAG", userTag},
		{"ONE_HOUR_HIGH", oneHourHigh},
	}

	log.Println("**********ALL ENV**********：")
	for _, v := range envVars {
		log.Printf("%-16s = %s\n", v.Name, v.Val)
	}
	log.Println("***************************：")
}

func runTask() {

	lastTrySucess = 0
	client, err := ovh.NewClient(region, appKey, appSecret, consumerKey)
	if err != nil {
		log.Printf("Failed to create OVH client: %v\n", err)
		return
	}

	var result []map[string]interface{}
	err = client.Get("/dedicated/server/datacenter/availabilities", &result)
	if err != nil {
		log.Printf("Failed to get datacenter availabilities: %v\n", err)
		return
	}

	foundAvailable := false
	var fqn, planCode, datacenter string
	datacenterOptions := strings.Split(datacenterenv, ",")
	fqnOptions := strings.Split(fqncode, ",")

	triedTimes += 1
	log.Println("-----------------------------------------------")
	log.Printf("Number of runs: %d\n", triedTimes)
	log.Println("-----------------------------------------------")

	for _, item := range result {
		if item["planCode"] == plancode {
			fqn = item["fqn"].(string)
			planCode = item["planCode"].(string)
			datacenters := item["datacenters"].([]interface{})

			for _, dcInfo := range datacenters {
				dc := dcInfo.(map[string]interface{})
				availability := dc["availability"].(string)
				datacenter = dc["datacenter"].(string)

				if triedTimes < 6 || Contains(fqnOptions, fqn) {
					log.Printf("[FQN]: %s", fqn)
					log.Printf("[DC] : %s -/- [Avail]: %s\n", datacenter, availability)
					log.Println("------------------------")
				}

				itemAvail := false
				if oneHourHigh == "true" {
					if availability == "1H-high" {
						itemAvail = true
					}
				} else {
					if availability != "unavailable" {
						itemAvail = true
					}
				}

				if itemAvail && Contains(datacenterOptions, datacenter) && Contains(fqnOptions, fqn) {
					foundAvailable = true
					break
				}
			}

			if foundAvailable {
				log.Printf("Proceeding to next step with FQN: %s Datacenter: %s\n", fqn, datacenter)
				break
			}
		}
	}

	if !foundAvailable {
		log.Println("No record to buy")
		log.Println("***********************************************")
		return
	}

	msg_available := fmt.Sprintf("🔥 提醒用户: %s 有货啦:\n地区: %s\n型号: %s\n配置: %s\n", userTag, datacenter, plancode, fqn)
	sendTelegramMsg(tgtoken, tgchatid, msg_available)

	if debugSw == "true" {
		os.Exit(0)
	}

	log.Println("Create cart")
	var cartResult map[string]interface{}
	err = client.Post("/order/cart", map[string]interface{}{
		"ovhSubsidiary": zone,
	}, &cartResult)
	if err != nil {
		log.Printf("Failed to create cart: %v\n", err)
		return
	}

	cartID := cartResult["cartId"].(string)
	log.Printf("Cart ID: %s\n", cartID)

	log.Println("Assign cart")
	err = client.Post("/order/cart/"+cartID+"/assign", nil, nil)
	if err != nil {
		log.Printf("Failed to assign cart: %v\n", err)
		return
	}

	log.Println("Put item into cart")
	var itemResult map[string]interface{}
	err = client.Post("/order/cart/"+cartID+"/eco", map[string]interface{}{
		"planCode":    planCode,
		"pricingMode": "default",
		"duration":    "P1M",
		"quantity":    1,
	}, &itemResult)
	if err != nil {
		log.Printf("Failed to add item to cart: %v\n", err)
		return
	}

	var itemID string
	if v, ok := itemResult["itemId"].(json.Number); ok {
		itemID = v.String()
	} else if v, ok := itemResult["itemId"].(string); ok {
		itemID = v
	} else {
		log.Printf("Unexpected type for itemId, expected json.Number or string, got %T\n", itemResult["itemId"])
		return
	}

	log.Printf("Item ID: %s\n", itemID)

	log.Println("Checking required configuration")
	var requiredConfig []map[string]interface{}
	err = client.Get("/order/cart/"+cartID+"/item/"+itemID+"/requiredConfiguration", &requiredConfig)
	if err != nil {
		log.Printf("Failed to get required configuration: %v\n", err)
		return
	}

	dedicatedOs := "none_64.en"
	var regionValue string
	for _, config := range requiredConfig {
		if config["label"] == "region" {
			if allowedValues, ok := config["allowedValues"].([]interface{}); ok && len(allowedValues) > 0 {
				regionValue = allowedValues[0].(string)
			}
		}
	}

	configurations := []map[string]interface{}{
		{"label": "dedicated_datacenter", "value": datacenter},
		{"label": "dedicated_os", "value": dedicatedOs},
		{"label": "region", "value": regionValue},
	}

	for _, config := range configurations {
		log.Printf("Configure %s\n", config["label"])
		err = client.Post("/order/cart/"+cartID+"/item/"+itemID+"/configuration", map[string]interface{}{
			"label": config["label"],
			"value": config["value"],
		}, nil)
		if err != nil {
			log.Printf("Failed to configure %s: %v\n", config["label"], err)
			return
		}
	}

	log.Println("Add options")
	options := strings.Split(optionsenv, ",")

	itemIDInt, _ := strconv.Atoi(itemID)
	for _, option := range options {
		err = client.Post("/order/cart/"+cartID+"/eco/options", map[string]interface{}{
			"duration":    "P1M",
			"itemId":      itemIDInt,
			"planCode":    option,
			"pricingMode": "default",
			"quantity":    1,
		}, nil)
		if err != nil {
			log.Printf("Failed to add option %s: %v\n", option, err)
			return
		}
	}

	log.Println("Checkout")
	var checkoutResult map[string]interface{}
	err = client.Get("/order/cart/"+cartID+"/checkout", &checkoutResult)
	if err != nil {
		log.Printf("Failed to get checkout: %v\n", err)
		return
	}

	autopayValue, err := strconv.ParseBool(autopay)
	if err != nil {
		log.Println("AUTOPAY value is invalid:", err)
		return
	}

	err = client.Post("/order/cart/"+cartID+"/checkout", map[string]interface{}{
		"autoPayWithPreferredPaymentMethod": autopayValue,
		"waiveRetractationPeriod":           true,
	}, nil)
	if err != nil {
		log.Printf("Failed to checkout: %v\n", err)
		return
	}
	log.Println("Ordered!")

	boughtNum += 1
	lastTrySucess = 1

	msg_ordered := fmt.Sprintf("🎉 用户: %s 订购成功:\n地区: %s\n型号: %s\n配置: %s\n", userTag, datacenter, plancode, fqn)
	sendTelegramMsg(tgtoken, tgchatid, msg_ordered)

	if boughtNum >= buyNumInt {
		os.Exit(0)
	}
}

func sendTelegramMsg(botToken, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %v", resp.Status)
	}

	return nil
}

func main() {
	printEnvVars()
	buyNumInt, _ = strconv.Atoi(buyNum)
	freq, _ := strconv.Atoi(frequency)

	for {
		runTask()
		if lastTrySucess == 1 && debugSw == "fast" {
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(time.Duration(freq) * time.Second)
		}
	}
}
