package prometheus

import (
	"strings"

	"github.com/pkg/errors"
)

var MetricsMap = map[string]string{
	"sum_tkapi_request_7d":               "ceil(sum(increase(tkapi_request_total{$}[7d])))",
	"sum_tkapi_request_24h":              "ceil(sum(increase(tkapi_request_total{$}[24h])))",
	"avg_tkapi_request_latency_7d":       "sum(increase(tkapi_request_duration_seconds_sum{$}[7d])) /sum(increase(tkapi_request_duration_seconds_count{$}[7d]))",
	"sum_tkapi_request_1h":               "ceil(sum(increase(tkapi_request_total{$}[1h])))",
	"p999_tkapi_request_latency":         "histogram_quantile(0.999, sum by (le) (rate(tkapi_request_duration_seconds_bucket{$}[1d])))",
	"p95_tkapi_request_latency":          "histogram_quantile(0.95, sum by (le) (rate(tkapi_request_duration_seconds_bucket{$}[1d])))",
	"p99_tkapi_request_latency":          "histogram_quantile(0.99, sum by (le) (rate(tkapi_request_duration_seconds_bucket{$}[1d])))",
	"upstream_msg":                       "sum(iothub_msg_total{direction='upstream',$})",                                                                                       // 上行消息数量
	"downstream_msg":                     "sum(iothub_msg_total{direction='downstream',$})",                                                                                     // 下行消息数量
	"subscribe_num":                      "(sum(subscribe_num{$})) / (count (sum by (pod) (subscribe_num)))",                                                                    // 订阅数
	"subscribe_entities_num":             "sum(subscribe_entities_num{$}) / (count (sum by (pod) (subscribe_entities_num)))",                                                    // 订阅的实体数量
	"rule_num":                           "sum(rule_num{$}) / (count (sum by (pod) (rule_num)))",                                                                                // 路由数量
	"rule_execute_num_24h":               "sum(increase(rule_execute_num{$}[24h])) / (count (sum by (pod) (rule_num)))",                                                         // 24小时 规则执行数
	"rate_rule_failure_24h":              "ceil((sum(rule_execute_num{status='failure',$}) / sum(rule_execute_num{$}))*100)",                                                    // 规则执行失败率
	"sum_device_num":                     "sum(device_num_total{$}) / (count (sum by (pod) (device_num_total)))",                                                                // 设备数量
	"sum_template_num":                   "sum(device_template_total{$}) / (count (sum by (pod) (device_template_total)))",                                                      // 模板数量
	"rate_online":                        "ceil((sum(iothub_connected_total{$}) / sum(device_num_total{$})) * 100)",                                                             // 在线率
	"msg_storage_days":                   "((sum(msg_storage_seconds{tenant_id='admin'})) / (count (sum by (pod) (msg_storage_seconds)))) / 86400",                              // 消息存储天数
	"msg_storage_space_unused_bytes":     "(sum(msg_storage_space{space_type='total'})-sum( msg_storage_space{space_type='used'} ))/(count (sum by (pod) (msg_storage_space)))", // 消息存储可用空间
	"msg_storage_space_usage_percentage": "ceil((sum(msg_storage_space{space_type='used'}) / sum(msg_storage_space{space_type='total'}))*100)",                                  // 消息存储空间使用占比
	"core_msg_days":                      "ceil(sum(increase(core_msg_total{$}[1d])))",                                                                                          // 日消息量
}

func ExpressFromMetricsMap(name, label string) (string, error) {
	if v, ok := MetricsMap[name]; ok {
		expr := strings.Replace(v, "$", label, -1)
		return expr, nil
	}
	return "", errors.New("metrics name not existed")
}