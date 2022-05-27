package service

import (
	"context"
	"time"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	monv1alp1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1alpha1"
	"github.com/prometheus/client_golang/api"
	promv1cli "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-monitor/api/prometheus/v1"
	mprom "github.com/tkeel-io/tkeel-monitor/pkg/model/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type PrometheusService struct {
	PromNS       string
	TKeelNS      string
	monV1Cli     *monv1.MonitoringV1Client
	monV1Alp1Cli *monv1alp1.MonitoringV1alpha1Client

	pAPI promv1cli.API
}

func NewPrometheusService(promNamespace, tKeelNamespace string) *PrometheusService {
	conf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("in cluster config: %s", err)
	}
	mc, err := monv1.NewForConfig(conf)
	if err != nil {
		log.Fatalf("new monitor client: %s", err)
	}
	mca1, err := monv1alp1.NewForConfig(conf)
	if err != nil {
		log.Fatalf("new monitor client: %s", err)
	}
	l, err := mc.Prometheuses(promNamespace).List(context.TODO(), metav1.ListOptions{Limit: 1})
	if err != nil {
		log.Errorf("list prometheus: %s", err)
		return nil
	}
	if len(l.Items) == 0 {
		log.Fatalf("promNamespace(%s) not has prometheus", promNamespace)
	}
	url := ""
	p := l.Items[0]
	if p.Spec.ExternalURL != "" {
		url = p.Spec.ExternalURL
	} else {
		url = "http://" + p.Spec.ServiceAccountName + "." + p.ObjectMeta.Namespace + ":9090"
	}
	pc, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		log.Fatalf("new prometheus client err: %s", err)
	}
	pAPI := promv1cli.NewAPI(pc)
	res, err := pAPI.Buildinfo(context.TODO())
	if err != nil {
		log.Errorf("new prometheus api build.info err: %s", err)
	}
	log.Debugf("prometheus build.info:\n %v", res)
	return &PrometheusService{
		PromNS:       promNamespace,
		TKeelNS:      tKeelNamespace,
		pAPI:         pAPI,
		monV1Cli:     mc,
		monV1Alp1Cli: mca1,
	}
}

func (s *PrometheusService) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	var (
		value       model.Value
		warn        promv1cli.Warnings
		err         error
		metricsData *pb.MetricsData
	)
	st := time.UnixMilli(req.GetSt())
	et := time.UnixMilli(req.GetEt())
	if req.GetStep() == "" {
		value, warn, err = s.pAPI.Query(ctx, req.GetQuery(), et)
		if warn != nil {
			log.Warnf("query %s warn: %v", req, warn)
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}
		metricsData = mprom.Parse2pbQueryResp(value, nil)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		step, err1 := time.ParseDuration(req.GetStep())
		if err1 != nil {
			log.Errorf("time step parse err: %s", err1)
			return nil, pb.ResourceErrUnknown()
		}
		value, warn, err = s.pAPI.QueryRange(ctx, req.GetQuery(), promv1cli.Range{
			Start: st,
			End:   et,
			Step:  step,
		})
		if warn != nil {
			log.Warnf("query range %s warn: %v", req, warn)
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}
		metricsData = mprom.Parse2pbQueryRangeResp(value, nil)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return &pb.QueryResponse{
		Result: metricsData,
	}, nil
}

func (s *PrometheusService) Rules(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	s.monV1Cli.PrometheusRules(s.TKeelNS)
	s.monV1Cli.Alertmanagers(s.TKeelNS)
	s.monV1Alp1Cli.AlertmanagerConfigs(s.TKeelNS)

	return nil, nil
}
