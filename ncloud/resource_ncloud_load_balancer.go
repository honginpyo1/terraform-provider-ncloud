package ncloud

import (
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	"github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceNcloudLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNcloudLoadBalancerCreate,
		Read:   resourceNcloudLoadBalancerRead,
		Delete: resourceNcloudLoadBalancerDelete,
		Update: resourceNcloudLoadBalancerUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(DefaultCreateTimeout),
			Delete: schema.DefaultTimeout(DefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"load_balancer_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringLengthInRange(3, 30),
				Description:  "Name of a load balancer to create. Default: Automatically specified by Ncloud.",
			},
			"load_balancer_algorithm_type_code": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIncludeValues([]string{"RR", "LC", "SIPHS"}),
				Description:  "Load balancer algorithm type code. The available algorithms are as follows: [ROUND ROBIN (RR) | LEAST_CONNECTION (LC)]. Default: ROUND ROBIN (RR)",
			},
			"load_balancer_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringLengthInRange(1, 1000),
				Description:  "Description of a load balancer to create",
			},
			"load_balancer_rule_list": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        loadBalancerRuleSchemaResource,
				Description: "Load balancer rules are required to create a load balancer.",
			},
			"server_instance_no_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of server instance numbers to be bound to the load balancer",
			},
			"internet_line_type_code": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIncludeValues([]string{"PUBLC", "GLBL"}),
				Description:  "Internet line identification code. PUBLC(Public), GLBL(Global). default : PUBLC(Public)",
			},
			"network_usage_type_code": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateIncludeValues([]string{"PBLIP", "PRVT"}),
				Description:  "Network usage identification code. PBLIP(PublicIP), PRVT(PrivateIP). default : PBLIP(PublicIP)",
			},
			"region_no": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Get available values using the getRegionList action.",
			},
			"load_balancer_instance_no": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtual_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_algorithm_type": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     commonCodeSchemaResource,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"internet_line_type": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     commonCodeSchemaResource,
			},
			"load_balancer_instance_status_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_instance_status": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     commonCodeSchemaResource,
			},
			"load_balancer_instance_operation": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     commonCodeSchemaResource,
			},
			"network_usage_type": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     commonCodeSchemaResource,
			},
			"is_http_keep_alive": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"connection_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"certificate_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balanced_server_instance_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceNcloudLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NcloudSdk).conn

	reqParams := buildCreateLoadBalancerInstanceParams(d)
	resp, err := conn.CreateLoadBalancerInstance(reqParams)
	if err != nil {
		logErrorResponse("CreateLoadBalancerInstance", err, reqParams)
		return err
	}
	logCommonResponse("CreateLoadBalancerInstance", reqParams, resp.CommonResponse)

	LoadBalancerInstance := &resp.LoadBalancerInstanceList[0]
	d.SetId(LoadBalancerInstance.LoadBalancerInstanceNo)

	if err := waitForLoadBalancerInstance(conn, LoadBalancerInstance.LoadBalancerInstanceNo, "USED", DefaultCreateTimeout); err != nil {
		return err
	}
	return resourceNcloudLoadBalancerRead(d, meta)
}

func resourceNcloudLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NcloudSdk).conn

	lb, err := getLoadBalancerInstance(conn, d.Id())
	if err != nil {
		return err
	}
	if lb != nil {
		d.Set("virtual_ip", lb.VirtualIP)
		d.Set("load_balancer_name", lb.LoadBalancerName)
		d.Set("load_balancer_algorithm_type", map[string]interface{}{
			"code":      lb.LoadBalancerAlgorithmType.Code,
			"code_name": lb.LoadBalancerAlgorithmType.CodeName,
		})
		d.Set("load_balancer_description", lb.LoadBalancerDescription)
		d.Set("create_date", lb.CreateDate)
		d.Set("domain_name", lb.DomainName)
		d.Set("internet_line_type", map[string]interface{}{
			"code":      lb.InternetLineType.Code,
			"code_name": lb.InternetLineType.CodeName,
		})
		d.Set("load_balancer_instance_status_name", lb.LoadBalancerInstanceStatusName)
		d.Set("load_balancer_instance_status", map[string]interface{}{
			"code":      lb.LoadBalancerInstanceStatus.Code,
			"code_name": lb.LoadBalancerInstanceStatus.CodeName,
		})
		d.Set("load_balancer_instance_operation", map[string]interface{}{
			"code":      lb.LoadBalancerInstanceOperation.Code,
			"code_name": lb.LoadBalancerInstanceOperation.CodeName,
		})
		d.Set("network_usage_type", map[string]interface{}{
			"code":      lb.NetworkUsageType.Code,
			"code_name": lb.NetworkUsageType.CodeName,
		})
		d.Set("is_http_keep_alive", lb.IsHTTPKeepAlive)
		d.Set("connection_timeout", lb.ConnectionTimeout)
		d.Set("certificate_name", lb.CertificateName)

		if len(lb.LoadBalancerRuleList) != 0 {
			d.Set("load_balancer_rule_list", getLoadBalancerRuleList(lb.LoadBalancerRuleList))
		}
		if len(lb.LoadBalancedServerInstanceList) != 0 {
			d.Set("load_balanced_server_instance_list", getLoadBalancedServerInstanceList(lb.LoadBalancedServerInstanceList))
		} else {
			d.Set("load_balanced_server_instance_list", nil)
		}
	}

	return nil
}

func getLoadBalancerRuleList(lbRuleList []sdk.LoadBalancerRule) []interface{} {
	list := make([]interface{}, 0, len(lbRuleList))

	for _, r := range lbRuleList {
		rule := map[string]interface{}{
			"protocol_type_code": map[string]interface{}{
				"code":      r.ProtocolType.Code,
				"code_name": r.ProtocolType.CodeName,
			},
			"load_balancer_port":    r.LoadBalancerPort,
			"server_port":           r.ServerPort,
			"l7_health_check_path":  r.L7HealthCheckPath,
			"certificate_name":      r.CertificateName,
			"proxy_protocol_use_yn": r.ProxyProtocolUseYn,
		}
		log.Printf("%#v", rule)
		list = append(list, rule)
		for key, value := range rule {
			log.Printf("%#v %#v", key, value)
		}
	}
	return list
}

func getLoadBalancedServerInstanceList(loadBalancedServerInstanceList []sdk.LoadBalancedServerInstance) []string {
	list := make([]string, 0, len(loadBalancedServerInstanceList))

	for _, instance := range loadBalancedServerInstanceList {
		list = append(list, instance.ServerInstanceList[0].ServerInstanceNo)
	}

	return list
}

func resourceNcloudLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NcloudSdk).conn
	return deleteLoadBalancerInstance(conn, d.Id())
}

func resourceNcloudLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceNcloudLoadBalancerRead(d, meta)
}

func buildCreateLoadBalancerInstanceParams(d *schema.ResourceData) *sdk.RequestCreateLoadBalancerInstance {
	lbRuleList := make([]sdk.RequestLoadBalancerRule, 0, len(d.Get("load_balancer_rule_list").([]interface{})))

	for _, v := range d.Get("load_balancer_rule_list").([]interface{}) {
		lbRule := new(sdk.RequestLoadBalancerRule)
		for key, value := range v.(map[string]interface{}) {
			switch key {
			case "protocol_type_code":
				lbRule.ProtocolTypeCode = value.(string)
			case "load_balancer_port":
				lbRule.LoadBalancerPort = value.(int)
			case "server_port":
				lbRule.ServerPort = value.(int)
			case "l7_health_check_path":
				lbRule.L7HealthCheckPath = value.(string)
			case "certificate_name":
				lbRule.CertificateName = value.(string)
			case "proxy_protocol_use_yn":
				lbRule.ProxyProtocolUseYn = value.(string)
			}
		}
		lbRuleList = append(lbRuleList, *lbRule)
	}

	reqParams := &sdk.RequestCreateLoadBalancerInstance{
		LoadBalancerName:              d.Get("load_balancer_name").(string),
		LoadBalancerAlgorithmTypeCode: d.Get("load_balancer_algorithm_type_code").(string),
		LoadBalancerDescription:       d.Get("load_balancer_description").(string),
		LoadBalancerRuleList:          lbRuleList,
		ServerInstanceNoList:          StringList(d.Get("server_instance_no_list").([]interface{})),
		InternetLineTypeCode:          d.Get("internet_line_type_code").(string),
		NetworkUsageTypeCode:          d.Get("network_usage_type_code").(string),
		RegionNo:                      d.Get("region_no").(string),
	}
	return reqParams
}

