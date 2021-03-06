package elbv2

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/coreos/alb-ingress-controller/pkg/util/log"
)

type mockedELBV2DescribeLoadBalancers struct {
	elbv2iface.ELBV2API
	Resp elbv2.DescribeLoadBalancersOutput
}

func (m mockedELBV2DescribeLoadBalancers) DescribeLoadBalancersPagesWithContext(ctx aws.Context, input *elbv2.DescribeLoadBalancersInput, fn func(*elbv2.DescribeLoadBalancersOutput, bool) bool, opts ...request.Option) error {
	fn(&m.Resp, false)
	return nil
}

func TestClusterLoadBalancers(t *testing.T) {
	loadBalancers := []*elbv2.LoadBalancer{
		{LoadBalancerName: aws.String("prod-abc123456789")},
		{LoadBalancerName: aws.String("dev-abc123456789")},
		{LoadBalancerName: aws.String("prod-123456789abc")},
		{LoadBalancerName: aws.String("qa-abc123456789")},
	}

	cases := []struct {
		Resp        elbv2.DescribeLoadBalancersOutput
		ClusterName string
		Expected    []*elbv2.LoadBalancer
	}{
		{
			Resp:        elbv2.DescribeLoadBalancersOutput{LoadBalancers: loadBalancers},
			ClusterName: "prod",
			Expected: []*elbv2.LoadBalancer{
				{LoadBalancerName: aws.String("prod-abc123456789")},
				{LoadBalancerName: aws.String("prod-123456789abc")},
			},
		},
		{
			Resp:        elbv2.DescribeLoadBalancersOutput{LoadBalancers: loadBalancers},
			ClusterName: "miss",
			Expected:    []*elbv2.LoadBalancer{},
		},
		{
			Resp:        elbv2.DescribeLoadBalancersOutput{LoadBalancers: loadBalancers},
			ClusterName: "",
			Expected:    []*elbv2.LoadBalancer{},
		},
	}

	for _, c := range cases {
		e := ELBV2{mockedELBV2DescribeLoadBalancers{Resp: c.Resp}}
		loadbalancers, err := e.ClusterLoadBalancers(&c.ClusterName)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}
		if a, e := len(loadbalancers), len(c.Expected); a != e {
			t.Fatalf("%v, expected %d load balancers, got %d", c.ClusterName, e, a)
		}
		for j, loadbalancer := range loadbalancers {
			if a, e := loadbalancer, c.Expected[j]; *a.LoadBalancerName != *e.LoadBalancerName {
				t.Errorf("%v, expected %v loadbalancer, got %v", c.ClusterName, e, a)
			}
		}
	}
}

func TestSortLoadBalancerAttributes(t *testing.T) {
	key1 := "hello"
	value1 := "world"
	key2 := "other"
	value2 := "value"
	key3 := "something"
	value3 := "else"
	attributes1 := Attributes{
		Items: []*elbv2.LoadBalancerAttribute{
			&elbv2.LoadBalancerAttribute{
				Key:   &key2,
				Value: &value2,
			},
			&elbv2.LoadBalancerAttribute{
				Key:   &key1,
				Value: &value1,
			},
		},
	}
	attributes2 := Attributes{
		Items: []*elbv2.LoadBalancerAttribute{
			&elbv2.LoadBalancerAttribute{
				Key:   &key1,
				Value: &value1,
			},
			&elbv2.LoadBalancerAttribute{
				Key:   &key2,
				Value: &value2,
			},
		},
	}
	sort.Sort(attributes1)
	sort.Sort(attributes2)
	if log.Prettify(attributes1) != log.Prettify(attributes2) {
		t.Errorf("LoadBalancerAttribute sort failed, expected attributes to be inequal.")
	}
	attributes2 = Attributes{
		Items: []*elbv2.LoadBalancerAttribute{
			&elbv2.LoadBalancerAttribute{
				Key:   &key1,
				Value: &value1,
			},
			&elbv2.LoadBalancerAttribute{
				Key:   &key2,
				Value: &value2,
			},
			&elbv2.LoadBalancerAttribute{
				Key:   &key3,
				Value: &value3,
			},
		},
	}
	sort.Sort(attributes1)
	sort.Sort(attributes2)
	if log.Prettify(attributes1) == log.Prettify(attributes2) {
		t.Errorf("LoadBalancerAttribute sort failed, expected attributes to be equal.")
	}
}
