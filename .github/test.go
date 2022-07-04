package main

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
)

func main() {
	client, err := sdk.NewClientWithAccessKey("cn-shenzhen", "LTAIIPSQM9vK0iec", "fUuOI5Mc5NXuUgscINMIkjhSxQfyoj")
	/* use STS Token
	client, err := sdk.NewClientWithStsToken("cn-shenzhen", "<your-access-key-id>", "<your-access-key-secret>", "<your-sts-token>")
	*/
	if err != nil {
		panic(err)
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "swas.cn-shenzhen.aliyuncs.com"
	request.Version = "2020-06-01"
	request.ApiName = "ListFirewallRules"
	request.QueryParams["InstanceId"] = "854d0d200a7745a3b5149b44eeb4e6ce"
	request.QueryParams["RegionId"] = "cn-hongkong"
	request.QueryParams["PageSize"] = "10"
	request.QueryParams["PageNumber"] = "1"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		panic(err)
	}
	fmt.Print(response.GetHttpContentString())
}