func getLoadBalancerInstance(conn *sdk.Conn, LoadBalancerInstanceNo string) (*sdk.LoadBalancerInstance, error) {
	reqParams := &sdk.RequestLoadBalancerInstanceList{
		LoadBalancerInstanceNoList: []string{LoadBalancerInstanceNo},
	}
	resp, err := conn.GetLoadBalancerInstanceList(reqParams)
	if err != nil {
		logErrorResponse("GetLoadBalancerInstanceList", err, reqParams)
		return nil, err
	}
	logCommonResponse("GetLoadBalancerInstanceList", reqParams, resp.CommonResponse)

	for _, inst := range resp.LoadBalancerInstanceList {
		if LoadBalancerInstanceNo == inst.LoadBalancerInstanceNo {
			return &inst, nil
		}
	}
	return nil, nil
}

func deleteLoadBalancerInstance(conn *sdk.Conn, LoadBalancerInstanceNo string) error {
	reqParams := &sdk.RequestDeleteLoadBalancerInstances{
		LoadBalancerInstanceNoList: []string{LoadBalancerInstanceNo},
	}
	resp, err := conn.DeleteLoadBalancerInstances(reqParams)
	if err != nil {
		logErrorResponse("DeleteLoadBalancerInstance", err, LoadBalancerInstanceNo)
		return err
	}
	var commonResponse = common.CommonResponse{}
	if resp != nil {
		commonResponse = resp.CommonResponse
	}
	logCommonResponse("DeleteLoadBalancerInstance", LoadBalancerInstanceNo, commonResponse)

	return waitForDeleteLoadBalancerInstance(conn, LoadBalancerInstanceNo)
}

func waitForLoadBalancerInstance(conn *sdk.Conn, id string, status string, timeout time.Duration) error {
	c1 := make(chan error, 1)

	go func() {
		for {
			instance, err := getLoadBalancerInstance(conn, id)

			if err != nil {
				c1 <- err
				return
			}

			if instance == nil || instance.LoadBalancerInstanceStatus.Code == status {
				c1 <- nil
				return
			}

			log.Printf("[DEBUG] Wait get load balancer instance [%s] status [%s] to be [%s]", id, instance.LoadBalancerInstanceStatus.Code, status)
			time.Sleep(time.Second * 1)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(time.Second * timeout):
		return fmt.Errorf("TIMEOUT : delete load balancer instance [%s] ", id)
	}
}

func waitForDeleteLoadBalancerInstance(conn *sdk.Conn, id string) error {
	c1 := make(chan error, 1)

	go func() {
		for {
			instance, err := getLoadBalancerInstance(conn, id)

			if err != nil {
				c1 <- err
				return
			}

			if instance == nil {
				c1 <- nil
				return
			}

			log.Printf("[DEBUG] Wait delete load balancer instance [%s] ", id)
			time.Sleep(time.Second * 1)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(time.Second * DefaultTimeout):
		return fmt.Errorf("TIMEOUT : delete load balancer instance [%s] ", id)
	}
}

var loadBalancerRuleSchemaResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"protocol_type_code": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Protocol type code of load balancer rules. The following codes are available. [HTTP | HTTPS | TCP | SSL]",
		},
		"protocol_type": {
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        commonCodeSchemaResource,
			Description: "",
		},
		"load_balancer_port": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Load balancer port of load balancer rules",
		},
		"server_port": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Server port of load balancer rules",
		},
		"l7_health_check_path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Health check path of load balancer rules. Required when the loadBalancerRuleList.N.protocolTypeCode is HTTP/HTTPS.",
		},
		"certificate_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Load balancer SSL certificate. Required when the loadBalancerRuleList.N.protocloTypeCode value is SSL/HTTPS.",
		},
		"proxy_protocol_use_yn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Use 'Y' if you want to check client IP addresses by enabling the proxy protocol while you select TCP or SSL.",
		},
	},
}
